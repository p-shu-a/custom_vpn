package tcp

import (
	"crypto/tls"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"fmt"
	"log"
	"net"
)

/*
	whats the issue i'm having right now?
	- ideally the client, would only accept one incoming connection on port 2022. not multiple. multiple is the default
	- what should happen is that if the server kills the connection, you should be able to reattempt
*/

func ConnectRemoteSecure(serverConnPort int, conn net.Conn, serverAddr string, caCertLoc string) error {
	
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


func ConnectRemoteUnsec(serverConnPort int, conn net.Conn, serverAddr string) error {

	serverConn, err := net.Dial("tcp", fmt.Sprintf("%v:%d", serverAddr, serverConnPort))
	if err != nil{
		return fmt.Errorf("client: error dialing to server (%v:%d): %v", serverAddr, serverConnPort, err)
	} else{
		log.Printf("client: insecure connection to server successfully established (%v:%d)\n", serverAddr, serverConnPort)
	}

	tunnel.CreateTunnel(serverConn, conn)

	return nil
}