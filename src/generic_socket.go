package main

import (
	"log"
	"time"
	"fmt"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 100 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// GenericClient handles a connection
type GenericClient struct {
	conn *websocket.Conn

	hub *Hub

	send chan []byte
}

// Conn is a getter to access connection from interface
func (g GenericClient) Conn() *websocket.Conn {
	return g.conn
}

// Send is a getter to access the send channel from interface
func (g GenericClient) Send() chan []byte {
	return g.send
}

// makeWS is a constructor for generic client
func makeWS(conn *websocket.Conn, hub *Hub) GenericClient {
	fmt.Println("CONNEddCT")
	return GenericClient{
		conn: conn,
		hub:  hub,
		send: make(chan []byte, 100),
	}
}

// SocketInterface implements the generic socket functions
type SocketInterface interface {
	Conn() *websocket.Conn
	Send() chan []byte

	closeConnection()
	sendMessage(message []byte)
}

func readMessages(i SocketInterface) {

	defer i.closeConnection()

	i.Conn().SetReadDeadline(time.Now().Add(pongWait))
	i.Conn().SetPongHandler(func(string) error { i.Conn().SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := i.Conn().ReadMessage()
	//	fmt.Println(message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		i.sendMessage(message)
	}

}

func writeMessages(i SocketInterface) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		i.closeConnection()
	}()

	for {
		select {
		case message, ok := <-i.Send():
			i.Conn().SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// channel is closed
				i.Conn().WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

		//	log.Println(message)
		//	fmt.Println(message)

			w, err := i.Conn().NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			i.Conn().SetWriteDeadline(time.Now().Add(writeWait))
			if err := i.Conn().WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		default:
		}
	}

}
