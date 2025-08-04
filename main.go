package main

import (
	"fmt"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8070")
	if err != nil {
		fmt.Println("Error listening: ", err)
		return
	}
	defer ln.Close()

	server := NewServer()
	go server.HandleBroadcasting()
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