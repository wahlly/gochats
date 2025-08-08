package main

import (
	"bufio"
	"fmt"
	"maps"
	"net"
	"slices"
	"strings"
	"sync"
	"time"
)

type Message struct {
	from *net.Conn
	to *net.Conn
	content string
}

//tcp server connection struct
type Server struct{
	clients map[net.Conn]string
	msgChan chan Message
	mutex *sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		clients: make(map[net.Conn]string),
		msgChan: make(chan Message),
		mutex: &sync.RWMutex{},
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
	//add new connection to server clients map
	server.mutex.Lock()
	server.clients[conn] = name
	server.mutex.Unlock()
	server.msgChan <- Message{content: fmt.Sprintf("%s has joined the chat\n", name)}

	//timer for client connections, to detect idle/inactive users
	timeoutDuration := 5 * time.Minute	//allow inactivity for 5 minutes
	inactivityTimer := time.NewTimer(timeoutDuration)

	//goroutine to track inactivity timer
	go func() {
		<-inactivityTimer.C
		conn.Write([]byte("you have been disconnected due to inactivity\n"))
		server.mutex.Lock()
		delete(server.clients, conn)
		server.mutex.Unlock()

		server.msgChan <- Message{content: fmt.Sprintf("%s has been disconnected due to inactivity.\n", name)}
		conn.Close()
	}()

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading message: ", err)
			break
		}
		inactivityTimer.Reset(timeoutDuration)	//reset inactivity timer on every new message
		message = strings.TrimSpace(message)
		msgTimestamp := time.Now().Format("15:04")	//message timestamp

		if message == "/exit" {
			break
		}

		if strings.HasPrefix(message, "/msg "){	//private messaging
			msgParts := strings.SplitN(message, " ", 3)
			if len(msgParts) < 3{
				server.msgChan <- Message{
					to: &conn,
					content: fmt.Sprintf("[%s]: private messaging follows the format, '/msg <recipientName> <message>\n", msgTimestamp),
				}
				continue
			}

			recipientName := msgParts[1]
			msg := msgParts[2]
			recipientConn := server.getUserConn(recipientName)
			if recipientConn == nil {
				server.msgChan <- Message{
					to: &conn,
					content: fmt.Sprintf("[%s]: user %s is not available\n", msgTimestamp, recipientName),
				}
				continue
			}

			server.msgChan <- Message{
				from: &conn,
				to: &recipientConn,
				content: fmt.Sprintf("[%s] %s: %s\n", msgTimestamp, name, msg),
			}
		}else if message == "/active-users" {	//broadcast list of active users
			server.mutex.RLock()
			users := slices.Collect(maps.Values(server.clients))
			server.mutex.RUnlock()

			server.msgChan <- Message{
				from: &conn,
				content: fmt.Sprintf("[%s] active users: %s\n", msgTimestamp, strings.Join(users, ", ")),
			}
		} else{	//broadcast
			server.msgChan <- Message{
				from: &conn,
				content: fmt.Sprintf("[%s] %s: %s\n", msgTimestamp, name, message),
			}
		}
	}

	server.mutex.Lock()
	delete(server.clients, conn)
	server.mutex.Unlock()
	server.msgChan <- Message{content: fmt.Sprintf("%s has left the chat\n", name)}
	inactivityTimer.Stop()
	conn.Close()
}

//delivery of messages
func (server *Server) MessageDispatcher() {
	for msg := range server.msgChan {
		if msg.to == nil {	//has no receiver, hence a general broadcast
			server.mutex.RLock()
			for conn := range server.clients {
				conn.Write([]byte(msg.content))
			}
			server.mutex.RUnlock()
		} else{	//a private message
			conn := *msg.to	//recipient connection
			conn.Write([]byte(msg.content))
		}
		
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