package tun

import (
	"context"
	"custom_vpn/config"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"log"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

func StartTunReader(ctx context.Context, wg *sync.WaitGroup, errCh chan<-error, remoteAddr net.Addr, details *TunDetails) {
	defer wg.Done()
	//defer details.TunIface.Close()

	// local udp port
	outgoing := net.UDPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: 0,
	}
	udpConn, err := net.ListenUDP("udp", &outgoing)
	if err != nil {
		log.Fatalf("failed to create local UDP conn: %v", err)
	}
	defer udpConn.Close()

	// quick transport
	tr := quic.Transport{
		Conn: udpConn,
	}

	// quic config
	// in config file

	// TLS config
	tlsConf, err := tlsconfig.ClientTLSConfig("")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("about to dial: %v", remoteAddr.String())
	// dial
	qConn, err := tr.Dial(ctx, remoteAddr, tlsConf, &config.ClientQuicConfig)
	if err != nil{
		log.Fatalf("failed to dial remote: %v", err)
	}
	log.Print("finished dial")
	// use the conn
	workThePipe(qConn.Context(), wg, errCh, details, qConn)
	
}


func workThePipe(ctx context.Context, wg *sync.WaitGroup, errCh chan<-error,details *TunDetails, qConn quic.Connection){

	str, err := qConn.OpenStream()
	if err != nil {
		log.Fatal(err)
	}
	defer str.Close()

	log.Print("about to use QuicTunTunnel to read from interface and write to stream... could die here...and its not a go func")
	tunnel.QuicTunTunnel(str, details.TunIface)


}
