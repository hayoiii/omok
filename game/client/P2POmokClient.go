package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

type ServerMessage struct {
	Nickname string
	UdpAddr string
}

func readIndexFromServer(conn net.Conn) (int, error) {
	buffer := make([]byte, 16)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
		return -1, err
	}

	index, err := strconv.Atoi(string(buffer[:n]))
	if err != nil {
		log.Fatal(err)
	}

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
	log.Printf("%s joined from %s. UDP endpoint: %s\n", message.Nickname, conn.RemoteAddr().String(), message.UdpAddr)
	return *message, nil
}

func encodeMessage(message ServerMessage) []byte {
	buffer := new(bytes.Buffer)
	err := gob.NewEncoder(buffer).Encode(message)
	if err != nil {
		log.Fatal("failed to encode message ", err)
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
		fmt.Printf("%s joined (%s). You play first.", opponent.Nickname, opponent.UdpAddr)
	} else {
		fmt.Printf("%s is waiting for you (%s).\n %s plays first.", opponent.Nickname, opponent.UdpAddr, opponent.Nickname)
	}

	return index, opponent
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run P2POmokClient.go <nickname>")
	}
	nickname := os.Args[1]

	// index, opponent := 
	connectServer(nickname)

}

