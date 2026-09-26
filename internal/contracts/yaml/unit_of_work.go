package yamlcontracts

import "github.com/go-playground/validator/v10"

type UnitOfWorkType string

const (
	UnitOfWorkTypeWorkflow UnitOfWorkType = "workflow"
)

// UnitOfWork is root of YAML manifest for dispatching works.
type UnitOfWork struct {
	Name      string               `yaml:"name" validate:"required,min=3,max=100"`
	Targets   []string             `yaml:"targets" validate:"required,min=1,dive,required"` // ['*'] means for all available ones
	Type      UnitOfWorkType       `yaml:"type" validate:"required,oneof=workflow"`
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

func (u UnitOfWork) Validate(validator *validator.Validate) error {
	return validator.Struct(u)
}
