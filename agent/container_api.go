package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func createContainer(c echo.Context) error {
	var container Container
	if err := c.Bind(&container); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Failed to bind request",
			Error:   err.Error(),
		})
	}
	err := container.Create()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to create container",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Successfully created container",
		Data:    container,
	})
}

func deleteContainer(c echo.Context) error {
	uuid := c.Param("uuid")
	container, err := FetchContainerByUUID(uuid)
	if err != nil {
		return c.JSON(http.StatusNotFound, Response{
			Message: "Failed to fetch container",
			Error:   err.Error(),
		})
	}
	err = container.Remove()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to remove container",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Successfully removed container",
		Data:    nil,
	})
}

func statusOfContainer(c echo.Context) error {
	uuid := c.Param("uuid")

	container, err := FetchContainerByUUID(uuid)
	if err != nil {
		return c.JSON(http.StatusNotFound, Response{
			Message: "Failed to fetch container",
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, Response{
		Message: "Successfully fetched status",
		Data:    container.GetStatus(),
	})
}
