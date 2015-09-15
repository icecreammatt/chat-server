package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
)

type Client struct {
	id   int
	conn net.Conn
}

type Message struct {
	ClientId int    `json:"clientId"`
	Message  string `json:"message"`
}

type ClientList struct {
	mutex   sync.Mutex
	clients []Client
}

func (c *ClientList) Add(client *Client) {
	c.mutex.Lock()
	c.clients = append(c.clients, *client)
	c.mutex.Unlock()
}

func (c *ClientList) Remove(clientId int) {
	c.mutex.Lock()
	var index int
	for i, client := range c.clients {
		if client.id == clientId {
			index = i
			break
		}
	}
	fmt.Printf("closing client %d\n", clientId)
	c.clients = c.clients[:index+copy(c.clients[index:], c.clients[index+1:])]
	fmt.Println("updated clients list")
	c.mutex.Unlock()
}

func (c *ClientList) Broadcast(message []byte) {
	c.mutex.Lock()
	for _, client := range c.clients {
		if client.conn != nil {
			fmt.Printf("Writing to client %d\n", client.id)
			client.conn.Write(message)
		}
	}
	c.mutex.Unlock()
}

var clientList ClientList
var clientId int

func main() {

	service := ":1201"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	clientId = 0

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		client := Client{clientId, conn}
		clientList.Add(&client)
		clientId++

		// run as a goroutine
		go handleClient(client)
	}
}

func handleClient(client Client) {
	// close connection on exit
	fmt.Printf("New client joined %d\n", client.id)
	defer client.conn.Close()

	var buf [512]byte
	for {
		// read upto 512 bytes
		n, err := client.conn.Read(buf[0:])
		if err != nil {
			break
		}

		// write the n bytes read
		broadcastMessage(buf, n, client.id)
	}
	clientList.Remove(client.id)
}

func broadcastMessage(buf [512]byte, length int, clientId int) {
	msg := fmt.Sprintf("%s", buf[0:length])
	message := Message{clientId, msg}
	jsonResponse, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	clientList.Broadcast(jsonResponse)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}
