package quic

import (
	"context"
	"custom_vpn/config"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

func ConnectRemoteQuic(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, remoteAddr *net.UDPAddr, caCertLoc string, conn net.Conn) {
	defer wg.Done()

	tlsConf, err := tlsconfig.ClientTLSConfig(caCertLoc)
	if err != nil{
		errCh <- fmt.Errorf("QUIC Client: %v. Will not initiate quic connection", err)
		return
	}
	
	// UDP Addr for a local bind
	// no port val means one is randomly choosen, but it can't be accessed
	udpAddrForRemoteComms := net.UDPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: 0,
	}
	
	udpConn, err := net.ListenUDP("udp4", &udpAddrForRemoteComms)
	if err != nil {
		errCh <- fmt.Errorf("QUIC Client: %v", err)
		return
	}

	// wrap UDP conn in quic
	tr := quic.Transport{Conn: udpConn}

	qConn, err := tr.Dial(ctx, remoteAddr, tlsConf, &config.ClientQuicConfig)
	if err != nil {
		errCh <- fmt.Errorf("quic client died while dialing remote: %v", err)
		return
	}else{
		log.Printf("established QUIC conn to remote")
	}

	// handle streams
	createStream(errCh, qConn, conn, config.ClientListnerPort, remoteAddr)

}

func createStream(errCh chan<- error, qConn quic.Connection, conn net.Conn, clientListenerPort int, remoteAddr *net.UDPAddr){
	// add a header based on local port num use incomming port to determing header

	// Determine protocol based on which port the client is listening on
	proto, err := determineProto(clientListenerPort)
	if err != nil{
		errCh <- err
		return
	}

	str, err := qConn.OpenStream()
	if err != nil{
		errCh <- fmt.Errorf("createStream: %v", err)
		return
	}
	
	// Write a Header to the stream before piping the conn
	// These IPs and Ports are useless. They mean nothing, and tell the end user nothing
	str.Write(proto)
	str.Write(remoteAddr.IP.To16())
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