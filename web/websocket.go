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
	// Use Server-Sent Events (SSE) for real-time updates
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
	
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
		slog.Info("SSE client disconnected")
	}()
	
	slog.Info("SSE client connected")
	
	// Send initial connection message
	initialMsg := []byte("data: {\"type\":\"connected\",\"data\":{\"message\":\"Connected to speedtest server\"}}\n\n")
	w.Write(initialMsg)
	flusher.Flush()
	
	// Send a test message after 1 second
	go func() {
		time.Sleep(1 * time.Second)
		select {
		case client.send <- []byte("{\"type\":\"log\",\"data\":{\"level\":\"info\",\"message\":\"SSE connection established\",\"time\":\"" + time.Now().Format("15:04:05") + "\"}}"):
		case <-r.Context().Done():
		}
	}()
	
	// Listen for messages
	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				return
			}
			
			// Write SSE format: data: <json>\n\n
			w.Write([]byte("data: "))
			w.Write(message)
			w.Write([]byte("\n\n"))
			flusher.Flush()
			
		case <-r.Context().Done():
			return
		}
	}
}
