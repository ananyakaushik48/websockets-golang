package main

import (
	"errors"
	"fmt"
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

	handlers map[string]EventHandler
}

// Now that the client list is expected to create a manager
// we create a client list along every new manager created
func NewManager() *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
}

func SendMessage(event Event, c *Client) error {
	fmt.Println(event)
	return nil
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	// Check if event string signature(Type) is a part of event handlers map 
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("no matching event signature")
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
	go client.writeMessages()
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
