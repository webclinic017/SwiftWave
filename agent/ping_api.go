package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func getPing(c echo.Context) error {
	return c.JSON(http.StatusOK, Response{
		Message: "Successfully fetched ping",
		Data:    "pong",
	})
}
