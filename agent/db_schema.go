package main

type ContainerStatus string

const (
	ContainerStatusImagePulling ContainerStatus = "image_pulling"
	ContainerStatusImagePullErr ContainerStatus = "image_pull_error"
	ContainerStatusPending      ContainerStatus = "pending"
	ContainerStatusRunning      ContainerStatus = "running"
	ContainerStatusStopped      ContainerStatus = "stopped"
	ContainerStatusExited       ContainerStatus = "exited"
)

type Container struct {
	Name              string          `gorm:"column:name;index"`
	ServiceName       string          `gorm:"column:service_name;index"`
	Image             string          `gorm:"column:image;index"`
	IPAddress         string          `gorm:"column:ip_address"`
	Status            ContainerStatus `gorm:"column:status;default:pending"`
	IsHealthy         bool            `gorm:"column:is_healthy;default:false"`
	ExitReason        string          `gorm:"column:exit_reason"`
	Entrypoint        string          `gorm:"column:entrypoint"`
	Command           string          `gorm:"column:command;default:[]"`
	NoNewPrivileges   bool            `gorm:"column:no_new_privileges;default:true"`
	AddCapabilities   string          `gorm:"column:add_capabilities"`
	MemorySoftLimitMB int64           `gorm:"column:memory_soft_limit_mb"`
	MemoryHardLimitMB int64           `gorm:"column:memory_hard_limit_mb"`
}

type EnvironmentVariable struct {
	Name          string `gorm:"column:name"`
	Value         string `gorm:"column:value"`
	ContainerName string `gorm:"column:container_name;index"`
}

type VolumeMount struct {
	MountPath     string `gorm:"column:mount_path"`
	VolumeUUID    string `gorm:"column:volume_uuid;index"`
	ContainerName string `gorm:"column:container_name;index"`
}

type VolumeType string

const (
	LocalVolume VolumeType = "local"
	NFSVolume   VolumeType = "nfs"
	CIFSVolume  VolumeType = "cifs"
)

type Volume struct {
	UUID        string            `gorm:"column:uuid;primaryKey"`
	Type        VolumeType        `gorm:"column:type;index"`
	LocalConfig LocalVolumeConfig `gorm:"embedded;embeddedPrefix:local_config_"`
	NFSConfig   NFSVolumeConfig   `gorm:"embedded;embeddedPrefix:nfs_config_"`
	CIFSConfig  CIFSVolumeConfig  `gorm:"embedded;embeddedPrefix:cifs_config_"`
}

type LocalVolumeConfig struct {
	IsCustomPath bool   `gorm:"column:is_custom_path"`
	CustomPath   string `gorm:"column:custom_path"`
}

type NFSVolumeConfig struct {
	Host    string `gorm:"column:host"`
	Path    string `gorm:"column:path"`
	Version int    `gorm:"column:version"`
}

type CIFSVolumeConfig struct {
	Share    string `gorm:"column:share"`
	Host     string `gorm:"column:host"`
	Username string `gorm:"column:username"`
	Password string `gorm:"column:password"`
	FileMode string `gorm:"column:file_mode"`
	DirMode  string `gorm:"column:dir_mode"`
	Uid      int    `gorm:"column:uid;default:0"`
	Gid      int    `gorm:"column:gid;default:0"`
}

type DNSEntry struct {
	Domain string `gorm:"column:domain;index" json:"domain"`
	IP     string `gorm:"column:ip;index" json:"ip"`
}

type WireguardPeer struct {
	PublicKey  string `gorm:"column:public_key;primaryKey" json:"public_key"`
	AllowedIPs string `gorm:"column:allowed_ips" json:"allowed_ips"` // Allowed ips - [ip1/cidr1,ip2/cidr2,...]
	EndpointIP string `gorm:"column:endpoint_ip" json:"endpoint_ip"` // ip1:51820
}

type StaticRoute struct {
	Destination string `gorm:"destination"` // ip/cidr format
	Gateway     string `gorm:"gateway"`
}

type NFRule struct {
	UUID  string `gorm:"uuid;primaryKey"`
	Table string `gorm:"table"`
	Chain string `gorm:"chain"`
	Args  string `gorm:"args;default:'[]'"` // json string
}
