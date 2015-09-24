package main

import (
	"fmt"
	"net"
	"os"

	"github.com/wkjagt/telnet-chat/client"
)

// Wire is the collection of all the channels that orchestrate
// the communiction between the clients
type Wire struct {
	broadcast          chan *client.Message
	clientConnected    chan *client.Client
	clientInitialized  chan *client.Client
	clientDisconnected chan *client.Client
}

// NewWire is a constructor for Wire
func NewWire() Wire {
	return Wire{
		make(chan *client.Message),
		make(chan *client.Client),
		make(chan *client.Client),
		make(chan *client.Client),
	}
}

var (
	wire    = NewWire()
	clients = make(map[*client.Client]*client.Client)
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Missing arguments")
		os.Exit(1)
	}

	ln, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		os.Exit(2)
	}

	fmt.Println("Listening on " + os.Args[1])

	go handleMessages()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

// client messages
func handleConnection(conn net.Conn) {
	client := client.NewClient(conn, wire.clientInitialized)
	wire.clientConnected <- client

	client.HandleMessages(wire.broadcast)

	wire.clientDisconnected <- client
}

// global messages
func handleMessages() {
	for {
		select {
		case message := <-wire.broadcast:
			for _, cl := range clients {
				if message.Client != cl {
					go func(receivingChannel chan<- *client.Message) {
						receivingChannel <- message
					}(cl.Receiving)
				}
			}
		case cl := <-wire.clientConnected:
			clients[cl] = cl
		case cl := <-wire.clientInitialized:
			cl.Outgoing <- &client.Message{Client: cl, Content: "[entered the room]"}
		case client := <-wire.clientDisconnected:
			delete(clients, client)
		}
	}
}
