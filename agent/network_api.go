package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func createStaticRoute(c echo.Context) error {
	var s StaticRoute
	if err := c.Bind(&s); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}
	if err := s.Create(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to create static route",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Static route created",
		Data:    s,
	})
}

func deleteStaticRoute(c echo.Context) error {
	destination := c.QueryParam("destination")
	if destination == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   "destination is required",
		})
	}
	record, err := FetchStaticRouteByDestination(destination)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch static route",
			Error:   err.Error(),
		})
	}
	if err := record.Delete(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to delete static route",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Static route deleted",
	})
}

func fetchAllStaticRoutes(c echo.Context) error {
	routes, err := FetchAllStaticRoutes()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch static routes",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Static routes fetched",
		Data:    routes,
	})
}
