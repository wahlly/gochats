package main

import (
	"fmt"
	"net"
)

func main() {
	//tcp listener
	ln, err := net.Listen("tcp", ":8070")
	if err != nil {
		fmt.Println("Error listening: ", err)
		return
	}
	defer ln.Close()

	//server struct initialization
	server := NewServer()
	//handle messages sent to the message channel in another goroutine concurrently
	go server.MessageDispatcher()
	fmt.Println("Server is running on port :8070")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			return
		}
		go server.HandleConnection(conn)
	}
}