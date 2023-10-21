package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// Because we expect many socket connections coming in
// we prevent deadlock with a sync READWRITE mutex in the Manager
type Manager struct {
	clients ClientList
	sync.RWMutex
}

// Now that the client list is expected to create a manager
// we create a client list along every new manager created
func NewManager() *Manager {
	return &Manager{
		clients: make(ClientList),
	}
}

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")
	// upgrade regular http into websocket
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn, m)

	m.addClient(client)
}

func (m *Manager) addClient(client *Client) {
	// If two people are connecting at the same time
	// lock the manager and don't modify the client map, we do this with
	// m.Lock()
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true

	// Start client processes
	go client.readMessages()
}

func (m *Manager) removeClient(client *Client) {
	// Same procedure followed here,
	// The client map is locked till the client value is updated
	// it is unlocked after the code
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}
