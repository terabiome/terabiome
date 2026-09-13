package models

type firewalldTargetAction string

const (
	FirewalldTargetActionDrop    firewalldTargetAction = "DROP"
	FirewalldTargetActionReject  firewalldTargetAction = "REJECT"
	FirewalldTargetActionAccept  firewalldTargetAction = "ACCEPT"
	FirewalldTargetActionUnknown firewalldTargetAction = ""
)

type firewalldZoneEntityAction string

const (
	FirewalldZoneEntityActionAdd     firewalldZoneEntityAction = "add"
	FirewalldZoneEntityActionRemove  firewalldZoneEntityAction = "remove"
	FirewalldZoneEntityActionUnknown firewalldZoneEntityAction = ""
)

// done
type Config struct {
	DebugPrintOnly bool          `yaml:"debug_print_only"`
	Shell          ShellConfig   `yaml:"shell"`
	Network        NetworkConfig `yaml:"network"`
	Disk           DiskConfig    `yaml:"swap"`
	Kernel         KernelConfig  `yaml:"kernel"`
}

// done - both enable + disable appear -> conflict, throw error
type ShellConfig struct {
	EnableService  bool `yaml:"enable_service"`
	DisableService bool `yaml:"disable_service"`
	StartService   bool `yaml:"start_service"`
}

// done
type NetworkConfig struct {
	Tailscale TailscaleConfig `yaml:"tailscale"`
	Firewall  FirewallConfig  `yaml:"firewall"`
}

// done - both enable + disable appear -> conflict, throw error
type TailscaleConfig struct {
	InstallPackage bool   `yaml:"install_package"`
	EnableService  bool   `yaml:"enable_service"`
	DisableService bool   `yaml:"disable_service"`
	StartService   bool   `yaml:"start_service"`
	AuthKey        string `yaml:"auth_key"`
}

// done - both enable + disable appear -> conflict, throw error
type FirewallConfig struct {
	EnableService  bool                  `yaml:"enable_service"`
	DisableService bool                  `yaml:"disable_service"`
	Zones          map[string]ZoneConfig `yaml:"zones"`
}

// done
type ZoneConfig struct {
	Interface InterfaceConfig       `yaml:"interface"`
	Services  []ServiceConfig       `yaml:"services"`
	SetTarget firewalldTargetAction `yaml:"set_service"`
}

// done
type InterfaceConfig struct {
	Name   string                    `yaml:"name"`
	Action firewalldZoneEntityAction `yaml:"action"`
}

// done
type ServiceConfig struct {
	Name   string                    `yaml:"name"`
	Action firewalldZoneEntityAction `yaml:"action"`
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

// done
type KernelModulesConfig struct {
	Immediate bool                 `yaml:"immediate"`
	Values    []KernelModuleConfig `yaml:"values"`
}

// done
type KernelModuleConfig struct {
	FilePath string `yaml:"path"` // e.g. /etc/modules-load.d/*.conf
	Upsert   bool   `yaml:"upsert"`
	Value    string `yaml:"value"`
}

// done
type SystemKernelParametersConfig struct {
	Immediate bool                          `yaml:"immediate"`
	Values    []SystemKernelParameterConfig `yaml:"values"`
}

// done
type SystemKernelParameterConfig struct {
	FilePath string `yaml:"path"` // e.g. /etc/sysctl.d/*.conf
	Upsert   bool   `yaml:"upsert"`
	Key      string `yaml:"key"`
	Value    string `yaml:"value"`
}
