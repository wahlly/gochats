package main

import (
	"bufio"
	"fmt"
	"maps"
	"net"
	"slices"
	"strings"
	"sync"
)

var (
	clients = make(map[net.Conn]string)
	broadcast = make(chan string)
	mutex = &sync.Mutex{}
)

func main() {
	server, err := net.Listen("tcp", ":8070")
	if err != nil {
		fmt.Println("Error listening: ", err)
		return
	}
	defer server.Close()
	go handleBroadcasting()
	fmt.Println("Server is running on port :8070")

	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			return
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	conn.Write([]byte("Enter your name: "))
	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading name: ", err)
		return
	}

	name = strings.TrimSpace(name)
	mutex.Lock()
	clients[conn] = name
	mutex.Unlock()
	broadcast <- fmt.Sprintf("%s has joined the chat\n", name)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message: ", err)
			break
		}
		message = strings.TrimSpace(message)

		if message == "/exit" {
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()

			broadcast <- fmt.Sprintf("%s has left chat", name)
			conn.Close()
			return
		}

		if message == "/active-users" {
			users := slices.Collect(maps.Values(clients))
			broadcast <- fmt.Sprintf("active users: %s\n", strings.Join(users, ", "))
		} else{
			broadcast <- fmt.Sprintf("%s: %s\n", name, message)
		}
	}

	mutex.Lock()
	delete(clients, conn)
	mutex.Unlock()
	broadcast <- fmt.Sprintf("%s has left the chat\n", name)
	conn.Close()
}

func handleBroadcasting() {
	for message := range broadcast {
		mutex.Lock()
		for conn := range clients {
			conn.Write([]byte(message))
		}
		mutex.Unlock()
	}
}