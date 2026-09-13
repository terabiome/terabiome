package models

type Config struct {
	DebugPrintOnly bool             `yaml:"debug_print_only,omitempty"`
	OS             OSConfig         `yaml:"os,omitempty"`
	Kubernetes     KubernetesConfig `yaml:"kubernetes,omitempty"`
}
