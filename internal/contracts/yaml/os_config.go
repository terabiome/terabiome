package yamlcontracts

import "errors"

type FirewalldTargetAction string

const (
	FirewalldTargetActionDrop    FirewalldTargetAction = "DROP"
	FirewalldTargetActionReject  FirewalldTargetAction = "REJECT"
	FirewalldTargetActionAccept  FirewalldTargetAction = "ACCEPT"
	FirewalldTargetActionUnknown FirewalldTargetAction = ""
)

type FirewalldZoneEntityAction string

const (
	FirewalldZoneEntityActionAdd     FirewalldZoneEntityAction = "add"
	FirewalldZoneEntityActionRemove  FirewalldZoneEntityAction = "remove"
	FirewalldZoneEntityActionUnknown FirewalldZoneEntityAction = ""
)

type KernelModuleAction string

const (
	KernelModuleActionAdd    KernelModuleAction = "add"
	KernelModuleActionRemove KernelModuleAction = "remove"
)

type SystemKernelParameterAction string

const (
	SystemKernelParameterActionAdd    SystemKernelParameterAction = "add"
	SystemKernelParameterActionRemove SystemKernelParameterAction = "remove"
)

// done
type OSConfig struct {
	Shell   ShellConfig   `yaml:"shell"`
	Network NetworkConfig `yaml:"network"`
	Disk    DiskConfig    `yaml:"swap"`
	Kernel  KernelConfig  `yaml:"kernel"`
}

// done - both enable + disable appear -> conflict, throw error
type ShellConfig struct {
	EnableService  bool `yaml:"enable_service,omitempty"`
	DisableService bool `yaml:"disable_service,omitempty"`
	StartService   bool `yaml:"start_service"`
}

func (cfg ShellConfig) SoftValidate() error {
	if cfg.EnableService == true && cfg.DisableService == true {
		return errors.New("shell config: both disable and enable service are present")
	}

	return nil
}

// done
type NetworkConfig struct {
	Tailscale TailscaleConfig `yaml:"tailscale"`
	Firewall  FirewallConfig  `yaml:"firewall"`
}

// done - both enable + disable appear -> conflict, throw error
type TailscaleConfig struct {
	InstallPackage bool   `yaml:"install_package"`
	EnableService  bool   `yaml:"enable_service,omitempty"`
	DisableService bool   `yaml:"disable_service,omitempty"`
	StartService   bool   `yaml:"start_service"`
	AuthKey        string `yaml:"auth_key"`
}

func (cfg TailscaleConfig) SoftValidate() error {
	if cfg.EnableService == true && cfg.DisableService == true {
		return errors.New("tailscale config: both disable and enable service are present")
	}

	return nil
}

// done - both enable + disable appear -> conflict, throw error
type FirewallConfig struct {
	EnableService  bool                  `yaml:"enable_service,omitempty"`
	DisableService bool                  `yaml:"disable_service,omitempty"`
	Immediate      bool                  `yaml:"immediate"`
	Zones          map[string]ZoneConfig `yaml:"zones"`
}

func (cfg FirewallConfig) SoftValidate() error {
	if cfg.EnableService == true && cfg.DisableService == true {
		return errors.New("firewall config: both disable and enable service are present")
	}

	return nil
}

// done
type ZoneConfig struct {
	Interface InterfaceConfig       `yaml:"interface"`
	Services  []ServiceConfig       `yaml:"services"`
	SetTarget FirewalldTargetAction `yaml:"set_service,omitempty"`
}

// done
type InterfaceConfig struct {
	Name   string                    `yaml:"name"`
	Action FirewalldZoneEntityAction `yaml:"action,omitempty"`
}

// done
type ServiceConfig struct {
	Name   string                    `yaml:"name"`
	Action FirewalldZoneEntityAction `yaml:"action,omitempty"`
}

// done
type DiskConfig struct {
	Swap SwapConfig `yaml:"swap"`
}

// done
type SwapConfig struct {
	DisableSwap             bool `yaml:"disable"`
	EditFstab               bool `yaml:"edit_fstab"`
	MaskSwapRelatedServices bool `yaml:"mask_swap_related_services"`
}

// done
type KernelConfig struct {
	KernelModules          KernelModulesConfig          `yaml:"kernel_modules"`
	SystemKernelParameters SystemKernelParametersConfig `yaml:"system_kernel_parameters"`
}

/*
Can refactor Kernel config to this and make things easier. Low priority though.
  kernel:
    kernel_modules:
      immediate: true
      mappings:
        - path: /etc/modules-load.d/k8s.conf
          values:
            - value: br_netfilter
              action: add
            - value: overlay
              action: add
*/

// done
type KernelModulesConfig struct {
	Immediate bool                 `yaml:"immediate"`
	Values    []KernelModuleConfig `yaml:"values"`
}

// done
type KernelModuleConfig struct {
	FilePath string             `yaml:"path"` // e.g. /etc/modules-load.d/*.conf
	Action   KernelModuleAction `yaml:"action"`
	Value    string             `yaml:"value"`
}

// done
type SystemKernelParametersConfig struct {
	Immediate bool                          `yaml:"immediate"`
	Values    []SystemKernelParameterConfig `yaml:"values"`
}

// done
type SystemKernelParameterConfig struct {
	FilePath string                      `yaml:"path"` // e.g. /etc/sysctl.d/*.conf
	Action   SystemKernelParameterAction `yaml:"action"`
	Key      string                      `yaml:"key"`
	Value    string                      `yaml:"value"`
}
