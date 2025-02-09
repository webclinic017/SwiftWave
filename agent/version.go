package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const Version = "0.0.1"

func getVersion(c echo.Context) error {
	return c.JSON(http.StatusOK, Response{
		Message: "Successfully fetched version",
		Data:    Version,
	})
}
