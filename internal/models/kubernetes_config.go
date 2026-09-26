package models

import "go.yaml.in/yaml/v4"

type KubernetesDistribution string

const (
	KubernetesDistributionK3s KubernetesDistribution = "k3s"
)

type KubernetesNodeRole string

const (
	KubernetesNodeRoleControlPlane KubernetesNodeRole = "control-plane"
	KubernetesNodeRoleWorker       KubernetesNodeRole = "worker"
)

type KubernetesConfig struct {
	Distribution   KubernetesDistribution `yaml:"distribution"`
	NodeRole       KubernetesNodeRole     `yaml:"node_role"`
	Version        string                 `yaml:"version"` // empty -> fallback to "latest"
	Configurations yaml.Node              `yaml:"configurations"`
}

type K3sConfigurations struct {
	ServerURL string `yaml:"server_url"`
	Token     string `yaml:"token"`
}

type RKE2Configurations struct {
}
