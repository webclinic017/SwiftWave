package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config/bootstrap"
)

// Initialize : Initialize the server and its routes
func (server *Server) Initialize() {
	// Initiating Routes for ACME Challenge
	server.ServiceManager.SslManager.InitHttpHandlers(server.EchoServer)
	// Initiating Routes for Project
	server.initiateProjectRoutes(server.EchoServer)
}

func (server *Server) initiateProjectRoutes(e *echo.Echo) {
	// Initiating Routes for Healthcheck
	e.GET("/healthcheck", server.healthcheck)
	// Initiating Routes for Version
	e.GET("/version", server.version)
	// Initiating Routes for Auth
	e.POST("/auth/login", server.login)
	e.GET("/verify-auth", server.verifyAuth)
	// Initiating Routes for Project
	e.POST("/upload/code", server.uploadTarFile)
	// Initiating Routes for PersistentVolume
	e.GET("/persistent-volume/backup/:id/download", server.downloadPersistentVolumeBackup)
	e.GET("/persistent-volume/backup/:id/filename", server.getPersistentVolumeBackupFileName)
	e.POST("/persistent-volume/:id/restore", server.uploadPersistentVolumeRestoreFile)
	// Initiating Routes for Webhook
	e.Any("/webhook/redeploy-app/:app-id/:webhook-token", server.redeployApp)
	// Initiating Routes for fetch and update system config
	e.GET("/config/system", bootstrap.FetchSystemConfigHandler)
	e.PUT("/config/system", bootstrap.UpdateSystemConfigHandler)
	// analytics
	e.POST("/service/analytics", server.analytics)
	// serve log file
	e.GET("/log/:log_file_name", server.fetchLog)
}
