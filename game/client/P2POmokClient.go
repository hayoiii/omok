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
	udpEndpoint string
}

func handleConnection(conn *net.UDPConn) {
	buffer := make([]byte, 1024)

	// udp connection으로 부터 값을 읽어들인다. 
	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		log.Fatal(err)
	}
    
    // 리턴 값은 전달 받은 클라이언트 서버의 address, msg
	fmt.Println("UDP client : ", addr)
	fmt.Println("Received from UDP client : ", string(buffer[:n]))

	// 클라이언트로 msg write
	msg := []byte("Hello UDP Client")
	_, err = conn.WriteToUDP(msg, addr)
	if err != nil {
		log.Fatal(err)
	}
}

func processServerMessage(conn net.Conn) (ServerMessage, error) {
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
	log.Printf("%s joined from %s. UDP endpoint: %s\n", message.nickname, conn.RemoteAddr().String(), message.udpEndpoint)
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
func connectServer(nickname string) {
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

	// 서버로부터 메시지를 읽어들인다.
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	// 몇 번째 클라이언트인지 확인
	index := int(buffer[0])

	if(index == 0) {
		fmt.Println("Waiting for an opponent...")
	}

	_, err = conn.Read(buffer[n+1:])
	if err != nil {
		log.Fatal(err)
	}

	opponent, err := processServerMessage(conn)
	if err != nil {
		log.Fatal(err)
	}

	if index == 0 {
		fmt.Printf("%s joined (%s). You play first.", opponent.nickname, opponent.udpEndpoint)
	} else {
		fmt.Printf("%s is waiting for you (%s).\n %s plays first.", opponent.nickname, opponent.udpEndpoint, opponent.nickname)
	}
}

func main() {
	nickname := os.Args[1]
	if len(nickname) == 0 {
		log.Fatal("Usage: go run P2POmokClient.go <nickname>")
	}

	connectServer(nickname)
}

