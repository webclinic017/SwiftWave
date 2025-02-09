package main

import (
	"os"
	"time"
)

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error"`
}

type FileInfo struct {
	Name         string      `json:"name"`
	RelativePath string      `json:"relative_path"`
	Size         int64       `json:"size"`
	Mode         os.FileMode `json:"mode"`
	ModTime      time.Time   `json:"mod_time"`
	UID          uint        `json:"uid"`
	GID          uint        `json:"gid"`
	IsDir        bool        `json:"is_dir"`
}

type WireguardPeerUpdate struct {
	EndpointIP string `json:"endpoint_ip"`
}

type StaticConfig struct {
	Content string `json:"content"`
}
