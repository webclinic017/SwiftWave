package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func createVolume(c echo.Context) error {
	var v Volume
	if err := c.Bind(&v); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}
	if err := v.Create(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to create volume",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Volume created successfully",
		Data:    v,
	})
}

func deleteVolume(c echo.Context) error {
	volumeUUID := c.Param("uuid")
	if volumeUUID == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "UUID is required",
		})
	}
	v, err := FetchVolumeByUUID(volumeUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch volume",
			Error:   err.Error(),
		})
	}
	if err := v.RemoveVolume(true); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to remove volume",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Volume removed successfully",
		Data:    v,
	})
}

func fetchAllVolumes(c echo.Context) error {
	volumes, err := FetchAllVolumes()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch volumes",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Volumes fetched successfully",
		Data:    volumes,
	})
}

func fetchVolume(c echo.Context) error {
	volumeUUID := c.Param("uuid")
	if volumeUUID == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "UUID is required",
		})
	}
	v, err := FetchVolumeByUUID(volumeUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch volume",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Volume fetched successfully",
		Data:    v,
	})
}

func fetchVolumeSize(c echo.Context) error {
	volumeUUID := c.Param("uuid")
	if volumeUUID == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "UUID is required",
		})
	}
	v, err := FetchVolumeByUUID(volumeUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch volume",
			Error:   err.Error(),
		})
	}
	size, err := v.Size()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch volume size",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Volume size fetched successfully",
		Data:    size,
	})
}
