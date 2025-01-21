package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func createDNSRecord(c echo.Context) error {
	var dnsEntry DNSEntry
	if err := c.Bind(&dnsEntry); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}
	if err := dnsEntry.Create(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to create DNS record",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "DNS record created successfully",
		Data:    dnsEntry,
	})
}

func deleteDNSRecord(c echo.Context) error {
	var dnsEntry DNSEntry
	if err := c.Bind(&dnsEntry); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}
	if err := dnsEntry.Delete(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to delete DNS record",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "DNS record deleted successfully",
		Data:    dnsEntry,
	})
}

func fetchAllDNSRecords(c echo.Context) error {
	dnsEntries, err := FetchAllDNSRecords()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch DNS records",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "DNS records fetched successfully",
		Data:    dnsEntries,
	})
}

func fetchDNSRecordsByDomain(c echo.Context) error {
	domain := c.Param("domain")
	dnsEntries, err := FetchDNSRecordsByDomain(domain)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch DNS records",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "DNS records fetched successfully",
		Data:    dnsEntries,
	})
}
