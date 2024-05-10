package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type ServerMessage struct {
	Nickname string
	UdpAddr  string
}

func notifyToServer(conn net.Conn, nickname string) error {
	// 서버에게 클라이언트의 닉네임과 UDP 주소를 전달한다.
	message := ServerMessage{nickname, CONN_ADDR}
	encodedMessage := encodeMessage(message)
	_, err := conn.Write(encodedMessage)
	if err != nil {
		return err
	}
	return nil
}

func readIndexFromServer(ch chan int, eCh chan error, conn net.Conn) {
	buffer := make([]byte, 16)
	_, err := conn.Read(buffer)
	if err != nil {
		eCh <- err
		return
	}

	index := int(buffer[0])
	ch <- index
}

func readMessageFromServer(ch chan ServerMessage, eCh chan error, conn net.Conn) {
	for {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err == io.EOF {
			continue
		}
		if err != nil {
			eCh <- err
			return
		}

		reader := bytes.NewBuffer(buffer)
		message := new(ServerMessage)

		err = gob.NewDecoder(reader).Decode(message)
		if err != nil {
			eCh <- err
			return
		}

		ch <- *message
	}
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
	CONN_ADDR        = "localhost:12345"
)

func connectServer(nickname string) net.Conn {
	conn, err := net.Dial(SERVER_CONN_TYPE, SERVER_CONN_ADDR)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Welcome to P2P Omok, %s! Server: %s\n", nickname, SERVER_CONN_ADDR)
	return conn
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run P2POmokClient.go <nickname>")
	}
	nickname := os.Args[1]

	// index, opponent :=
	conn := connectServer(nickname)
	defer conn.Close()

	err := notifyToServer(conn, nickname)
	if err != nil {
		log.Fatal("notifyToServer: ", err)
	}

	var index int
	var opponent ServerMessage

	indexCh := make(chan int)
	eCh := make(chan error)
	go func() {
		readIndexFromServer(indexCh, eCh, conn)
	}()

	msgCh := make(chan ServerMessage)

	for {
		select {
		case index = <-indexCh:
			if index == 0 {
				fmt.Println("Waiting for an opponent...")
			}

			go func() {
				readMessageFromServer(msgCh, eCh, conn)
			}()

		case opponent = <-msgCh:
			if index == 0 {
				fmt.Printf("%s joined (%s). You play first.", opponent.Nickname, opponent.UdpAddr)
			} else {
				fmt.Printf("%s is waiting for you (%s).\n %s plays first.", opponent.Nickname, opponent.UdpAddr, opponent.Nickname)
			}
		case err := <-eCh:
			log.Fatal(err)
		}
	}
}
