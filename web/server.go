package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/starudream/go-lib/core/v2/slog"
	"github.com/starudream/clash-speedtest/api/clash"
	"github.com/starudream/clash-speedtest/job"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	Port        int
	task        *job.Task
	taskMutex   sync.RWMutex
	isRunning   bool
	clients     map[*Client]bool
	clientMutex sync.RWMutex
	broadcast   chan Message
}

type Client struct {
	conn *WebSocketConn
	send chan []byte
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewServer(port int) *Server {
	return &Server{
		Port:      port,
		clients:   make(map[*Client]bool),
		broadcast: make(chan Message, 256),
	}
}

func (s *Server) Start() error {
	// Serve static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return err
	}
	
	http.Handle("/", http.FileServer(http.FS(staticFS)))
	
	// API routes
	http.HandleFunc("/api/config", s.handleGetConfig)
	http.HandleFunc("/api/proxies", s.handleGetProxies)
	http.HandleFunc("/api/speedtest/start", s.handleStartSpeedtest)
	http.HandleFunc("/api/speedtest/stop", s.handleStopSpeedtest)
	http.HandleFunc("/api/speedtest/status", s.handleGetStatus)
	http.HandleFunc("/ws", s.handleWebSocket)
	
	// Start broadcast handler
	go s.handleBroadcast()
	
	addr := fmt.Sprintf(":%d", s.Port)
	slog.Info("Web server starting on http://localhost%s", addr)
	fmt.Printf("\n🌐 Web UI: http://localhost%s\n\n", addr)
	
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	config := map[string]interface{}{
		"clash_addr":     "http://127.0.0.1:9090",
		"clash_secret":   "",
		"clash_proxy":    "",
		"size":           10,
		"threads":        1,
		"download":       "cloudflare",
		"concurrent":     1,
		"ping":           false,
		"ping_interval":  60,
		"ping_timeout":   5000,
		"timeout":        60,
		"format":         "txt",
	}
	
	s.sendJSON(w, config)
}

func (s *Server) handleGetProxies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse query parameters
	clashAddr := r.URL.Query().Get("clash_addr")
	if clashAddr == "" {
		clashAddr = "http://127.0.0.1:9090"
	}
	clashSecret := r.URL.Query().Get("clash_secret")
	
	client := clash.NewClient(clashAddr, clashSecret)
	
	providers, err := client.GetProviderProxies()
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to get proxies: %v", err), http.StatusInternalServerError)
		return
	}
	
	proxies := providers.FilterProxies()
	s.sendJSON(w, map[string]interface{}{
		"proxies": proxies,
		"total":   len(proxies),
	})
}

func (s *Server) handleStartSpeedtest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.taskMutex.Lock()
	if s.isRunning {
		s.taskMutex.Unlock()
		s.sendError(w, "Speedtest is already running", http.StatusConflict)
		return
	}
	s.isRunning = true
	s.taskMutex.Unlock()
	
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		s.taskMutex.Lock()
		s.isRunning = false
		s.taskMutex.Unlock()
		s.sendError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	
	// Start speedtest in background
	go func() {
		defer func() {
			s.taskMutex.Lock()
			s.isRunning = false
			s.taskMutex.Unlock()
		}()
		
		s.broadcast <- Message{
			Type: "log",
			Data: map[string]interface{}{
				"level":   "info",
				"message": "Starting speedtest...",
				"time":    time.Now().Format("15:04:05"),
			},
		}
		
		// Run speedtest with config
		err := s.runSpeedtest(config)
		if err != nil {
			s.broadcast <- Message{
				Type: "error",
				Data: map[string]interface{}{
					"message": fmt.Sprintf("Speedtest failed: %v", err),
					"time":    time.Now().Format("15:04:05"),
				},
			}
		} else {
			s.broadcast <- Message{
				Type: "complete",
				Data: map[string]interface{}{
					"message": "Speedtest completed successfully",
					"time":    time.Now().Format("15:04:05"),
				},
			}
		}
	}()
	
	s.sendJSON(w, map[string]interface{}{
		"status":  "started",
		"message": "Speedtest started successfully",
	})
}

func (s *Server) handleStopSpeedtest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()
	
	if !s.isRunning {
		s.sendError(w, "No speedtest is running", http.StatusConflict)
		return
	}
	
	// TODO: Implement graceful stop
	s.isRunning = false
	
	s.sendJSON(w, map[string]interface{}{
		"status":  "stopped",
		"message": "Speedtest stopped",
	})
}

func (s *Server) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.taskMutex.RLock()
	isRunning := s.isRunning
	s.taskMutex.RUnlock()
	
	s.sendJSON(w, map[string]interface{}{
		"is_running": isRunning,
	})
}

func (s *Server) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}

func (s *Server) handleBroadcast() {
	for msg := range s.broadcast {
		s.clientMutex.RLock()
		for client := range s.clients {
			select {
			case client.send <- s.marshalMessage(msg):
			default:
				close(client.send)
				delete(s.clients, client)
			}
		}
		s.clientMutex.RUnlock()
	}
}

func (s *Server) marshalMessage(msg Message) []byte {
	data, _ := json.Marshal(msg)
	return data
}

func (s *Server) runSpeedtest(config map[string]interface{}) error {
	// Convert broadcast channel to interface{} channel
	broadcastChan := make(chan interface{}, 256)
	
	// Forward messages from interface{} channel to Message channel
	go func() {
		for msg := range broadcastChan {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				msgType := ""
				if t, ok := msgMap["type"].(string); ok {
					msgType = t
				}
				data := msgMap["data"]
				s.broadcast <- Message{
					Type: msgType,
					Data: data,
				}
			}
		}
	}()
	
	task := job.NewTaskFromConfig(config, broadcastChan)
	err := task.Run()
	close(broadcastChan)
	return err
}
