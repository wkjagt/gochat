package client

import (
	"bufio"
	"net"

	"github.com/mgutz/ansi"
)

var phosphorize = ansi.ColorFunc("green+h:black")

// Client represents a client connected to the chat
type Client struct {
	Conn      net.Conn
	Receiving chan *Message
	Outgoing  chan *Message
	Name      string
	Input     *bufio.Reader
}

// Message is what is sent by a client
type Message struct {
	Client  *Client
	Content string
}

// NewClient is a constructor for Client
func NewClient(conn net.Conn, onInit chan *Client) *Client {
	client := Client{
		conn,
		make(chan *Message),
		make(chan *Message),
		"",
		bufio.NewReader(conn),
	}
	go client.Initialize(onInit)
	return &client
}

// Initialize the client
func (client *Client) Initialize(onInit chan *Client) {
	defer client.Close()

	client.SetName()
	onInit <- client

	for {
		line, err := client.Read()
		if err != nil {
			break
		}
		client.Outgoing <- &Message{client, line}
	}
}

func (client *Client) display(text string) error {
	_, err := client.Conn.Write([]byte(phosphorize(text) + "\r\n"))
	return err
}

// SetName sets the name of the client
func (client *Client) SetName() {
	client.display("What is your name?")
	name, err := client.Read()
	if err != nil {
		return
	}
	client.Name = name
}

func (client *Client) Read() (string, error) {
	in, _, err := client.Input.ReadLine()
	return string(in), err
}

// HandleMessages starts waiting for both of the clients channels
func (client *Client) HandleMessages(broadcast chan *Message) {
LOOP:
	for {
		select {
		case message, ok := <-client.Outgoing:
			if !ok {
				break LOOP
			}
			broadcast <- message
		case message := <-client.Receiving:
			err := client.display(message.Client.Name + ": " + message.Content)
			if err != nil {
				break LOOP
			}
		}
	}
}

// Close cleans up the client
func (client *Client) Close() {
	client.Conn.Close()
	client.Outgoing <- &Message{client, "[left the room]"}
	close(client.Outgoing)
}
