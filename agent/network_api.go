package main

import (
	"fmt"
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
	ip := c.QueryParam("ip")
	cidr := c.QueryParam("cidr")
	if ip == "" || cidr == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   "ip and cidr are required",
		})
	}
	destination := fmt.Sprintf("%s/%s", ip, cidr)
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

func fetchStaticRouteByDestination(c echo.Context) error {
	ip := c.QueryParam("ip")
	cidr := c.QueryParam("cidr")
	if ip == "" || cidr == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   "ip and cidr are required",
		})
	}
	destination := fmt.Sprintf("%s/%s", ip, cidr)
	if destination == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   "destination is required",
		})
	}
	route, err := FetchStaticRouteByDestination(destination)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch static route",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Static route fetched",
		Data:    route,
	})
}

func createNFRule(c echo.Context) error {
	var r NFRule
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}
	if err := r.Create(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to create nf rule",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "NF rule created",
		Data:    r,
	})
}

func getNFRule(c echo.Context) error {
	uuid := c.Param("uuid")
	if uuid == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   "uuid is required",
		})
	}
	record, err := FetchNFRuleByUUID(uuid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch nf rule",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "NF rule fetched",
		Data:    record,
	})
}

func deleteNFRule(c echo.Context) error {
	uuid := c.Param("uuid")
	if uuid == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   "uuid is required",
		})
	}
	record, err := FetchNFRuleByUUID(uuid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch nf rule",
			Error:   err.Error(),
		})
	}
	if err := record.Delete(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to delete nf rule",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "NF rule deleted",
	})
}

func fetchAllNFRules(c echo.Context) error {
	rules, err := FetchAllNFRules()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch nf rules",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "NF rules fetched",
		Data:    rules,
	})
}
