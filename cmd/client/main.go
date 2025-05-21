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
	clientListenerPort := flag.Int("p", 2022, "App Port") // shitty description
	transSec := flag.Bool("tls", true, "Choose TLS (default) or basic TCP (set false)")
	flag.Parse()
	if err := startLocalListener(*clientListenerPort, *transSec); err != nil {
		log.Fatalf("Client: %v", err)
	}
}

// this function is blocking...and we want it to blocks
func startLocalListener(clientListenerPort int, transSec bool) error {

	listener, err := net.Listen("tcp6", fmt.Sprintf(":%d",clientListenerPort))

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
			if err := connectRemoteSecure(9001, conn); err != nil {
				return err
			}
		} else{
			// no TLS
			if err := connectRemoteUnsec(9000, conn); err != nil {
				return err
			}
		}
	}
}

/*
	whats the issue i'm having right now?
	- ideally the client, would only accept one incoming connection on port 2022. not multiple. multiple is the default
	- what should happen is that if the server kills the connection, you should be able to reattempt
*/

func connectRemoteSecure(serverConnPort int, conn net.Conn) error {
	
	clientConfg, err := tlsconfig.ClientTLSConfig()
	if err != nil{
		return fmt.Errorf("error fetching client config: %v",err)
	}

	// if you wonder where the "conn.close()" are, they're in the tunnel logic
	serverConn, err := tls.Dial("tcp6", fmt.Sprintf(":%d", serverConnPort), clientConfg)
	if err != nil{
		return fmt.Errorf("error dialing to server on %d: %v", serverConnPort, err)
	} else{
		log.Printf("client: secure connection to server successfully established on port :%d\n", serverConnPort)
	}

	tunnel.CreateTunnel(serverConn, conn)

	return nil
}


func connectRemoteUnsec(serverConnPort int, conn net.Conn) error {

	serverConn, err := net.Dial("tcp6", fmt.Sprintf(":%d", serverConnPort))
	if err != nil{
		return fmt.Errorf("client: error dialing to server on %d: %v", serverConnPort, err)
	} else{
		log.Printf("client: insecure connection to server successfully established on port: %d", serverConnPort)
	}

	tunnel.CreateTunnel(serverConn, conn)

	return nil
}