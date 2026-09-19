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
				// something terrible, could not inspect the file -> fail
				return false, fmt.Errorf("failed to inspect file stat: %w", err)
			}
			if errors.Is(err, os.ErrNotExist) {
				return false, nil
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
	// if len(e.cfg.SystemKernelParameters.Values) == 0 {
	// 	return nil
	// }

	// sysctlParams := map[string]string{
	// 	"net.bridge.bridge-nf-call-iptables":  "1",
	// 	"net.bridge.bridge-nf-call-ip6tables": "1",
	// 	"net.ipv4.ip_forward":                 "1",
	// }
	// var lines []string
	// for k, v := range sysctlParams {
	// 	lines = append(lines, fmt.Sprintf("%s = %s", k, v))
	// }
	// if err := os.WriteFile("/etc/sysctl.d/k8s.conf", []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
	// 	return fmt.Errorf("write /etc/sysctl.d/k8s.conf: %w", err)
	// }

	// if e.cfg.SystemKernelParameters.Immediate {
	// 	output := e.executor.Execute(ctx, &executor.Input{
	// 		Mode:   executor.ModeSync,
	// 		Stdout: os.Stdout,
	// 		Stderr: os.Stderr,
	// 		Command: models.Command{
	// 			Executable: "sysctl",
	// 			Arguments:  []string{"--system"},
	// 		},
	// 	})
	// 	if output.Done() && output.Error != nil {
	// 		return fmt.Errorf("sysctl --system: %w", output.Error)
	// 	}
	// }

	return nil
}
