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

	go connectRemoteQuic(caCertLoc, serverAddr)

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


func connectRemoteQuic(caCertLoc string, remoteAddr string){

	tlsConf, err := tlsconfig.ClientTLSConfig(caCertLoc)
	if err != nil{
		log.Printf("failed to fetch tls config for client: %v\n",tlsConf)
		return
	}
	tlsConf.InsecureSkipVerify = true

	// this is the UDP address we'll be sending from and recieving data back
	udpAddr := net.UDPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: 2023,
	}
	udpConn, err := net.ListenUDP("udp4", &udpAddr)
	if err != nil {
		log.Printf("failed to start udp listener on %v\n", udpAddr)
		return
	}
	// wrap UDP conn in quic
	tr := quic.Transport{Conn: udpConn}
	qconf := quic.Config{
		EnableDatagrams: true,
	}
	// this is the remote address to dial
	remoteUDPAddr := net.UDPAddr{
		IP: net.ParseIP(remoteAddr),
		Port: 9002,
	}
	qConn, err := tr.Dial(context.Background(), &remoteUDPAddr, tlsConf, &qconf)
	if err != nil {
		log.Printf("client: died while dialing remote: %v", err)
		return
	}

	byteArr, err := qConn.ReceiveDatagram(qConn.Context())
	if err != nil {
		log.Printf("some error while receving datagram: %v", err)
		return
	}
	fmt.Printf("byte from server: %v",string(byteArr))

	// handle streams
	

}