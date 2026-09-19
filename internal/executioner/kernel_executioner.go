package executioner

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/terabiome/infrastructures/internal/executor"
	"github.com/terabiome/infrastructures/internal/models"
)

type kernelExecutioner struct {
	cfg      models.KernelConfig
	executor Executor
	logger   Logger
}

func (e *kernelExecutioner) Validate() error {
	return nil
}

func (e *kernelExecutioner) Process(ctx context.Context) error {
	return errors.Join(
		e.processKernelModules(ctx),
		e.processSystemKernelParameters(ctx),
	)
}

func (e *kernelExecutioner) processKernelModules(ctx context.Context) error {
	if len(e.cfg.KernelModules.Values) == 0 {
		return nil
	}

	type invertFileEntry struct {
		value  string
		action models.KernelModuleAction
	}

	var (
		invertedFileEntryLookupMap = map[string][]invertFileEntry{}
		invertedFileModuleMap      = map[string][]string{}
		invertedModuleLookupMap    = map[string]struct{}{}
		faultyEntries              = map[string]struct{}{}
	)

	var (
		kernelModuleDir    string      = "/etc/modules-load.d"
		filePermissionBits os.FileMode = 0o644
	)

	for _, value := range e.cfg.KernelModules.Values {
		value.Value = strings.TrimSpace(value.Value)
		faulty := false

		for _, lookupKey := range []string{
			fmt.Sprintf("module:%s", value.Value),
			fmt.Sprintf("file:%s-module:%s", value.FilePath, value.Value),
		} {
			if _, ok := invertedModuleLookupMap[lookupKey]; ok {
				faultyEntries[lookupKey] = struct{}{}
				faulty = true
			} else {
				invertedModuleLookupMap[lookupKey] = struct{}{}
			}
		}
		if faulty {
			continue
		}

		if _, ok := invertedFileEntryLookupMap[value.FilePath]; !ok {
			invertedFileEntryLookupMap[value.FilePath] = []invertFileEntry{}
		}
		invertedFileEntryLookupMap[value.FilePath] = append(invertedFileEntryLookupMap[value.FilePath],
			invertFileEntry{value.Value, value.Action},
		)
	}

	if len(faultyEntries) > 0 {
		return fmt.Errorf(
			"encountered %d faulty kernel module entries in config: %v",
			len(faultyEntries), slices.Collect(maps.Keys(faultyEntries)),
		)
	}

	filePathPattern, _ := regexp.Compile(fmt.Sprintf(`%s/[a-zA-Z0-9_\-]+\.conf$`, kernelModuleDir))
	for filePath, invertEntries := range invertedFileEntryLookupMap {
		if filePathPattern.FindString(filePath) == "" {
			return fmt.Errorf("kernel module file path does not match expected pattern %s: %s",
				filePathPattern.String(), filePath)
		}
		isFileAlreadyExist, inspectFileErr := func(filePath string) (bool, error) {
			_, err := os.Stat(filePath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return false, nil
				}
				// something terrible, could not inspect the file -> fail
				return false, fmt.Errorf("failed to inspect file stat: %w", err)
			}
			return true, nil
		}(filePath)

		// stop processing completely
		if inspectFileErr != nil {
			return fmt.Errorf("failed to check kernel module file stat %s: %w", filePath, inspectFileErr)
		}

		moduleMap := map[string]struct{}{}
		if isFileAlreadyExist {
			byteContent, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read kernel module file %s: %w",
					filePath, err,
				)
			}
			for module := range strings.SplitSeq(string(byteContent), "\n") {
				moduleMap[strings.TrimSpace(module)] = struct{}{}
			}
		} else {
			if _, err := os.OpenFile(filePath, os.O_CREATE, filePermissionBits); err != nil {
				return fmt.Errorf("failed to create kernel module file %s: %w",
					filePath, err,
				)
			}
		}

		for _, invertEntry := range invertEntries {
			switch invertEntry.action {
			case models.KernelModuleActionAdd:
				// allow overwriting since value is bare string, not KV
				moduleMap[invertEntry.value] = struct{}{}
			case models.KernelModuleActionRemove:
				// if exist in map -> remove
				delete(moduleMap, invertEntry.value)
			default:
				e.logger.WarnContext(ctx, "invalid action for kernel module",
					"module", invertEntry.value,
					"file_path", filePath,
					"action", invertEntry.action,
				)
			}
		}
		// don't process here, aggregate modules to stage in final writes
		// -> avoid partial update when processing individual files
		invertedFileModuleMap[filePath] = slices.Collect(maps.Keys(moduleMap))
	}

	for filePath, modules := range invertedFileModuleMap {
		if err := os.WriteFile(filePath, []byte(strings.Join(modules, "\n")), filePermissionBits); err != nil {
			return fmt.Errorf("failed to write kernel module entries to file %s: %w", filePath, err)
		}

		if e.cfg.KernelModules.Immediate {
			// activate
			input := executor.Input{
				Mode:   executor.ModeSync,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
				Command: models.Command{
					Executable: "modprobe",
					Arguments:  []string{"-a"},
				},
			}

			for _, module := range modules {
				input.Command.Arguments = append(input.Command.Arguments, module)
			}

			output := e.executor.Execute(ctx, &input)
			if output.Done() && output.Error != nil {
				return fmt.Errorf("failed to activate modules for file %s: %w", filePath, output.Error)
			}
		}
	}

	return nil
}

func (e *kernelExecutioner) processSystemKernelParameters(ctx context.Context) error {
	if len(e.cfg.SystemKernelParameters.Values) == 0 {
		return nil
	}

	type invertSysctlEntry struct {
		key    string
		value  string
		action models.SystemKernelParameterAction
	}

	var (
		invertedFileEntryLookupMap = map[string][]invertSysctlEntry{}
		invertedFileParamMap       = map[string]map[string]string{}
		faultyEntries              = map[string]struct{}{}
		seenKeys                   = map[string]struct{}{}
	)

	const (
		sysctlDir                      = "/etc/sysctl.d"
		filePermissionBits os.FileMode = 0o644
	)

	// Validate and deduplicate entries
	for _, value := range e.cfg.SystemKernelParameters.Values {
		value.Key = strings.TrimSpace(value.Key)
		value.Value = strings.TrimSpace(value.Value)

		// Basic validation of sysctl key format (e.g., net.ipv4.ip_forward)
		validKeyPattern := regexp.MustCompile(`^[a-z][a-z0-9_.-]+$`)
		if !validKeyPattern.MatchString(value.Key) {
			faultyEntries[value.Key] = struct{}{}
			continue
		}

		lookupKey := fmt.Sprintf("file:%s-key:%s", value.FilePath, value.Key)
		if _, ok := seenKeys[lookupKey]; ok {
			faultyEntries[lookupKey] = struct{}{}
			continue
		}
		seenKeys[lookupKey] = struct{}{}

		if _, ok := invertedFileEntryLookupMap[value.FilePath]; !ok {
			invertedFileEntryLookupMap[value.FilePath] = []invertSysctlEntry{}
		}
		invertedFileEntryLookupMap[value.FilePath] = append(
			invertedFileEntryLookupMap[value.FilePath],
			invertSysctlEntry{value.Key, value.Value, value.Action},
		)
	}

	if len(faultyEntries) > 0 {
		return fmt.Errorf(
			"encountered %d faulty sysctl entries in config: %v",
			len(faultyEntries), slices.Collect(maps.Keys(faultyEntries)),
		)
	}

	// Validate file paths match expected pattern
	filePathPattern := regexp.MustCompile(fmt.Sprintf(`%s/[a-zA-Z0-9_-]+\.conf$`, sysctlDir))
	for filePath := range invertedFileEntryLookupMap {
		if filePathPattern.FindString(filePath) == "" {
			return fmt.Errorf("sysctl file path does not match expected pattern %s: %s",
				filePathPattern.String(), filePath)
		}
	}

	// Stage all changes in memory before writing
	for filePath, invertEntries := range invertedFileEntryLookupMap {
		paramMap := map[string]string{}

		// Read existing file if it exists
		byteContent, err := os.ReadFile(filePath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to read sysctl file %s: %w", filePath, err)
		}
		if err == nil {
			for line := range strings.SplitSeq(string(byteContent), "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "#") {
					continue
				}
				parts := strings.SplitN(trimmed, "=", 2)
				if len(parts) == 2 {
					paramMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}

		// Apply staged changes
		for _, entry := range invertEntries {
			switch entry.action {
			case models.SystemKernelParameterActionAdd:
				paramMap[entry.key] = entry.value
			case models.SystemKernelParameterActionRemove:
				delete(paramMap, entry.key)
			default:
				e.logger.WarnContext(ctx, "invalid action for sysctl param",
					"key", entry.key,
					"file_path", filePath,
					"action", entry.action,
				)
			}
		}

		invertedFileParamMap[filePath] = paramMap
	}

	// Atomic write all files
	for filePath, params := range invertedFileParamMap {
		var lines []string
		// Sort keys for deterministic output
		sortedKeys := slices.Sorted(maps.Keys(params))
		for _, key := range sortedKeys {
			lines = append(lines, fmt.Sprintf("%s = %s", key, params[key]))
		}

		content := strings.Join(lines, "\n") + "\n"
		if err := os.WriteFile(filePath, []byte(content), filePermissionBits); err != nil {
			return fmt.Errorf("failed to write sysctl params to %s: %w", filePath, err)
		}
	}

	// Apply immediately if configured
	if e.cfg.SystemKernelParameters.Immediate {
		output := e.executor.Execute(ctx, &executor.Input{
			Mode:   executor.ModeSync,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Command: models.Command{
				Executable: "sysctl",
				Arguments:  []string{"--system"},
			},
		})
		if output.Done() && output.Error != nil {
			return fmt.Errorf("sysctl --system: %w", output.Error)
		}
	}

	return nil
}
