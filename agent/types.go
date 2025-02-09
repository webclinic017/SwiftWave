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

type StaticConfigs []StaticConfig

type StaticConfig struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type JournalRequest struct {
	Key       string   `json:"key"`
	Value     string   `json:"value"`
	Fields    []string `json:"fields"`
	SinceTime string   `json:"since_time"` // RFC3339 format timestamp
}
