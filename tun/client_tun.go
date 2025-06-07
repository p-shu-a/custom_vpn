package tun

import (
	"context"
	"crypto/tls"
	"custom_vpn/internal/helpers"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

// Reads from TUN and forwards to server
func ForwardToServer(ctx context.Context, errCh chan<- error, wg *sync.WaitGroup, details *TunDetails, caCertLoc string){
	defer wg.Done()

	//tlsConf, _ := tlsconfig.ClientTLSConfig(caCertLoc)
	tlsConf := tls.Config{
		InsecureSkipVerify: true,
		NextProtos: []string{"vpn-quic"},
	}
	
	udpAddr := net.UDPAddr{
		IP: net.IPv4zero,
		// choose a port. i don't care
	}

	log.Printf("my outgoing udp addr is : %v", udpAddr.String())
	udpConn, _ := net.ListenUDP("udp", &udpAddr)
	defer udpConn.Close()

	wg.Add(1)
	go helpers.CaptureCancel(ctx, wg, errCh, 9005, udpConn)

	
	tr := quic.Transport{
		Conn: udpConn,
	}
	defer tr.Close()

	quicConf := quic.Config{
		EnableDatagrams: true,
	}

	remoteAddr := net.UDPAddr{
		IP: net.ParseIP("127.0.0.1"),
		Port: 9005,
	}

	log.Printf("going to dial: %v:%d", remoteAddr.IP.String(), remoteAddr.Port)
	
	qConn, err := tr.Dial(ctx, &remoteAddr, &tlsConf, &quicConf) 
	if err != nil{
		errCh<- fmt.Errorf("failed to dial server: %v", err)
		return
	}else{
		log.Print("dialed remote")
	}

	log.Print("about to read from the interface")

	defer details.TunIface.Close()

	buff := make([]byte, 1500)
	for {
		n, err := details.TunIface.Read(buff)
		if err != nil {
			errCh<- fmt.Errorf("failed to read packets from interface: %v", err)
			continue
		}

		log.Printf("forarding %d bytes to server", n)
		err = qConn.SendDatagram(buff[:n])

		if err != nil {
			errCh<- fmt.Errorf("failed to send datagram to remote: %v", err)
			return
		}else{
			log.Printf("sent %d bytes to client", n)
		}
	}
}