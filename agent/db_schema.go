package main

type ContainerStatus string

const (
	ContainerStatusImagePullPending      ContainerStatus = "image_pull_pending"
	ContainerStatusImagePulling          ContainerStatus = "image_pulling"
	ContainerStatusImagePullAuthError    ContainerStatus = "image_pull_auth_error"
	ContainerStatusImagePullArchMismatch ContainerStatus = "image_pull_arch_mismatch"
	ContainerStatusImagePullFailed       ContainerStatus = "image_pull_failed"
	ContainerStatusImagePulled           ContainerStatus = "image_pulled"
	ContainerStatusCreated               ContainerStatus = "created"
	ContainerStatusCreationFailed        ContainerStatus = "creation_failed"

	// Logical status - fetch from docker

	ContainerStatusRunning    ContainerStatus = "running"
	ContainerStatusRestarting ContainerStatus = "restarting"
	ContainerStatusExited     ContainerStatus = "exited"
	ContainerStatusPaused     ContainerStatus = "paused"
	ContainerStatusNotFound   ContainerStatus = "not_found"
)

type Container struct {
	UUID            string          `gorm:"column:uuid;primaryKey"`
	ImageURI        string          `gorm:"column:image_uri"`
	ImageAuthHeader string          `gorm:"column:image_auth_header"`
	ImagePulled     bool            `gorm:"column:image_pulled"`
	Data            string          `gorm:"column:data"`
	StaticConfigs   string          `gorm:"column:static_configs"` // json string of []StaticConfig
	Status          ContainerStatus `gorm:"column:status"`
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
