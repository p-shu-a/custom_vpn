package tcp

import (
	"crypto/tls"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"fmt"
	"log"
	"net"
	"sync"
)

// Connect via TCP to remote server with TLS
func ConnectRemoteSecure(wg *sync.WaitGroup, errCh chan<- error, conn net.Conn, caCertLoc string, serverAddr *net.TCPAddr) error {
	defer wg.Done()
	
	clientConfg, err := tlsconfig.ClientTLSConfig(caCertLoc)
	if err != nil{
		return fmt.Errorf("error fetching TLS config for client: %v",err)
	}

	// if you wonder where the "conn.close()" are, they're in the tunnel logic
	serverConn, err := tls.Dial("tcp",
								serverAddr.String(), 
								clientConfg)
	if err != nil{
		return fmt.Errorf("error dialing to server (%v): %v", serverAddr.String(), err)
	} else {
		log.Printf("client: established secure TCP conn to server %v", serverAddr.String())
	}
	defer serverConn.Close()

	tunnel.CreateTunnel(serverConn, conn)

	return nil
}

// Connect to remote server with Raw TCP
func ConnectRemoteUnsec(wg *sync.WaitGroup, errCh chan<- error, conn net.Conn, serverAddr *net.TCPAddr) error {
	defer wg.Done()

	serverConn, err := net.Dial("tcp", serverAddr.String())
	if err != nil{
		return fmt.Errorf("client: error dialing to server (%v): %v", serverAddr.String(), err)
	} else{
		log.Printf("client: established insecure connection to server %v", serverAddr.String())
	}

	tunnel.CreateTunnel(serverConn, conn)

	return nil
}