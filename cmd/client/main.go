package main

import (
	"crypto/tls"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"flag"
	"fmt"
	"log"
	"net"
)

func main(){
	port:= flag.Int("p", 2022, "App Port") // shitty description
	transSec := flag.Bool("tls", true, "Choose TLS (default) or basic TCP (set false)")
	flag.Parse()
	startLocalListener(*port, *transSec)
}

// this function is blocking...and we want it to blocks
func startLocalListener(port int, transSec bool){
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
		if transSec {
			connectRemoteSecure(conn)		// provide TLS
		} else{
			connectRemoteUnsec(conn)		// no TLS
		}
		
	}
}

func connectRemoteSecure(conn net.Conn){
	port := 9001
	// if you wonder where the defer conn.close() are, they're in the tunnel logics
	fmt.Printf("client: recieved conn from: %v \n", conn.RemoteAddr().String())
	clientConfg, err := tlsconfig.ClientTLSConfig()
	if err != nil{
		log.Fatalf("client: error fetching client config: %v",err)
	}else{
		log.Printf("got client config. dialing remote on %d...\n", port)
	}
	serverConn, err := tls.Dial("tcp6", fmt.Sprintf(":%d", port), clientConfg)
	if err != nil{
		log.Fatalf("client: error dialing to server on %d: %v", port, err)
	}
	tunnel.CreateTunnel(serverConn, conn)
}


func connectRemoteUnsec(conn net.Conn){
	port := 9000
	fmt.Printf("client: recieved conn from: %v \n", conn.RemoteAddr().String())
	serverConn, err := net.Dial("tcp6", fmt.Sprintf(":%d", port))
	if err != nil{
		log.Fatalf("client: error dialing to server on %d: %v", port, err)
	}
	tunnel.CreateTunnel(serverConn, conn)
}