package main

import (
	"context"
	"crypto/tls"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/quic-go/quic-go"
)

func main(){
	clientListenerPort := flag.Int("p", 2022, "Port used to connect to client (via nc, socat)")
	remoteServerAddress := flag.String("addr", "localhost", "Server IP Address")
	transSec := flag.Bool("tls", true, "Use TLS or basic TCP")
	caCertLoc := flag.String("ca", "", "specify a custom CA cert")
	flag.Parse()
	
	if err := startLocalListener(*clientListenerPort, *transSec, *remoteServerAddress, *caCertLoc); err != nil {
		log.Fatalf("Client: %v", err)
	}
}

// this function is blocking...and we want it to blocks
func startLocalListener(clientListenerPort int, transSec bool, serverAddr string, caCertLoc string) error {

	// this listener is local to the client machine
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d",clientListenerPort))

	if err != nil{
		return fmt.Errorf("error creating listener: %v", err)
	} else{
		log.Printf("client: listener started successfully on %d", clientListenerPort)
	}
	defer listener.Close()
	go connectRemoteQuic(caCertLoc)
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
			if err := connectRemoteSecure(9001, conn, serverAddr, caCertLoc); err != nil {
				return err
			}
		} else{
			// no TLS
			if err := connectRemoteUnsec(9000, conn, serverAddr); err != nil {
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

func connectRemoteSecure(serverConnPort int, conn net.Conn, serverAddr string, caCertLoc string) error {
	
	clientConfg, err := tlsconfig.ClientTLSConfig(caCertLoc)
	if err != nil{
		return fmt.Errorf("error fetching client config: %v",err)
	}

	// if you wonder where the "conn.close()" are, they're in the tunnel logic
	serverConn, err := tls.Dial("tcp",
								fmt.Sprintf("%v:%d", serverAddr, serverConnPort), 
								clientConfg)
	if err != nil{
		return fmt.Errorf("error dialing to server (%v:%d): %v", serverAddr, serverConnPort, err)
	} else{
		log.Printf("client: secure connection to server successfully established %v:%d\n", serverAddr, serverConnPort)
	}

	tunnel.CreateTunnel(serverConn, conn)

	return nil
}


func connectRemoteUnsec(serverConnPort int, conn net.Conn, serverAddr string) error {

	serverConn, err := net.Dial("tcp", fmt.Sprintf("%v:%d", serverAddr, serverConnPort))
	if err != nil{
		return fmt.Errorf("client: error dialing to server (%v:%d): %v", serverAddr, serverConnPort, err)
	} else{
		log.Printf("client: insecure connection to server successfully established (%v:%d)\n", serverAddr, serverConnPort)
	}

	tunnel.CreateTunnel(serverConn, conn)

	return nil
}


func connectRemoteQuic(caCertLoc string){

	udpAddr := net.UDPAddr{
		IP: net.IPv4(0,0,0,0),
		Port: 2023,
	}
	fmt.Printf("ca cert loc is: %v\n", caCertLoc)
	tlsConf, _ := tlsconfig.ClientTLSConfig(caCertLoc)
	udpConn, _ := net.ListenUDP("udp", &udpAddr)

	conn, err := quic.Dial(context.Background(), udpConn, &net.UDPAddr{
		IP: net.IPv4(0,0,0,0),
		Port: 9002,
	}, tlsConf, nil)

	if err != nil{
		fmt.Printf("error while starting quic conn: %v. QUIC exit\n", err)
		return
	}else{
		fmt.Println("dialed server")
	}
	recevedBytes, _ := conn.ReceiveDatagram(conn.Context())
	fmt.Printf("msg from server: %v", string(recevedBytes))
}