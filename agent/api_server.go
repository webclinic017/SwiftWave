package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
)

func startHttpServer() {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.Recover())

	config, err := GetConfig()
	if err != nil {
		log.Fatalf("Failed to fetch config: %v", err)
	}

	// Auth middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get("Authorization")
			err := bcrypt.CompareHashAndPassword([]byte(config.AuthTokenHash), []byte(token))
			if err != nil || token == "" || config.AuthTokenHash == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}
			return next(c)
		}
	})

	// Volume API
	e.GET("/volumes", fetchAllVolumes)
	e.GET("/volumes/:uuid", fetchVolume)
	e.POST("/volumes/:uuid/size", fetchVolumeSize)
	e.POST("/volumes", createVolume)
	e.DELETE("/volumes/:uuid", deleteVolume)

	// DNS Record API
	e.GET("/dns", fetchAllDNSRecords)
	e.GET("/dns/:domain", fetchDNSRecordsByDomain)
	e.DELETE("/dns", deleteDNSRecord)
	e.POST("/dns", createDNSRecord)

	// Wireguard Peer API
	e.GET("/wireguard/peers", fetchAllWireguardPeers)
	e.POST("/wireguard/peers", createWireguardPeer)
	e.DELETE("/wireguard/peers/:publicKey", deleteWireguardPeer)
	e.PUT("/wireguard/peers/:publicKey", updateWireguardPeer)
	e.POST("/wireguard/peers/configure", configureWireguardPeers)

	// Static Route API
	e.GET("/static-routes", fetchAllStaticRoutes)
	e.POST("/static-routes", createStaticRoute)
	e.GET("/static-routes/:ip/:cidr", fetchStaticRouteByDestination)
	e.DELETE("/static-routes/:ip/:cidr", deleteStaticRoute)

	// NF Rule API
	e.GET("/nf-rules", fetchAllNFRules)
	e.POST("/nf-rules", createNFRule)
	e.GET("/nf-rules/:uuid", getNFRule)
	e.DELETE("/nf-rules/:uuid", deleteNFRule)

	// HAProxy API
	e.GET("/haproxy/service-status", getHAProxyStatus)
	e.Any("/haproxy/*", sendRequestToHAProxy)

	// Docker API
	e.Any("/docker/*", sendRequestToDocker)

	// Container API
	e.POST("/containers", createContainer)
	e.DELETE("/containers/:uuid", deleteContainer)
	e.GET("/containers/:uuid/status", statusOfContainer)

	ip, _, err := net.ParseCIDR(config.WireguardConfig.Address)
	if err != nil {
		log.Fatalf("Failed to parse wireguard address: %v", err)
	}

	// Log API
	e.GET("/journald/stream", streamJournalLogs)

	if err := e.Start(fmt.Sprintf("%s:3332", ip.String())); err != nil && !errors.Is(err, http.ErrServerClosed) {
		e.Logger.Fatal(err)
	}
}
