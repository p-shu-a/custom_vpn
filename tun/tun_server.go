package tun

import (
	"context"
	"custom_vpn/config"
	"custom_vpn/internal/helpers"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)



func StartTunListener(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, details *TunDetails) {
	
	defer wg.Done()
	defer details.TunIface.Close()

	// udp for local bind
	localBind := net.UDPAddr{
		IP: net.ParseIP("10.0.0.5"),
		Port: config.ServerPortTun,
	}

	// udp conn
	udpConn, err := net.ListenUDP("udp", &localBind)
	if err != nil {
		errCh <- fmt.Errorf("died firing up server's udp listener %v", err)
		return
	}

	// quic transport
	tr := quic.Transport{
		Conn: udpConn,
		ConnContext: func(ctx context.Context, ci *quic.ClientInfo) (context.Context, error) {
			connId, _ := helpers.GenUUID()
			return context.WithValue(ctx, helpers.ConnId, connId), nil
		},
	}

	// tlsconf
	tlsConf, err := tlsconfig.ServerTLSConfig()
	if err != nil {
		errCh <- fmt.Errorf("died fetching tls config")
		return
	}

	// quic conf...in config package

	// create listener
	ln, err := tr.Listen(tlsConf, &config.ServerQuicConf)
	if err != nil {
		errCh <- fmt.Errorf("failed opening quic listener")
		return
	}else{
		log.Printf("QUIC listener active on %v", localBind.Port)
	}

	wg.Add(1)
	go helpers.CaptureCancel(ctx, wg, errCh, localBind.Port, ln)

	// accept on listner
	for {
		qConn, err := ln.Accept(ctx)
		if err != nil {
			if errors.Is(err, net.ErrClosed){
				break
			}
			continue
		}
		log.Printf("Recieved qCONN on %v", tr.Conn.LocalAddr().String())
		wg.Add(1)
		go handleConn(qConn.Context(), wg, errCh, qConn, details)

	}


}

func handleConn(ctx context.Context, wg *sync.WaitGroup, errCh chan <- error, qConn quic.Connection, details *TunDetails){
	defer wg.Done()

	for{
		str, err := qConn.AcceptStream(ctx)
		if err != nil {
			errCh <- fmt.Errorf("failed to accept stream")
		}

		log.Print("about to fire off quic tun tunnels")
		wg.Add(1)
		go handleStream(wg, str, details)
	}

}

func handleStream(wg *sync.WaitGroup, str quic.Stream, details *TunDetails){
	defer wg.Done()
	defer str.Close()

	tunnel.QuicTunTunnel(str, details.TunIface)
}