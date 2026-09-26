package contracts

type UnitOfWorkType string

const (
	UnitOfWorkTypeWorkflow UnitOfWorkType = "workflow"
)

// UnitOfWork is root of YAML manifest for dispatching works.
type UnitOfWork struct {
	Name      string               `yaml:"name"`
	Targets   []string             `yaml:"targets"` // ['*'] means for all available ones
	Type      UnitOfWorkType       `yaml:"type"`
	Sequences []UnitOfWorkSequence `yaml:"sequences"`
}

// UnitOfWorkSequence is part of UnitOfWork, representing logical grouping of linear execution steps.
type UnitOfWorkSequence struct {
	Name      string           `yaml:"name"`
	DependsOn []string         `yaml:"depends_on"` // must not contain names of another dependent sequence, else fatal
	Steps     []UnitOfWorkStep `yaml:"steps"`
}

func (s UnitOfWorkSequence) IsIndependent() bool {
	return len(s.DependsOn) == 0
}

// UnitOfWorkStep is part of UnitOfWorkSequence, representing the smallest unit of execution.
type UnitOfWorkStep struct {
	Name       string         `yaml:"name"`
	IsAsync    bool           `yaml:"is_async"` // false by default, unsupported as of now
	Parameters map[string]any `yaml:"parameters"`
}
