# Go TCP Chat Server

A tcp based chat server, supporting concurrent connection management using goroutines, implemented in Go, using raw TCP connections. Supports multiple clients, broadcasting messages, private messaging, active users list, and an inactivity timeout to detect idle connections.

## Features

- **Raw TCP-based communication**
- **Concurrent connections** handled with goroutines
- **Broadcast messaging** to all connected clients
- **Private messaging** via `/msg <username> <message>`
- **List active users** via `/active-users`
- **Inactivity timeout** — clients are disconnected after 5 minutes of no activity
- **Thread-safe client management** using `sync.Mutex`

## How It Works

- **Main server loop** listens for new TCP connections.
- **Each client** is handled in its own goroutine.
- **Broadcast channel** is used for sending messages to all clients.
- **Private messages** are sent directly to a client’s TCP connection.

## Commands

| Command                          | Description                              |
|-----------------------------------|------------------------------------------|
| `/active-users`                   | Lists all currently connected usernames |
| `/msg <username> <message>`       | Sends a private message to a user       |

## Installation

```bash
git clone https://github.com/<your-username>/<repo-name>.git
cd <repo-name>
go mod tidy
go run main.go
```
using separate terminals, clients can connect to the opened connection.
```bash
nc localhost 8070
```
8070 represent the port number. "nc localhost 8070" represent connection for each client intending to connect.

On a successful connection, you will be prompted to provide a name to use for messaging.
