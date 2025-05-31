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

func ConnectRemoteSecure(wg *sync.WaitGroup, errCh chan<- error, conn net.Conn, serverConnPort int, serverAddr, caCertLoc string) error {
	defer wg.Done()
	
	clientConfg, err := tlsconfig.ClientTLSConfig(caCertLoc)
	if err != nil{
		return fmt.Errorf("error fetching TLS config for client: %v",err)
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
	defer serverConn.Close()

	tunnel.CreateTunnel(serverConn, conn)

	return nil
}


func ConnectRemoteUnsec(wg *sync.WaitGroup, errCh chan<- error, serverConnPort int, conn net.Conn, serverAddr string) error {
	defer wg.Done()

	serverConn, err := net.Dial("tcp", fmt.Sprintf("%v:%d", serverAddr, serverConnPort))
	if err != nil{
		return fmt.Errorf("client: error dialing to server (%v:%d): %v", serverAddr, serverConnPort, err)
	} else{
		log.Printf("client: insecure connection to server successfully established (%v:%d)\n", serverAddr, serverConnPort)
	}

	tunnel.CreateTunnel(serverConn, conn)

	return nil
}