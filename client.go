package main

import (
	"log"

	"github.com/gorilla/websocket"
)

// A map that contains all the client connections
type ClientList map[*Client]bool

// Client Structure contains pointers to
// WebSocket Connection
// Connection Manager
type Client struct {
	connection *websocket.Conn
	manager    *Manager
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
	}
}

func (c *Client) readMessages() {
	defer func() {
		// clean up connection
		c.manager.removeClient(c)
	}()
	for {
		messageType, payload, err := c.connection.ReadMessage()
		if err != nil {
			// we check for abnormal connection closures (close message coming in the socket under abnormal conditions)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			// this will trigger the go routine defered at the start for clean up of client connection
			break
			// this will break the connection because the client closed the connection
		}

		log.Println(messageType)
		log.Println(string(payload))
	}
}
