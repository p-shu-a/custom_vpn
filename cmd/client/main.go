package main

import (
	"custom_vpn/tunnel"
	"flag"
	"fmt"
	"log"
	"net"
)

func main(){
	portPtr := flag.Int("p", 2022, "specify port of comms") // shitty description
	flag.Parse()
	startLocalListener(*portPtr)
}

// this function is blocking...and we want it to blocks
func startLocalListener(port int){
	listener, err := net.Listen(
		"tcp6",
		fmt.Sprintf(":%d",port))

	if err != nil{
		log.Fatalf("client: error creating listener: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil{
			fmt.Printf("client: failed to accept connection: %v \n", err)
			continue
		}
		//no go-routine. we're only listening to a single conn
		handleLocalConnection(conn)
	}
}

func handleLocalConnection(conn net.Conn){
	fmt.Printf("client: recieved conn from: %v \n", conn.RemoteAddr().String())
	//conn.Write([]byte("connection establised with client"))
	serverConn, err := net.Dial("tcp6",":9000")
	if err != nil{
		log.Fatalf("client: error dialing to server: %v", err)
	}
	tunnel.CreateTunnel(serverConn, conn)
}