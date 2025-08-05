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

type Message struct {
	from *net.Conn
	to *net.Conn
	content string
}

type Server struct{
	clients map[net.Conn]string
	msgChan chan Message
	mutex *sync.Mutex
}

func NewServer() *Server {
	return &Server{
		clients: make(map[net.Conn]string),
		msgChan: make(chan Message),
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
	server.msgChan <- Message{content: fmt.Sprintf("%s has joined the chat\n", name)}

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message: ", err)
			break
		}
		message = strings.TrimSpace(message)

		if message == "/exit" {
			break
		}

		if strings.HasPrefix(message, "/msg "){
			msgParts := strings.SplitN(message, " ", 3)
			if len(msgParts) < 3{
				server.msgChan <- Message{
					to: &conn,
					content: "private messaging follows the format, '/msg <recipientName> <message>\n",
				}
				continue
			}

			recipientName := msgParts[1]
			msg := msgParts[2]
			recipientConn := server.getUserConn(recipientName)
			if recipientConn == nil {
				server.msgChan <- Message{
					to: &conn,
					content: fmt.Sprintf("user %s is not available\n", recipientName),
				}
				continue
			}

			server.msgChan <- Message{
				from: &conn,
				to: &recipientConn,
				content: fmt.Sprintf("%s: %s\n", name, msg),
			}
		}else if message == "/active-users" {
			server.mutex.Lock()
			users := slices.Collect(maps.Values(server.clients))
			server.mutex.Unlock()
			server.msgChan <- Message{
				from: &conn,
				content: fmt.Sprintf("active users: %s\n", strings.Join(users, ", ")),
			}
		} else{
			server.msgChan <- Message{
				from: &conn,
				content: fmt.Sprintf("%s: %s\n", name, message),
			}
		}
	}

	server.mutex.Lock()
	delete(server.clients, conn)
	server.mutex.Unlock()
	server.msgChan <- Message{content: fmt.Sprintf("%s has left the chat\n", name)}
	conn.Close()
}

func (server *Server) MessageDispatcher() {
	for msg := range server.msgChan {
		server.mutex.Lock()
		if msg.to == nil {
			for conn := range server.clients {
				conn.Write([]byte(msg.content))
			}
		} else{
			conn := *msg.to	//recipient connection
			conn.Write([]byte(msg.content))
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