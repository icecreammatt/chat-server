package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

var clients []Client
var clientId int

type Client struct {
	id   int
	conn net.Conn
}

type Message struct {
	ClientId int    `json:"clientId"`
	Message  string `json:"message"`
}

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
		clients = append(clients, client)
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
	fmt.Println("closing client %d", client.id)

	var index int
	for i, c := range clients {
		if c.id == client.id {
			index = i
			break
		}
	}
	clients = clients[:index+copy(clients[index:], clients[index+1:])]
	fmt.Println("updated clients list")
}

func broadcastMessage(buf [512]byte, length int, clientId int) {
	msg := fmt.Sprintf("%s", buf[0:length])
	message := Message{clientId, msg}
	jsonResponse, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error", err.Error())
	}
	for _, client := range clients {
		if client.conn != nil && client.id != clientId {
			fmt.Printf("Writing to client %d\n", client.id)
			client.conn.Write(jsonResponse)
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
