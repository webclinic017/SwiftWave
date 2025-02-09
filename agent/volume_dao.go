package main

import (
	"fmt"
)

func (v *Volume) Validate() error {
	if v == nil {
		return fmt.Errorf("provided record is nil")
	}
	if v.UUID == "" {
		return fmt.Errorf("UUID is required for volume")
	}
	switch v.Type {
	case LocalVolume:
		if v.LocalConfig.IsCustomPath && v.LocalConfig.CustomPath == "" {
			return fmt.Errorf("custom path is required for local volume")
		}
	case NFSVolume:
		if v.NFSConfig.Host == "" {
			return fmt.Errorf("host is required for NFS volume")
		}
		if v.NFSConfig.Path == "" {
			return fmt.Errorf("path is required for NFS volume")
		}

	case CIFSVolume:
		if v.CIFSConfig.Host == "" {
			return fmt.Errorf("host is required for CIFS volume")
		}
		if v.CIFSConfig.Share == "" {
			return fmt.Errorf("share is required for CIFS volume")
		}
		if v.CIFSConfig.FileMode == "" {
			return fmt.Errorf("file mode is required for CIFS volume")
		}
		if v.CIFSConfig.DirMode == "" {
			return fmt.Errorf("dir mode is required for CIFS volume")
		}
	}
	return nil
}

func (v *Volume) Create() error {
	if err := v.Validate(); err != nil {
		return err
	}
	// Check if the volume already exists
	exists, err := ExistsVolume(v.UUID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("volume already exists")
	}
	// Create the record
	if err := rwDB.Create(v).Error; err != nil {
		return err
	}
	// Try to create the volume
	if err := v.CreateVolume(); err != nil {
		// If volume creation fails, remove the record
		_ = v.Delete(false)
		return err
	}
	return nil
}

func (v *Volume) Delete(deleteDirectory bool) error {
	// Pass the deleteDirectory flag to remove the data directory in case of local volume
	// Do this if we get this explicitly from the caller
	err := v.RemoveVolume(deleteDirectory)
	if err != nil {
		return err
	}
	return rwDB.Delete(v).Error
}

func FetchVolumeByUUID(uuid string) (*Volume, error) {
	var volume Volume
	if err := rDB.Where("uuid = ?", uuid).First(&volume).Error; err != nil {
		return nil, err
	}
	return &volume, nil
}

func FetchAllVolumes() ([]Volume, error) {
	var volumes []Volume
	if err := rDB.Find(&volumes).Error; err != nil {
		return nil, err
	}
	return volumes, nil
}

func ExistsVolume(uuid string) (bool, error) {
	var count int64
	if err := rDB.Model(&Volume{}).Where("uuid = ?", uuid).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
