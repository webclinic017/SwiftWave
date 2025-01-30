package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func fetchAllWireguardPeers(c echo.Context) error {
	peers, err := FetchAllWireguardPeers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch wireguard peers",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Wireguard peers fetched successfully",
		Data:    peers,
	})
}

func createWireguardPeer(c echo.Context) error {
	var p WireguardPeer
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}
	if err := p.Create(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to create wireguard peer",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Wireguard peer created successfully",
		Data:    p,
	})
}

func deleteWireguardPeer(c echo.Context) error {
	publicKey := c.Param("publicKey")
	peer, err := FetchWireguardPeerByPublicKey(publicKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to fetch wireguard peer",
			Error:   err.Error(),
		})
	}
	if err := peer.Remove(); err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to delete wireguard peer",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Wireguard peer deleted successfully",
	})
}

func updateWireguardPeer(c echo.Context) error {
	publicKey := c.Param("publicKey")
	var update WireguardPeerUpdate
	if err := c.Bind(&update); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}
	err := UpdateEndpointIP(publicKey, update.EndpointIP)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to update wireguard peer",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Wireguard peer updated successfully",
	})
}

func configureWireguardPeers(c echo.Context) error {
	err := ConfigureWireguardPeers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Message: "Failed to configure wireguard peers",
			Error:   err.Error(),
		})
	}
	return c.JSON(http.StatusOK, Response{
		Message: "Wireguard peers configured successfully",
	})
}
