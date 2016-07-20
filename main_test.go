package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

const human_interval = 50 * time.Millisecond

func expect(t *testing.T, a interface{}, b interface{}) {
	t.Log("Expected", b, "got", a)
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

type test_client struct{ Conn *websocket.Conn }

func newWSClient(t *testing.T, url string) *test_client {
	// replace default http with websocket scheme
	ws_addr := "ws" + url[4:]
	ws_conn, r, err := websocket.DefaultDialer.Dial(ws_addr, nil)

	t.Log("New ws client connection to " + ws_addr)

	if err != nil {
		t.Error("dial", err)
	}
	if r.StatusCode != 101 {
		t.Error("expected 101 switching protocol status code, got", r.StatusCode)
	}

	return &test_client{ws_conn}
}

func newTestServer(h *hub) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(w, r, h)
	}))

	return server
}

func (c *test_client) write(m string) {
	c.Conn.WriteMessage(websocket.TextMessage, []byte(m))
}

func (c *test_client) readLoop(t *testing.T, r chan<- string) {

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			t.Log("client.ReadMessage err", err)
			break
		}
		t.Log("WS Client Readloop received:", string(message))

		r <- string(message)
	}

}

type testStruct struct {
	t      *testing.T
	r      chan string
	done   chan struct{}
	h      *hub
	server *httptest.Server
	c      *test_client
}

func newTestStruct(t *testing.T, expected_msgs []string) *testStruct {

	tS := &testStruct{
		t:    t,
		r:    make(chan string),
		done: make(chan struct{}),
		h:    NewHub(),
	}

	tS.server = newTestServer(tS.h)
	tS.c = newWSClient(t, tS.server.URL)
	go tS.h.run()
	go tS.c.readLoop(t, tS.r)
	go tS.testSlice(expected_msgs)

	return tS

}

func (tS *testStruct) testSlice(expected_msgs []string) {

	count := 0
	for {
		select {
		case rt := <-tS.r:
			expect(tS.t, rt, expected_msgs[count])

			if count == len(expected_msgs)-1 {
				tS.t.Log("Looped through all expected_msgs")
				close(tS.done)
				return
			}
			count++

		case <-time.After(5 * time.Second):
			tS.t.Error("testSlice timeout, probably too many received messages")
			close(tS.done)
			return
		}
	}

}

func TestOneConnection(t *testing.T) {
	expected_msgs := []string{"M"}
	tS := newTestStruct(t, expected_msgs)

	<-tS.done
	tS.c.Conn.Close()
}

func TestTwoConnections(t *testing.T) {
	expected_msgs := []string{"M", "Cnewuser"}
	tS := newTestStruct(t, expected_msgs)

	time.Sleep(human_interval)
	
	c2 := newWSClient(t, tS.server.URL)
	c2.write("Cnewuser")

	time.Sleep(human_interval)

	if len(tS.h.clients) != 2 {
		tS.t.Errorf("expected 2 clients in hub map, got %d", len(tS.h.clients))
	}

	time.Sleep(human_interval)
	c2.Conn.Close()
	time.Sleep(human_interval)

	if len(tS.h.clients) != 1 {
		tS.t.Errorf("expected 1 clients in hub map, got %d", len(tS.h.clients))
	}

	<-tS.done
	tS.c.Conn.Close()

}

func TestMultipleEvents(t *testing.T) {

	expected_msgs := []string{"M", "Cnewuser2",
		"Dnewuser2", "Cnewuser3",
		"Wnewuser3", "Mhello", "Dnewuser3"}

	tS := newTestStruct(t, expected_msgs)

	c2 := newWSClient(t, tS.server.URL)

	time.Sleep(human_interval)

	c2.write("Cnewuser2")

	time.Sleep(human_interval)

	if len(tS.h.clients) != 2 {
		tS.t.Errorf("expected 2 clients in hub map, got %d", len(tS.h.clients))
	}

	c2.Conn.Close()
	
	time.Sleep(human_interval)

	if len(tS.h.clients) != 1 {
		tS.t.Errorf("expected 2 clients in hub map, got %d", len(tS.h.clients))
	}

	c3 := newWSClient(t, tS.server.URL)
	c3.write("Cnewuser3")
	
	time.Sleep(human_interval)
	
	c3.write("W")
	
	time.Sleep(human_interval)
	
	c3.write("Mhello")

	if len(tS.h.clients) != 2 {
		tS.t.Errorf("expected 2 clients in hub map, got %d", len(tS.h.clients))
	}

	time.Sleep(human_interval)
	
	c3.Conn.Close()

	<-tS.done
	tS.c.Conn.Close()

}

func TestConcurrentMessages(t *testing.T) {

	expected_msgs := []string{"M", "Cnewuser2", "Wnewuser2",
		"MAandB", "MAandC", "Wnewuser2",
		"Mblabla", "Dnewuser2"}

	tS := newTestStruct(t, expected_msgs)

	time.Sleep(human_interval)

	c2 := newWSClient(t, tS.server.URL)
	c2.write("Cnewuser2")

	time.Sleep(human_interval)

	c2.write("W")
	tS.c.write("W")

	time.Sleep(human_interval)

	// both clients writing to server
	c2.write("MAandB")

	time.Sleep(human_interval)
	
	tS.c.write("MAandC")

	time.Sleep(human_interval)

	// check if last state of textarea is saved in hub (without prefix "M")
	expect(t, tS.h.content, "AandC")

	time.Sleep(human_interval)

	c2.write("W")

	time.Sleep(human_interval)

	c2.write("Mblabla")

	time.Sleep(human_interval)

	expect(t, tS.h.content, "blabla")

	time.Sleep(human_interval)
	
	c2.Conn.Close()

	<-tS.done
	tS.c.Conn.Close()
}
