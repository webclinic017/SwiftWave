package main

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/labstack/echo/v4"
)

func sendRequestToDocker(c echo.Context) error {
	// Create a reverse proxy to the Docker socket
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = "docker"
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/docker")
			req.Host = "docker"
		},
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	// Serve the request
	proxy.ServeHTTP(c.Response(), c.Request())
	return nil
}
