package tun

import (
	"context"
	"custom_vpn/internal/helpers"
	"custom_vpn/tlsconfig"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

func InterceptFromClient(ctx context.Context, errCh chan<- error, wg *sync.WaitGroup, port int, details *TunDetails){
	defer wg.Done()

	udpAddr := net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port}

	tlsConf, _ := tlsconfig.ServerTLSConfig()
	//tlsConf.NextProtos = []string{"vpn-quic"}

	udpConn, _ := net.ListenUDP("udp", &udpAddr)
	uuid, _ := helpers.GenUUID()

	qConf := quic.Config{
		EnableDatagrams: true,
	}

	tr := quic.Transport{
		Conn: udpConn,
		ConnContext: func(ctx context.Context, ci *quic.ClientInfo) (context.Context, error) {
			connId := uuid
			return context.WithValue(ctx, helpers.ConnId, connId), nil
		},
	}
	defer tr.Close()

	// create quic listener
	ln, err := tr.Listen(tlsConf, &qConf)
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}
	defer ln.Close()

	wg.Add(1)
	go helpers.CaptureCancel(ctx, wg, errCh, udpAddr.Port, ln)

	// read datagram
	for {
		conn, err := ln.Accept(ctx)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Println("listener closed")
				return	// just because the listener is closed, doesn't mean that we should return. there could be active streams
			}
			errCh <- fmt.Errorf("listener error : %v", err)
			continue
		}
		
		wg.Add(1)
		go GetDatagram(ctx, conn, errCh, details)
	}
}

func GetDatagram(ctx context.Context, conn quic.Connection, errCh chan<-error, details *TunDetails){

	for {
		data, err := conn.ReceiveDatagram(ctx)

		if err != nil{
			break;
		}
		
		// do some analysis on data
		log.Printf("got some data of len: %v", len(data))
		log.Printf("data is: %v", data[:min(20, len(data))])
		_ , err = details.TunIface.Write(data)
		if err != nil {
			errCh <- fmt.Errorf("failed to write to server TUN: %v", err)
		}
	}

}