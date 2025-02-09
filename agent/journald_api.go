package main

import (
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

func streamJournalLogs(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		var req JournalRequest
		if err := websocket.JSON.Receive(ws, &req); err != nil {
			c.Logger().Error("Failed to receive request:", err)
			return
		}

		if req.Key == "" || req.Value == "" || len(req.Fields) == 0 {
			websocket.JSON.Send(ws, map[string]string{"error": "missing required fields"})
			return
		}

		// Parse the since_time parameter
		var sinceTime time.Time
		if req.SinceTime != "" {
			var err error
			sinceTime, err = time.Parse(time.RFC3339, req.SinceTime)
			if err != nil {
				websocket.JSON.Send(ws, map[string]string{"error": "invalid since_time format, use RFC3339"})
				return
			}
		}

		dataChan := make(chan map[string]string)
		stopChan := make(chan struct{})
		defer close(stopChan)

		go func() {
			if err := WatchJournalLogs(req.Key, req.Value, req.Fields, sinceTime, dataChan, stopChan); err != nil {
				websocket.JSON.Send(ws, map[string]string{"error": err.Error()})
			}
		}()

		for {
			select {
			case data, ok := <-dataChan:
				if !ok {
					return // Channel closed
				}
				if err := websocket.JSON.Send(ws, data); err != nil {
					c.Logger().Debug("WebSocket send error:", err)
					return
				}
			case <-c.Request().Context().Done():
				c.Logger().Debug("Client disconnected")
				return
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
