package main

type hub struct {
	// Registered clients
	clients map[*client]bool
	// Inbound messages
	broadcast chan *msg
	// Typing cue, emit to others except origin
	broadcastOthers chan *msg
	// Register requests
	register chan *client
	// Unregister requests
	unregister chan *client
	// Textarea content
	content string
}

func NewHub() *hub {

	var h = &hub{
		broadcast:       make(chan *msg),
		broadcastOthers: make(chan *msg),
		register:        make(chan *client),
		unregister:      make(chan *client),
		clients:         make(map[*client]bool),
		content:         "",
	}

	return h
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			for k, _ := range h.clients {
				logDebug("sending C to k.name " + k.ip)
				c.send <- []byte("C" + k.name)
			}
			h.clients[c] = true
			c.send <- []byte("M" + h.content)
			logDebug("case c := <-h.register: " + c.ip + " " + c.name)
			break

		case c := <-h.unregister:
			_, ok := h.clients[c]
			if ok {
				logDebug("case c := <-h.unregister: " + c.ip + " " + c.name)
				logVerbose(c.name + " disconnected")
				h.broadcastMessage(false, &msg{"D", c.name})
				delete(h.clients, c)
				close(c.send)
			}
			break

		case m := <-h.broadcast:
			logDebug("case m := <-h.broadcast: " + m.key + m.data)
			if m.key == "M" {
				h.content = m.data
			}
			h.broadcastMessage(true, m)
			break

		case m := <-h.broadcastOthers:
			logDebug("case m:= <-h.broadcastOthers: " + m.key + m.data)
			h.broadcastMessage(false, m)
			break
		}
	}
}

func (h *hub) broadcastMessage(emit bool, m *msg) {

	for c := range h.clients {
		if !emit {
			if c.name != m.data {
				logDebug("emitting " + m.key + m.data + " to " + c.ip + " " + c.name)
				c.send <- []byte(m.key + m.data)
			}
		} else {
			logDebug("broadcasting " + m.key + m.data + " to " + c.ip + " " + c.name)
			c.send <- []byte(m.key + m.data)	
		}

	}
}


