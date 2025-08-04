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

type Server struct{
	clients map[net.Conn]string
	broadcast chan string
	mutex *sync.Mutex
}

func NewServer() *Server {
	return &Server{
		clients: make(map[net.Conn]string),
		broadcast: make(chan string),
		mutex: &sync.Mutex{},
	}
}



func (server *Server) HandleConnection(conn net.Conn) {
	conn.Write([]byte("Enter your name: "))
	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading name: ", err)
		return
	}

	name = strings.TrimSpace(name)
	server.mutex.Lock()
	server.clients[conn] = name
	server.mutex.Unlock()
	server.broadcast <- fmt.Sprintf("%s has joined the chat\n", name)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message: ", err)
			break
		}
		message = strings.TrimSpace(message)

		if message == "/exit" {
			server.mutex.Lock()
			delete(server.clients, conn)
			server.mutex.Unlock()

			server.broadcast <- fmt.Sprintf("%s has left chat", name)
			conn.Close()
			return
		}

		if strings.HasPrefix(message, "/msg "){
			msgParts := strings.SplitN(message, " ", 3)
			if len(msgParts) < 3{
				fmt.Println("private messaging follows the format, '/msg <recipientName> <message>")
				break
			}

			recipientName := msgParts[1]
			msg := msgParts[2]
			recipientConn := server.getUserConn(recipientName)
			if recipientConn == nil {
				conn.Write([]byte("user is not available"))
				break
			}

			recipientConn.Write([]byte(fmt.Sprintf("%s\n", msg)))
			conn.Write([]byte(fmt.Sprintf("message sent to %s successfully\n", recipientName)))
		}else if message == "/active-users" {
			server.mutex.Lock()
			users := slices.Collect(maps.Values(server.clients))
			server.mutex.Unlock()
			server.broadcast <- fmt.Sprintf("active users: %s\n", strings.Join(users, ", "))
		} else{
			server.broadcast <- fmt.Sprintf("%s: %s\n", name, message)
		}
	}

	server.mutex.Lock()
	delete(server.clients, conn)
	server.mutex.Unlock()
	server.broadcast <- fmt.Sprintf("%s has left the chat\n", name)
	conn.Close()
}

func (server *Server) HandleBroadcasting() {
	for message := range server.broadcast {
		server.mutex.Lock()
		for conn := range server.clients {
			conn.Write([]byte(message))
		}
		server.mutex.Unlock()
	}
}

func (server *Server) getUserConn(name string) net.Conn {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for conn, username := range server.clients {
		if username == name {
			return conn
		}
	}

	return nil
}