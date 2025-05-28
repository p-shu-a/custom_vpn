package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"custom_vpn/internal/tcp"
	"custom_vpn/internal/quic"
)

func main(){
	clientListenerPort := flag.Int("p", 2022, "Port used to connect to client (via nc, socat)")
	remoteServerAddress := flag.String("addr", "localhost", "Server IP Address")
	transSec := flag.Bool("tls", true, "Use TLS or basic TCP")
	caCertLoc := flag.String("ca", "", "specify a custom CA cert")
	flag.Parse()

	/// launch quic server and continue
	go quic.ConnectRemoteQuic(*caCertLoc, *remoteServerAddress)

	if err := startLocalListener(*clientListenerPort, *transSec, *remoteServerAddress, *caCertLoc); err != nil {
		log.Fatalf("Client: %v", err)
	}
}


func startLocalListener(clientListenerPort int, transSec bool, serverAddr string, caCertLoc string) error {


	// this listener is local to the client machine
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d",clientListenerPort))

	if err != nil{
		return fmt.Errorf("error creating listener: %v", err)
	} else{
		log.Printf("client: listener started successfully on %d", clientListenerPort)
	}
	defer listener.Close()
	
	for {
		conn, err := listener.Accept()
		if err != nil{
			_ , match := err.(net.Error) 
			if match {
				continue
			}
			return fmt.Errorf("failed to accept connection: %v", err)
		}
		log.Printf("client: recieved connection from: %v\n", conn.RemoteAddr().String())

		//no go-routine. we're only listening to a single conn
		if transSec {
			// provide TLS
			if err := tcp.ConnectRemoteSecure(9001, conn, serverAddr, caCertLoc); err != nil {
				return err
			}
		} else{
			// no TLS
			if err := tcp.ConnectRemoteUnsec(9000, conn, serverAddr); err != nil {
				return err
			}
		}

	}
}


