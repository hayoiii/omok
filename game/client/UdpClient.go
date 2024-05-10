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

func HandleUdpRequest(server *UdpServer, opponent ServerMessage) {
	buffer := make([]byte, 1024)
	_, _, err := server.conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("failed to read", err)
		return
	}

	if buffer[0] != byte('/') {
		// chat
		fmt.Printf("%s> %s\n", opponent.Nickname, string(buffer))
	} else {
		// TODO: command
		fmt.Printf("%s> %s\n", opponent.Nickname, string(buffer))
	}
	server.channel <- string(buffer)
}

func RequestToServer(client *UdpClient, opponent ServerMessage) {
	ch := make(chan string)
	go sendMessage(ch)

	buffer := <-ch
	_, err := client.conn.Write([]byte(buffer))

	if err != nil {
		fmt.Print(err)
		client.channel <- "[System] Failed to send message. Try again."
		return
	}

	client.channel <- string(buffer)
}

func sendMessage(ch chan string) {
	var tmp string
	fmt.Scanln(&tmp)

	ch <- tmp
}

func StartGame(udpServer *UdpServer, opponent ServerMessage) {
	udpClient, err := CreateUdpClient(opponent)
	if err != nil {
		log.Fatal("failed to connect to opponent", err)
	}
	for {
		go RequestToServer(udpClient, opponent)
		go HandleUdpRequest(udpServer, opponent)

		select {
		case buffer := <-udpClient.channel:
			fmt.Printf("send: %s\n", string(buffer))
		case buffer := <-udpServer.channel:
			fmt.Printf("received: %s\n", string(buffer))
		}
	}
}
