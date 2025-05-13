package main

import (
	"custom_vpn/tunnel"
	"fmt"
	"flag"
	"log"
	"net"
)


// we'll fire off main, and use it to parse the port flag
// -p flag specifies what port listen on
func main(){
	portPtr := flag.Int("p", 9000, "port to listen on")
	flag.Parse()
	listenAndServe(*portPtr)
}

// Listen for incoming connections
func listenAndServe(port int){

	listener, err := net.Listen("tcp6", fmt.Sprintf(":%d", port))
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

func handleClientConn(clientConn net.Conn){
	fmt.Printf("server: recieved a conn RemoteAddr: %v\n", clientConn.RemoteAddr())
	clientConn.Write([]byte("conn established with server"))
	targetConn, err := net.Dial("tcp", "127.0.0.1:22")
	if err != nil{
		log.Fatalf("server: error while connecting to ssh on server: %v \n", err)
	}
	tunnel.CreateTunnel(targetConn, clientConn)
}