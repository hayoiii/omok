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

func encodeMessage(message ServerMessage) []byte {
	buffer := new(bytes.Buffer)
	err := gob.NewEncoder(buffer).Encode(message)
	if err != nil {
		log.Fatal("failed to encode message ", err)
	}
	return buffer.Bytes()
}

func connectServer(nickname string) net.Conn {
	conn, err := net.Dial(SERVER_CONN_TYPE, SERVER_CONN_ADDR)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Welcome to P2P Omok, %s! Server: %s\n", nickname, SERVER_CONN_ADDR)
	return conn
}

func notifyToServer(conn net.Conn, nickname string, udpAddr string) error {
	// 서버에게 클라이언트의 닉네임과 UDP 주소를 전달한다.
	message := ServerMessage{nickname, udpAddr}
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

const (
	SERVER_CONN_TYPE = "tcp"
	SERVER_CONN_ADDR = "localhost:5999"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go run P2POmokClient.go <nickname>")
	}
	nickname := os.Args[1]

	udpServer, err := CreateUdpServer()
	if err != nil {
		log.Fatal(err)
	}

	conn := connectServer(nickname)
	defer conn.Close()

	err = notifyToServer(conn, nickname, udpServer.conn.LocalAddr().String())
	if err != nil {
		log.Fatal("notifyToServer: ", err)
	}

	var index int
	var opponent ServerMessage

	indexCh := make(chan int)
	eCh := make(chan error)
	go readIndexFromServer(indexCh, eCh, conn)

	msgCh := make(chan ServerMessage)

	for {
		select {
		case index = <-indexCh:
			if index == 0 {
				fmt.Println("Waiting for an opponent...")
			}
			go readMessageFromServer(msgCh, eCh, conn)

		case opponent = <-msgCh:
			if index == 0 {
				fmt.Printf("%s joined (%s). You play first.\n\n", opponent.Nickname, opponent.UdpAddr)
			} else {
				fmt.Printf("%s is waiting for you (%s).\n %s plays first.\n\n", opponent.Nickname, opponent.UdpAddr, opponent.Nickname)
			}
			StartGame(udpServer, opponent)
			return
		case err := <-eCh:
			log.Fatal(err)
		}
	}
}
