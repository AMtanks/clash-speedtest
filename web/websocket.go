package web

import (
	"net/http"
	"time"

	"github.com/starudream/go-lib/core/v2/slog"
)

type WebSocketConn struct {
	w http.ResponseWriter
	r *http.Request
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check if it's a WebSocket upgrade request
	if r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "Expected WebSocket connection", http.StatusBadRequest)
		return
	}
	
	// Create SSE connection as fallback
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	
	client := &Client{
		conn: &WebSocketConn{w: w, r: r},
		send: make(chan []byte, 256),
	}
	
	s.clientMutex.Lock()
	s.clients[client] = true
	s.clientMutex.Unlock()
	
	defer func() {
		s.clientMutex.Lock()
		delete(s.clients, client)
		s.clientMutex.Unlock()
		close(client.send)
	}()
	
	slog.Info("WebSocket client connected")
	
	// Send initial connection message
	w.Write([]byte("data: {\"type\":\"connected\",\"data\":{\"message\":\"Connected to speedtest server\"}}\n\n"))
	flusher.Flush()
	
	// Listen for messages
	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				return
			}
			
			w.Write([]byte("data: "))
			w.Write(message)
			w.Write([]byte("\n\n"))
			flusher.Flush()
			
		case <-r.Context().Done():
			slog.Info("WebSocket client disconnected")
			return
		}
	}
}
