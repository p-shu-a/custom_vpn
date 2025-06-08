package quic

import (
	"context"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

func ConnectRemoteQuic(wg *sync.WaitGroup, errCh chan<- error, remotePort int, remoteAddr, caCertLoc string, conn net.Conn) {
	defer wg.Done()

	tlsConf, err := tlsconfig.ClientTLSConfig(caCertLoc)
	if err != nil{
		errCh <- fmt.Errorf("quic: failed to fetch tls config for client: %v. Will not initiate quic connection", err)
		return
	}
	
	// crate addr object to bind to some local port for comms with server
	// no port val means one is randomly choosen, but it can't be accessed
	udpAddr := net.UDPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: 0,
	}
	
	udpConn, err := net.ListenUDP("udp4", &udpAddr)
	if err != nil {
		errCh <- fmt.Errorf("failed to bind to a local UDP port. %v", err)
		return
	}
	// assignedPort := udpConn.LocalAddr().(*net.UDPAddr).Port

	// wrap UDP conn in quic
	tr := quic.Transport{Conn: udpConn}
	qconf := quic.Config{
		EnableDatagrams: true,
	}
	
	// this is the remote address to dial
	remoteUDPAddr := net.UDPAddr{
		IP: net.ParseIP(remoteAddr),
		Port: remotePort,
	}
	qConn, err := tr.Dial(context.Background(), &remoteUDPAddr, tlsConf, &qconf)
	if err != nil {
		errCh <- fmt.Errorf("quic client died while dialing remote: %v", err)
		return
	}else{
		log.Printf("established QUIC conn to remote")
	}

	// handle streams
	createStream(errCh, qConn, conn, 2022, &udpAddr)

}

func createStream(errCh chan<- error, qConn quic.Connection, conn net.Conn, incomingPort int, remoteAddr *net.UDPAddr){
	// add a header based on local port num use incomming port to determing header

	proto, err := determineProto(incomingPort)
	if err != nil{
		errCh <- err
		return
	}

	str, err := qConn.OpenStream()
	if err != nil{
		errCh <- fmt.Errorf("failed to open stream: %v", err)
		return
	}
	
	str.Write(proto)
	str.Write(remoteAddr.IP.To16())
	log.Printf("the prort is: %v", qConn.LocalAddr().(*net.UDPAddr).Port)
	binary.Write(str, binary.BigEndian, uint16(qConn.LocalAddr().(*net.UDPAddr).Port))

	tunnel.QuicTcpTunnel(conn, str)
}

// This is not effective validation.
// Values defining protocols should be constants shared across client and server
// and should live in a helper
func determineProto(port int) ([]byte,error) {
	if port == 2022{
		return []byte("HTTP"), nil			// these values should be consts/enums
	} else if port == 2024{
		return []byte("SSH"), nil
	} 
	return nil, fmt.Errorf("unsupported protocol")
}