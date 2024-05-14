// TODO: wirte 용, read 용 모두 connection 필요
package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	go_getport "github.com/jsumners/go-getport"
)

type UdpServer struct {
	conn    *net.UDPConn
	channel chan string
}

type UdpClient struct {
	conn    *net.UDPConn
	channel chan string
}

func GetUdpAddr() (*net.UDPAddr, error) {
	portResult, err := go_getport.GetUdp4PortForAddress("localhost")
	if err != nil {
		return nil, err
	}
	udpAddr := portResult.IP + ":" + strconv.Itoa(portResult.Port)
	resolvedUdpAddr, err := net.ResolveUDPAddr("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	return resolvedUdpAddr, nil
}

func CreateUdpClient(opponent ServerMessage) (*UdpClient, error) {
	localAddr, err := GetUdpAddr()
	if err != nil {
		return nil, err
	}

	resolvedUdpAddr, err := net.ResolveUDPAddr("udp", opponent.UdpAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", localAddr, resolvedUdpAddr)
	if err != nil {
		return nil, err
	}
	channel := make(chan string)
	return &UdpClient{
		conn,
		channel,
	}, nil
}

func CreateUdpServer() (*UdpServer, error) {
	addr, err := GetUdpAddr()
	if err != nil {
		log.Fatal(err)
	}

	udpConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return &UdpServer{}, err
	}

	channel := make(chan string)
	return &UdpServer{
		udpConn,
		channel,
	}, nil
}

func HandleUdpRequest(server *UdpServer, opponent ServerMessage, isMyTurn bool) {
	buffer := make([]byte, 1024)
	_, _, err := server.conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("failed to read", err)
		return
	}

	if buffer[0] != byte('/') {
		// chat
		fmt.Printf("%s> %s\n", opponent.Nickname, string(buffer))
		server.channel <- string("")
	} else {
		command := string(buffer)[1:]
		if command == "gg" {
			// 상대방 항복
			fmt.Println("You are winner!")
			// server.channel <- gg
		}
		if command == "x y" {
			// printBoard, checkWin
			fmt.Println("Your turn")
			// server.channel <- checkWin
		}
	}
}

func RequestToServer(client *UdpClient, opponent ServerMessage, isMyTurn bool) {
	ch := make(chan string)
	go sendMessage(ch)

	input := <-ch
	_, err := client.conn.Write([]byte(input))

	if err != nil {
		fmt.Print(err)
		client.channel <- "[System] Failed to send message. Try again."
		return
	}

	if input == "/gg" {
		// 항복
		fmt.Println("You are loser!")
	}
	if input == "//1 2" {
		// printBoard, checkWin
		fmt.Printf("%s's turn\n", opponent.Nickname)
		// server.channel <- checkWin
	}
}

func sendMessage(ch chan string) {
	var tmp string
	fmt.Scanln(&tmp)

	ch <- tmp
}

func StartGame(udpServer *UdpServer, opponent ServerMessage, index int) {
	udpClient, err := CreateUdpClient(opponent)
	if err != nil {
		log.Fatal("failed to connect to opponent", err)
	}
	defer func() {
		udpServer.conn.Close()
		udpClient.conn.Close()
	}()

	turn := 0
	for {
		go RequestToServer(udpClient, opponent, turn == index)
		go HandleUdpRequest(udpServer, opponent, turn == index)

		select {
		case buffer := <-udpClient.channel:
			if buffer == "내가 움직였으면" {
				// 턴 넘기기
				turn = (turn + 1) % 2
			}
			if buffer == "내가 이겼으면" {
				fmt.Println("You are winner!")
			}
			if buffer == "항복" {
				fmt.Println("You are loser!")
			}
		case buffer := <-udpServer.channel:
			if buffer == "상대방이 움직였으면" {
				// 턴 넘기기
				turn = (turn + 1) % 2
			}
			if buffer == "상대방이 이겼으면" {
				fmt.Println("You are loser!")
			}
			if buffer == "항복" {
				fmt.Println("You are winner!")
			}
		}
	}
}
