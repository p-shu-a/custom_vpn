package main

import (
	"crypto/tls"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"fmt"
	"log"
	"net"
)

// we'll fire off main, and use it to parse the port flag
// -p flag specifies what port listen on
func main(){
	go listenAndServeWithTLS(9001)
	go listenAndServeNoTLS(9000)
	select {}
}

// Listen for incoming connections
func listenAndServeWithTLS(port int){

	serverConfig, err := tlsconfig.ServerTLSConfig()
	if err != nil {
		log.Fatalf("server: error getting server config: %v", err)
	}
	listener, err := tls.Listen("tcp6", fmt.Sprintf(":%d", port), serverConfig)
	if err != nil{
		log.Fatalf("server: error while starting listener on %d:%v", port, err)
	}
	defer listener.Close()
	
	// a for without a condition is go-lang's "while". this will continue until a break, or a return
	for {
		clientConn, err := listener.Accept()
		if err != nil{
			fmt.Printf("server: unable to accept connection: %v", err)
			continue
		}
		go handleClientConn(clientConn)
	}
	
}

func listenAndServeNoTLS(port int){
	listener, err := net.Listen("tcp6", fmt.Sprintf(":%d",port))
	if err != nil{
		log.Fatalf("server: error starting listener (on-tls) on %d: %v", port, err)
	}
	defer listener.Close()

	for {
		clientConn, err := listener.Accept()
		if err != nil{
			log.Printf("server: unable to accept connection on %d: %v", port, err)
		}
		go handleClientConn(clientConn)
	}
}


func handleClientConn(clientConn net.Conn){
	
	fmt.Printf("server: recieved a conn RemoteAddr: %v\n", clientConn.RemoteAddr())
	//clientConn.Write([]byte("conn established with server"))
	targetConn, err := net.Dial("tcp", "127.0.0.1:22")
	if err != nil{
		log.Fatalf("server: error while connecting to ssh on server: %v \n", err)
	}
	tunnel.CreateTunnel(targetConn, clientConn)
}