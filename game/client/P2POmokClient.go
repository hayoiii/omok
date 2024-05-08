package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
)

type ServerMessage struct {
	nickname string
	udpAddr string
}

func handleConnection(conn *net.UDPConn) {
	buffer := make([]byte, 1024)

	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		log.Fatal(err)
	}
    
	fmt.Println("UDP client : ", addr)
	fmt.Println("Received from UDP client : ", string(buffer[:n]))

	msg := []byte("Hello UDP Client")
	_, err = conn.WriteToUDP(msg, addr)
	if err != nil {
		log.Fatal(err)
	}
}

func readIndexFromServer(conn net.Conn) (int, error) {
	buffer := make([]byte, 16)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
		return -1, err
	}

	index := int(buffer[n])
	return index, nil
}

func readMessageFromServer(conn net.Conn) (ServerMessage, error) {
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
	log.Printf("%s joined from %s. UDP endpoint: %s\n", message.nickname, conn.RemoteAddr().String(), message.udpAddr)
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

const (
	SERVER_CONN_TYPE = "tcp"
	SERVER_CONN_ADDR = "localhost:5999"
	CONN_ADDR = "localhost:12345"
)
func connectServer(nickname string) (int, ServerMessage) {
	conn, err := net.Dial(SERVER_CONN_TYPE, SERVER_CONN_ADDR)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Printf("Welcome to P2P Omok, %s! Server: %s\n", nickname, SERVER_CONN_ADDR)

	// 서버에게 클라이언트의 닉네임과 UDP 주소를 전달한다.
	message:= ServerMessage{nickname, CONN_ADDR}
	encodedMessage := encodeMessage(message)
	_, err = conn.Write(encodedMessage)
	if err != nil {
		log.Fatal(err)
	}

	index, err := readIndexFromServer(conn)
	if err != nil {
		log.Fatal(err)
	}
	if(index == 0) {
		fmt.Println("Waiting for an opponent...")
	}

	opponent, err := readMessageFromServer(conn)
	if err != nil {
		log.Fatal(err)
	}

	if index == 0 {
		fmt.Printf("%s joined (%s). You play first.", opponent.nickname, opponent.udpAddr)
	} else {
		fmt.Printf("%s is waiting for you (%s).\n %s plays first.", opponent.nickname, opponent.udpAddr, opponent.nickname)
	}

	return index, opponent
}

func main() {
	nickname := os.Args[1]
	if len(nickname) == 0 {
		log.Fatal("Usage: go run P2POmokClient.go <nickname>")
	}

	// index, opponent := 
	connectServer(nickname)

}

