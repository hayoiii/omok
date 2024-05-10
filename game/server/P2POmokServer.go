package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"net"
	"sync"
)

type ServerMessage struct {
	Nickname string
	UdpAddr string
}
type Client struct {
	conn net.Conn
	message ServerMessage
}

var clients [2]Client
var mutex = &sync.Mutex{}

func processRequest(conn net.Conn) (ServerMessage, error) {
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal("failed to read", err)
		return ServerMessage{}, err
	}

	reader := bytes.NewBuffer(buffer)
	message := new(ServerMessage)

	err = gob.NewDecoder(reader).Decode(message)
	if err != nil {
		return ServerMessage{}, err
	}
	log.Printf("%s joined from %s. UDP endpoint: %s\n", message.Nickname, conn.RemoteAddr().String(), message.UdpAddr)
	return *message, nil
}

func encodeMessage(message ServerMessage) []byte {
	buffer := new(bytes.Buffer)
	err := gob.NewEncoder(buffer).Encode(message)
	if err != nil {
		log.Fatal("failed to encode message", err)
	}
	return buffer.Bytes()
}

func handleRequest(conn net.Conn) {
	var index int
	if clients[0].conn == nil {
		index = 0
	} else {
		index = 1
	}
	
	message, err := processRequest(conn)
	if err != nil {
		log.Fatal("failed to process request", err)
	}

	clients[index].conn = conn
	clients[index].message = message

	indexBytes := []byte{byte(index)}
	conn.Write(indexBytes)
	
	if index == 0 {
		log.Println("1 user connected, waiting for another user to join...")
	} else {
		log.Printf("2 users connected, notifying %s and %s\n", clients[0].message.Nickname, clients[1].message.Nickname)
		message0 := encodeMessage(clients[0].message)
		message1 := encodeMessage(clients[1].message)
		
		clients[0].conn.Write(message1)
		clients[1].conn.Write(message0)

		clients[0].conn.Close()
		clients[1].conn.Close()

		clients = [2]Client{}
	}
}

func waitForConnections(listener net.Listener) {
	var wg sync.WaitGroup
	wg.Add(2) // 2개의 연결을 기다립니다.

	for i := 0; i < 2; i++ {
			go func() {
					conn, err := listener.Accept()
					if err != nil {
							log.Fatal(err)
					}
					defer conn.Close()

					mutex.Lock()
					handleRequest(conn)
					mutex.Unlock()
					wg.Done() // 연결 처리 완료
			}()
	}

	wg.Wait()
}


const (
	CONN_TYPE = "tcp"
	CONN_ADDR = "localhost:5999"
)

func main() {
	// tcp server
	l, err := net.Listen(CONN_TYPE, CONN_ADDR)
	if err != nil {
		log.Fatal("listen failed:", err)
	}

	defer l.Close()
	waitForConnections(l)
}
