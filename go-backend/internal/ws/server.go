package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"

	"go-backend/internal/auth"
	"go-backend/internal/security"
	"go-backend/internal/store/sqlite"
)

type encryptedMessage struct {
	Encrypted bool   `json:"encrypted"`
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

type broadcastMessage struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
	Data string `json:"data"`
}

type connWrap struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type nodeSession struct {
	nodeID int64
	secret string
	conn   *connWrap
}

type Server struct {
	repo      *sqlite.Repository
	jwtSecret string
	upgrader  websocket.Upgrader

	mu     sync.RWMutex
	admins map[*connWrap]struct{}
	nodes  map[int64]*nodeSession
	byConn map[*websocket.Conn]*nodeSession
}

func NewServer(repo *sqlite.Repository, jwtSecret string) *Server {
	return &Server{
		repo:      repo,
		jwtSecret: jwtSecret,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		admins: make(map[*connWrap]struct{}),
		nodes:  make(map[int64]*nodeSession),
		byConn: make(map[*websocket.Conn]*nodeSession),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	typeVal := query.Get("type")
	secret := query.Get("secret")

	if typeVal == "1" {
		node, err := s.repo.GetNodeBySecret(secret)
		if err != nil || node == nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		s.handleNode(w, r, node.ID, secret)
		return
	}

	if typeVal == "0" {
		if _, ok := auth.ValidateToken(secret, s.jwtSecret); !ok {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		s.handleAdmin(w, r)
		return
	}

	http.Error(w, "bad request", http.StatusBadRequest)
}

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	cw := &connWrap{conn: conn}

	s.mu.Lock()
	s.admins[cw] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.admins, cw)
		s.mu.Unlock()
		_ = conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (s *Server) handleNode(w http.ResponseWriter, r *http.Request, nodeID int64, secret string) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	cw := &connWrap{conn: conn}

	version := r.URL.Query().Get("version")
	httpVal := parseIntDefault(r.URL.Query().Get("http"), 0)
	tlsVal := parseIntDefault(r.URL.Query().Get("tls"), 0)
	socksVal := parseIntDefault(r.URL.Query().Get("socks"), 0)

	s.mu.Lock()
	if old, ok := s.nodes[nodeID]; ok {
		_ = old.conn.conn.Close()
		delete(s.byConn, old.conn.conn)
	}
	ns := &nodeSession{nodeID: nodeID, secret: secret, conn: cw}
	s.nodes[nodeID] = ns
	s.byConn[conn] = ns
	s.mu.Unlock()

	_ = s.repo.UpdateNodeOnline(nodeID, 1, version, httpVal, tlsVal, socksVal)
	s.broadcastStatus(nodeID, 1)

	defer func() {
		needOfflineBroadcast := false
		s.mu.Lock()
		current, ok := s.nodes[nodeID]
		if ok && current.conn.conn == conn {
			delete(s.nodes, nodeID)
			needOfflineBroadcast = true
		}
		delete(s.byConn, conn)
		s.mu.Unlock()
		if needOfflineBroadcast {
			_ = s.repo.UpdateNodeStatus(nodeID, 0)
			s.broadcastStatus(nodeID, 0)
		}
		_ = conn.Close()
	}()

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}

		msg := decryptIfNeeded(payload, secret)
		s.broadcastInfo(nodeID, msg)
	}
}

func (s *Server) broadcastStatus(nodeID int64, status int) {
	payload := map[string]interface{}{
		"id":   strconv.FormatInt(nodeID, 10),
		"type": "status",
		"data": status,
	}
	raw, _ := json.Marshal(payload)
	s.broadcastToAdmins(string(raw))
}

func (s *Server) broadcastInfo(nodeID int64, data string) {
	payload := broadcastMessage{ID: nodeID, Type: "info", Data: data}
	raw, _ := json.Marshal(payload)
	s.broadcastToAdmins(string(raw))
}

func (s *Server) broadcastToAdmins(message string) {
	s.mu.RLock()
	admins := make([]*connWrap, 0, len(s.admins))
	for c := range s.admins {
		admins = append(admins, c)
	}
	s.mu.RUnlock()

	for _, c := range admins {
		c.mu.Lock()
		err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
		c.mu.Unlock()
		if err != nil {
			log.Printf("websocket broadcast failed: %v", err)
		}
	}
}

func decryptIfNeeded(payload []byte, secret string) string {
	text := string(payload)
	var wrap encryptedMessage
	if err := json.Unmarshal(payload, &wrap); err != nil || !wrap.Encrypted || strings.TrimSpace(wrap.Data) == "" {
		return text
	}

	crypto, err := security.NewAESCrypto(secret)
	if err != nil {
		return text
	}
	plain, err := crypto.Decrypt(wrap.Data)
	if err != nil {
		return text
	}
	return string(plain)
}

func parseIntDefault(v string, fallback int) int {
	x, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return x
}
