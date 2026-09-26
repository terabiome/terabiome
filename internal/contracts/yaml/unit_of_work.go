package yamlcontracts

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
)

type UnitOfWorkType string

const (
	UnitOfWorkTypeWorkflow UnitOfWorkType = "workflow"
)

// UnitOfWork is root of YAML manifest for dispatching works.
type UnitOfWork struct {
	Name      string               `yaml:"name" validate:"required,min=3,max=100"`
	Targets   []string             `yaml:"targets" validate:"required,min=1,dive,required"` // ['*'] means for all available ones
	Type      UnitOfWorkType       `yaml:"type" validate:"required"`
	Sequences []UnitOfWorkSequence `yaml:"sequences" validate:"required,min=1,dive"`
}

// UnitOfWorkSequence is part of UnitOfWork, representing logical grouping of linear execution steps.
type UnitOfWorkSequence struct {
	Name      string           `yaml:"name" validate:"required,alphanum,min=3,max=50"`
	DependsOn []string         `yaml:"depends_on" validate:"dive,required,alphanum"` // must not contain names of another dependent sequence, else fatal
	Steps     []UnitOfWorkStep `yaml:"steps" validate:"required,min=1,dive"`
}

func (s UnitOfWorkSequence) IsIndependent() bool {
	return len(s.DependsOn) == 0
}

// UnitOfWorkStep is part of UnitOfWorkSequence, representing the smallest unit of execution.
type UnitOfWorkStep struct {
	Name       string         `yaml:"name" validate:"required,alphanum,min=3,max=50"`
	IsAsync    bool           `yaml:"is_async"` // false by default, unsupported as of now
	Parameters map[string]any `yaml:"parameters"`
}

type UnitOfWorkSequenceDependencyGraph struct {
	// map{idx_of_dependant}map{idx_of_dependency}
	FirstDegreeGraph map[string]map[string]struct{}
	Violations       UnitOfWorkSequenceDependencyGraphViolations
}

type UnitOfWorkSequenceDependencyGraphViolations struct {
	NonFlatSuperstructure  map[string][]string
	FirstDegreeCircularity map[string][]string
}

func (u UnitOfWork) Validate(validator *validator.Validate) error {
	var err error

	if err = validator.Struct(u); err != nil {
		return err
	}

	switch u.Type {
	case UnitOfWorkTypeWorkflow:
	default:
		err = errors.Join(err, fmt.Errorf("unsupported unit-of-work type: %s", u.Type))
	}

	// sequence names must be unique
	charCountMap := map[string]int{}
	for _, seq := range u.Sequences {
		charCountMap[seq.Name] = charCountMap[seq.Name] + 1
	}
	var msgSegments []string
	for k, v := range charCountMap {
		if v > 1 {
			msgSegments = append(msgSegments, fmt.Sprintf("%s->%d", k, v))
		}
	}
	if len(msgSegments) > 0 {
		err = errors.Join(err, fmt.Errorf("duplicated name: %s", strings.Join(msgSegments, ",")))
	}

	// flat superstructure + no 1st-degree circularity (flat superstructure is superset rule of 1st-degree circularity)
	graph := u.GetSequenceDependencyGraph()

	if len(graph.Violations.NonFlatSuperstructure) > 0 {
		err = errors.Join(err, fmt.Errorf("non-flat superstructure violations: %v", graph.Violations.NonFlatSuperstructure))
	}

	if len(graph.Violations.FirstDegreeCircularity) > 0 {
		err = errors.Join(err, fmt.Errorf("circular dependencies detected: %v", graph.Violations.FirstDegreeCircularity))
	}

	return err
}

// GetSequenceDependencyGraph ...
func (u UnitOfWork) GetSequenceDependencyGraph() UnitOfWorkSequenceDependencyGraph {
	sortedSequences := u.GetSortedSequences()

	// N keys: include ones without dependant.
	// map{idx_of_dependant}map{idx_of_dependency}
	firstDegreeGraph := make(map[string]map[string]struct{}, len(sortedSequences))

	// atp sequences' name are verified to be unique, can be used as graph key to detect circularity
	for _, seq := range sortedSequences {
		firstDegreeGraph[seq.Name] = make(map[string]struct{}, len(seq.DependsOn))
		for _, name := range seq.DependsOn {
			firstDegreeGraph[seq.Name][name] = struct{}{}
		}
	}

	nonFlatSuperstructureViolations := map[string][]string{}
	firstDegreeCircularityViolations := map[string][]string{}

	for _, seq := range sortedSequences {
		// already past independent sequences
		if len(seq.DependsOn) > 0 {
			// A -> {}, B -> {}, C -> {}
			// D -> {A, B}, E -> {D} -> dead
			// D -> {E}, E -> {A} -> dead (how to detect this?)
			// D -> {E}, E -> {D} -> dead (how to detect this?)
			for _, dependencyName := range seq.DependsOn { // {E}
				// (1) if E has dependencies -> violate <flat superstructure rule>
				if len(firstDegreeGraph[dependencyName]) == 0 {
					continue
				}
				if _, ok := nonFlatSuperstructureViolations[seq.Name]; !ok {
					nonFlatSuperstructureViolations[seq.Name] = []string{}
				}
				nonFlatSuperstructureViolations[seq.Name] = append(nonFlatSuperstructureViolations[seq.Name], dependencyName)
				// (1.1) if it points back to D -> violate <non circularity rule> (this only applies for 1st-degree connection)
				if _, ok := firstDegreeCircularityViolations[seq.Name]; !ok {
					firstDegreeCircularityViolations[seq.Name] = []string{}
				}
				if _, ok := firstDegreeGraph[dependencyName][seq.Name]; ok {
					firstDegreeCircularityViolations[seq.Name] = append(firstDegreeCircularityViolations[seq.Name], dependencyName)
				}
			}
		}
	}

	return UnitOfWorkSequenceDependencyGraph{
		FirstDegreeGraph: firstDegreeGraph,
		Violations: UnitOfWorkSequenceDependencyGraphViolations{
			NonFlatSuperstructure:  nonFlatSuperstructureViolations,
			FirstDegreeCircularity: firstDegreeCircularityViolations,
		},
	}
}

// GetSortedSequences returns copy of Sequences sorted by number of dependencies.
// Ones without dependencies are put first, others retain their relative position before shuffle.
func (u UnitOfWork) GetSortedSequences() []UnitOfWorkSequence {
	fn := func(x, y UnitOfWorkSequence) int {
		// [a] < [a, b]
		// [] < [a]
		return len(x.DependsOn) - len(y.DependsOn)
	}
	return slices.SortedStableFunc(slices.Values(u.Sequences), fn)
}
