package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn        *websocket.Conn
	send        chan []byte
	currentSite string // "brazil", "usa", etc., or "" for all
}

type WSClientMessage struct {
	Action string `json:"action"`
	Site   string `json:"site"`
}

type Hub struct {
	clients       map[*Client]bool
	siteInterests map[string]int // site -> client count interested
	mutex         sync.Mutex
	upgrader      websocket.Upgrader
	subscribeFn   func(subject string) error
	unsubscribeFn func(subject string)
}

func NewHub(subscribeFn func(subject string) error, unsubscribeFn func(subject string)) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		siteInterests: make(map[string]int),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		subscribeFn:   subscribeFn,
		unsubscribeFn: unsubscribeFn,
	}
}

func getSubjectForSite(site string) string {
	if site == "" {
		return "*.>"
	}
	return fmt.Sprintf("iot_domain_%s.>", site)
}

func extractSiteFromSubject(subject string) string {
	parts := strings.Split(subject, ".")
	if len(parts) > 0 {
		first := parts[0]
		subParts := strings.Split(first, "_")
		if len(subParts) >= 3 && subParts[0] == "iot" && subParts[1] == "domain" {
			return subParts[2]
		}
	}
	return ""
}

func (h *Hub) registerInterest(c *Client, newSite string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	oldSite := c.currentSite
	if oldSite == newSite {
		return
	}

	// 1. decrement interest in the old site
	if oldSite != "NONE" {
		h.siteInterests[oldSite]--
		if h.siteInterests[oldSite] <= 0 {
			delete(h.siteInterests, oldSite)
			oldSubject := getSubjectForSite(oldSite)
			log.Printf("[NATS] No clients interested. Unsubscribing from: %s", oldSubject)
			go h.unsubscribeFn(oldSubject)
		}
	}

	// 2. increment interest in the new site
	c.currentSite = newSite
	h.siteInterests[newSite]++
	if h.siteInterests[newSite] == 1 {
		newSubject := getSubjectForSite(newSite)
		log.Printf("[NATS] First client interested. Subscribing to: %s", newSubject)
		go func(sub string) {
			if err := h.subscribeFn(sub); err != nil {
				log.Printf("[NATS] Error subscribing to topic %s: %v", sub, err)
			}
		}(newSubject)
	}
}

// broadcast sends a filtered message to clients based on site interest
func (h *Hub) Broadcast(subject string, msg []byte) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	msgSite := extractSiteFromSubject(subject)

	for client := range h.clients {
		// send if client wants "all" ("") or if client's site matches the message site
		if client.currentSite == "" || client.currentSite == msgSite {
			select {
			case client.send <- msg:
			default:
				// discard if client is slow
			}
		}
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	log.Println("Client connected")

	client := &Client{
		conn:        conn,
		send:        make(chan []byte, 100),
		currentSite: "NONE", // initialize with NONE to force initial registration in registerInterest
	}

	h.addClient(client)

	// register initial interest in "" (all plants)
	h.registerInterest(client, "")

	go h.writePump(client)
	go h.readPump(client)
}

func (h *Hub) addClient(c *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.clients[c] = true
}

func (h *Hub) removeClient(c *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, exists := h.clients[c]; exists {
		delete(h.clients, c)
		close(c.send)
		log.Println("Client disconnected")

		// remove client interest in current site
		oldSite := c.currentSite
		if oldSite != "NONE" {
			h.siteInterests[oldSite]--
			if h.siteInterests[oldSite] <= 0 {
				delete(h.siteInterests, oldSite)
				oldSubject := getSubjectForSite(oldSite)
				log.Printf("[NATS] No clients interested. Unsubscribing from: %s", oldSubject)
				go h.unsubscribeFn(oldSubject)
			}
		}
	}
}

func (h *Hub) writePump(c *Client) {
	defer c.conn.Close()
	for msg := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}

func (h *Hub) readPump(c *Client) {
	defer func() {
		h.removeClient(c)
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var clientMsg WSClientMessage
		if err := json.Unmarshal(message, &clientMsg); err == nil {
			if clientMsg.Action == "subscribe" {
				log.Printf("Client requested subscription change to site: %q", clientMsg.Site)
				h.registerInterest(c, clientMsg.Site)
			}
		} else {
			log.Printf("Invalid message format received from client: %v", err)
		}
	}
}
