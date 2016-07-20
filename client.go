package main

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 1024
)

type client struct {
	ip   string
	ws   *websocket.Conn
	send chan []byte
	name string
	h    *hub
}

type msg struct {
	key  string
	data string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  maxMessageSize,
	WriteBufferSize: maxMessageSize,
}

func ServeWs(w http.ResponseWriter, r *http.Request, h *hub) {
	logDebug(r.RemoteAddr + " requested upgrade to websocket")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c := &client{
		ip:   r.RemoteAddr,
		send: make(chan []byte, maxMessageSize),
		ws:   ws,
		h:    h,
	}

	h.register <- c

	go c.writeLoop()
	go c.readLoop()
}

func (c *client) readLoop() {
	defer func() {
		logDebug(c.ip + " disconnected")
		c.h.unregister <- c
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}

		logDebug("got message: " + string(message) + " from: " + c.ip)
		parsedMSg, err := c.parse(message)
		if err != nil {
			log.Println(err)
			continue
		}
		logDebug("parsed it to: " + parsedMSg.key + " " + parsedMSg.data)

		switch parsedMSg.key {
		case "M":
			c.h.broadcast <- parsedMSg
		case "C":
			c.name = parsedMSg.data
			c.h.broadcastOthers <- parsedMSg
		case "W":
			c.h.broadcastOthers <- parsedMSg
		}
	}

}

func (c *client) writeLoop() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *client) parse(m []byte) (*msg, error) {

	if len(m) == 0 {
		return nil, errors.New("ws msg is empty")
	}

	i := string(m[0])
	/* WebSocket text messages are parsed, 
	   first we look for the prefix code, 
	   then if any content follows        */

	switch i {
	// Client is typing
	case "W":
		logVerbose(c.name + " is typing")
		return &msg{"W", c.name}, nil
	// Sending the current state of client's textarea
	case "M":
		logVerbose(c.name + " typing timeout, text area refreshed")
		logDebug(c.name + " typing timeout, text area refreshed with: " + string(m[1:]))
		return &msg{"M", string(m[1:])}, nil
	// Client connected with chosen username
	case "C":
		c.name = string(m[1:])
		if len(m[1:]) > 0 {
			logVerbose(string(m[1:]) + " connected")
			return &msg{"C", string(m[1:])}, nil
		} else {
			return nil, errors.New(c.ip + " did not provide a username")
		}

	default:
		return nil, errors.New("wrong prefix code: " + i + " from: " + c.ip)
	}
}

func (c *client) write(mt int, message []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, message)
}
