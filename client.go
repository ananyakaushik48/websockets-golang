package main

import (
	"encoding/json"
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

	// egress is used to avoid concurrent writes on the websocket connection
	egress chan Event // byte array channel
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
	}
}

func (c *Client) readMessages() {
	defer func() {
		// clean up connection
		c.manager.removeClient(c)
	}()
	for {
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			// we check for abnormal connection closures (close message coming in the socket under abnormal conditions)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			// this will trigger the go routine defered at the start for clean up of client connection
			break
			// this will break the connection because the client closed the connection
		}
		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error mashalling event :%v", err)
			break
		}
		//
		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("error handling message: ", err)
		}
	}
}

func (c *Client) writeMessages() {
	defer func() {
		// cleanup go routine
		c.manager.removeClient(c)
	}()
	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				// this means were having problems with the client
				// so it write the close message to the client with error message
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				// breaks the for loop and triggers the cleanup
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return
			}
			// This is if we were successful and can receive a message
			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("failed to send message: %v", err)
			}
			log.Println("message sent")
		}
	}
}
