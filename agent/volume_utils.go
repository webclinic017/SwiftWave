package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/volume"
)

var volumeBindsDefaultPath string

func init() {
	volumePath, err := filepath.Abs("/home/tanmoy/docker-volumes")
	if err != nil {
		fmt.Println("Failed to get absolute path for volumes directory")
		panic(err)
	}
	volumeBindsDefaultPath = volumePath
}

func (v *Volume) LocalVolumeFullPath() string {
	if v.LocalConfig.IsCustomPath {
		return v.LocalConfig.CustomPath
	}
	return filepath.Join(volumeBindsDefaultPath, v.UUID)
}

func ExistsDockerVolume(volumeUUID string) bool {
	_, err := dockerClient.VolumeInspect(context.Background(), volumeUUID)
	return err == nil
}

func (v *Volume) RemoveVolume(deleteDirectory bool) error {
	if !ExistsDockerVolume(v.UUID) {
		return nil
	}

	// Remove forcefully
	err := dockerClient.VolumeRemove(context.Background(), v.UUID, true)
	if err != nil {
		return err
	}

	// Remove volume directory
	if v.Type == LocalVolume && deleteDirectory {
		path := v.LocalVolumeFullPath()
		_ = os.RemoveAll(path)
	}
	return nil
}

func (v *Volume) CreateVolume() error {
	switch v.Type {
	case LocalVolume:
		return createLocalVolume(v)
	case NFSVolume:
		return createNFSVolume(v)
	case CIFSVolume:
		return createCIFSVolume(v)
	default:
		return fmt.Errorf("unsupported volume type: %s", v.Type)
	}
}

func (v *Volume) Size() (int64, error) {
	return 0, nil
}

func (v *Volume) Backup(uploadUrl string) error {
	return nil
}

func (v *Volume) Restore(downloadUrl string) error {
	return nil
}

// Private functions

func createLocalVolume(v *Volume) error {
	// create volume directory
	err := os.MkdirAll(v.LocalVolumeFullPath(), 0755)
	if err != nil {
		return err
	}

	_, err = dockerClient.VolumeCreate(context.Background(), volume.CreateOptions{
		Name:   v.UUID,
		Driver: "local",
		DriverOpts: map[string]string{
			"type":   "volume",
			"o":      "bind",
			"device": v.LocalVolumeFullPath(),
		},
	})
	return err
}

func createNFSVolume(v *Volume) error {
	_, err := dockerClient.VolumeCreate(context.Background(), volume.CreateOptions{
		Name:   v.UUID,
		Driver: "local",
		DriverOpts: map[string]string{
			"type":   "nfs",
			"o":      "addr=" + v.NFSConfig.Host + ",rw,nfsvers=" + fmt.Sprint(v.NFSConfig.Version),
			"device": ":" + v.NFSConfig.Path,
		},
	})
	return err
}

func createCIFSVolume(v *Volume) error {
	_, err := dockerClient.VolumeCreate(context.Background(), volume.CreateOptions{
		Name:   v.UUID,
		Driver: "local",
		DriverOpts: map[string]string{
			"type":   "cifs",
			"o":      fmt.Sprintf("addr=%s,username=%s,password=%s,file_mode=%s,dir_mode=%s,uid=%d,gid=%d", v.CIFSConfig.Host, v.CIFSConfig.Username, v.CIFSConfig.Password, v.CIFSConfig.FileMode, v.CIFSConfig.DirMode, v.CIFSConfig.Uid, v.CIFSConfig.Gid),
			"device": v.CIFSConfig.Share,
		},
	})
	return err
}
