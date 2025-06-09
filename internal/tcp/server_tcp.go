package tcp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"custom_vpn/internal/helpers"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
)

// Creates a TCP connection on the specified port. Utilizes transport layer scurity
func ListenAndServeWithTLS(cancelCtx context.Context, errCh chan<- error, wg *sync.WaitGroup, port int, endpointService net.TCPAddr) {
	defer wg.Done()

	serverConfig, err := tlsconfig.ServerTLSConfig()
	if err != nil {
		errCh <- fmt.Errorf("TLS Server: error getting server config: %v", err)
		return
	}

	tcpAddr := net.TCPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: port,
	}

	listener, err := tls.Listen("tcp", tcpAddr.String(), serverConfig)
	if err != nil {
		errCh <- fmt.Errorf("TLS Server: error while starting listener: %v", err)
		return
	} else {
		log.Printf("TLS Server: listening on port %d",port)
	}
	defer listener.Close()

	/*
		This go func's jobs is to listen for cancel() which gets called when SIGTERM OR SIGINT is sent.
		It then closes the listener and sends the error down the channel.
		Added the wg.Add() to ensure the error gets printed to the screen before the program exits
	*/
	wg.Add(1)
	go helpers.CaptureCancel(cancelCtx, wg, errCh, tcpAddr.Port, listener)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed){
				return
			}
			errCh <- fmt.Errorf("unable to accept connection: %v", err)
			continue
		}
		go handleClientConn(clientConn, errCh, endpointService)
	}
}

// Starts a raw TCP listener on given port
func ListenAndServeNoTLS(cancelCtx context.Context, errCh chan<- error, wg *sync.WaitGroup, port int, endpointService net.TCPAddr) {
	defer wg.Done()

	tcpAddr := net.TCPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: port,
	}
	// start listener
	listener, err := net.ListenTCP("tcp", &tcpAddr)
	if err != nil{
		errCh <- fmt.Errorf("TCP Server: failed to start listener (on-tls): %v", err)
		return
	} else {
		log.Printf("TCP Server: listening on port %d", tcpAddr.Port)
	}
	defer listener.Close()

	// capture cancel()
	wg.Add(1)
	go helpers.CaptureCancel(cancelCtx, wg, errCh, tcpAddr.Port, listener)

	// start accepting connections
	for {
		clientConn, err := listener.Accept()
		if err != nil{
			if errors.Is(err, net.ErrClosed){
				return
			}
			errCh <- fmt.Errorf("TCP Server: unable to accept connection: %v", err)
			continue
		}
		go handleClientConn(clientConn, errCh, endpointService)
	}
}

// Dials the provided endpoint service 
func handleClientConn(clientConn net.Conn, errCh chan<- error, endpointService net.TCPAddr) {

	log.Printf("server: Recieved a conn on %v from %v\n", clientConn.LocalAddr(), clientConn.RemoteAddr())

	targetConn, err := net.Dial("tcp", endpointService.String())
	if err != nil{
		errCh <- fmt.Errorf("error while connecting to ssh on server: %v", err)
		return
	}
	tunnel.CreateTunnel(targetConn, clientConn)
}
