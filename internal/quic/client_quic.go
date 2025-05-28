package quic

import (
	"context"
	"custom_vpn/tlsconfig"
	"fmt"
	"log"
	"net"

	"github.com/quic-go/quic-go"
)

func ConnectRemoteQuic(caCertLoc string, remoteAddr string) {

	tlsConf, err := tlsconfig.ClientTLSConfig(caCertLoc)
	if err != nil{
		log.Printf("failed to fetch tls config for client: %v\n",tlsConf)
		return
	}
	

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
	str, _ := qConn.OpenStream()
	str.Write([]byte("hello from client"))
	str.Close()

}