package quic

import (
	"context"
	"custom_vpn/tlsconfig"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

func ConnectRemoteQuic(wg *sync.WaitGroup, errCh chan<- error, port int, remoteAddr, caCertLoc string) {
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
		Port: 9002,
	}

	qConn, err := tr.Dial(context.Background(), &remoteUDPAddr, tlsConf, &qconf)
	if err != nil {
		errCh <- fmt.Errorf("quic client died while dialing remote: %v", err)
		return
	}else{
		log.Printf("established successful QUIC conn to remote")
	}

	byteArr, err := qConn.ReceiveDatagram(qConn.Context())
	if err != nil {
		errCh <- fmt.Errorf("encountered an error while receving datagram: %v", err)
		return
	}
	fmt.Printf("byte from server: %v",string(byteArr))

	// handle streams
	str, _ := qConn.OpenStream()
	str.Write([]byte("hello from client"))
	str.Close()

}