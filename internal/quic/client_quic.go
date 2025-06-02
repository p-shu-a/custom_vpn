package quic

import (
	"context"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
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
	
	// this is the UDP address we'll be sending from and recieving data back
	udpAddr := net.UDPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: 2023,
	}
	udpConn, err := net.ListenUDP("udp4", &udpAddr)
	if err != nil {
		errCh <- fmt.Errorf("failed to start udp listener on %v", udpAddr)
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
		Port: remotePort,
	}
	qConn, err := tr.Dial(context.Background(), &remoteUDPAddr, tlsConf, &qconf)
	if err != nil {
		errCh <- fmt.Errorf("quic client died while dialing remote: %v", err)
		return
	}else{
		log.Printf("established successful QUIC conn to remote")
	}

	// handle streams
	createStream(errCh, qConn, conn)

}

func createStream(errCh chan<- error, qConn quic.Connection, conn net.Conn){
	// add a header based on local port num
	// header should be: Type (4byte), IP (4bytes for ipv4, 16bytes for ipv6), and Port (1byte)
	header := make([]byte, 4)
	copy(header[0:4], []byte("HTTP"))
	// copy(header[5:10], []byte("127.0.0.1"))			// this is garbage. we're hard coding it
	// copy(header[10:], []byte("8080"))				// 

	// whats the diff between openstreamsync and openstream?
	str, err := qConn.OpenStream()   // make stream uni for now
	if err != nil{
		errCh <- fmt.Errorf("failed to open stream: %v", err)
	}else{
		log.Print("successfully opened stream to remote")
	}
	
	str.Write(header)
	tunnel.QuicTcpTunnel(conn, str)
	
}