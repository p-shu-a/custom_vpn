package tcp

import (
	"context"
	"crypto/tls"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"custom_vpn/internal/helpers"
)

// listen and server, with transport layer scurity
func ListenAndServeWithTLS(port int, ctx context.Context, errCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	serverConfig, err := tlsconfig.ServerTLSConfig()
	if err != nil {
		errCh <- fmt.Errorf("error getting server config: %v", err)
		return
	} else {
		log.Println("TLS Server: TLS config successfully acquired")
	}

	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), serverConfig)
	if err != nil {
		errCh <- fmt.Errorf("error while starting listener on %d: %v", port, err)
		return
	} else {
		log.Printf("TLS Server: Listening on %d\n",port)
	}
	defer listener.Close()

	/*
		This go func's jobs is to listen for cancel() which gets called when SIGTERM OR SIGINT.
		It then closes the listener and sends the error down the channel.
		Added the wg.Add() to ensure the error gets printed to the screen before the program exits
	*/
	wg.Add(1)
	go helper.CaptureCancel(wg, ctx, errCh, port, listener)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed){
				return
			}
			errCh <- fmt.Errorf("unable to accept connection: %v", err)
			continue
		}
		go handleClientConn(clientConn, errCh)
	}
}

func ListenAndServeNoTLS(port int, ctx context.Context, errCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	// start listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d",port))
	if err != nil{
		errCh <- fmt.Errorf("error starting listener (on-tls) on %d: %v", port, err)
		return
	} else{
		log.Printf("server: non-TLS Listening on %d\n",port)
	}
	defer listener.Close()

	// capture cancel()
	wg.Add(1)
	go helper.CaptureCancel(wg, ctx, errCh, port, listener)

	// start accepting connections
	for {
		clientConn, err := listener.Accept()
		if err != nil{
			if errors.Is(err, net.ErrClosed){
				return
			}
			errCh <- fmt.Errorf("server: unable to accept connection on %d: %v", port, err)
			continue
		}
		go handleClientConn(clientConn, errCh)		
	}
}

func handleClientConn(clientConn net.Conn, errCh chan<- error) {

	log.Printf("server: Recieved a conn on %v from %v\n", clientConn.LocalAddr(), clientConn.RemoteAddr())

	targetConn, err := net.Dial("tcp", "127.0.0.1:22")
	if err != nil{
		errCh <- fmt.Errorf("error while connecting to ssh on server: %v", err)
		return
	}
	tunnel.CreateTunnel(targetConn, clientConn)
}
