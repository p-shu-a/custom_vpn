package quic

import (
	"context"
	"custom_vpn/tlsconfig"
	"custom_vpn/internal/helpers"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

/*
	Shifting to QUIC. Things will be handled differently:
	- you have create a UDP socket and get a UDP conn
	- convert that udp conn as a quic conn (using quic.transport)
	- start a transport listener which will accept connections
	- handle connections, and since quic conns have streams...
	- handle streams
*/

// start a QUIC listener on a port
func QuicServer(errCh chan<- error, ctx context.Context, port int, wg *sync.WaitGroup){ // a tls conf (so go back and modularize that)
	defer wg.Done()

	udpAddr := net.UDPAddr{Port: 9002, IP: net.IPv4(0,0,0,0)}
	// Unlike TCP's listen, we need to create a UDP socket
	// If you want to just listen to the plain-jane UDP port, start reading bytes
	udpConn, err := net.ListenUDP("udp", &udpAddr)
	if err != nil{
		errCh <- fmt.Errorf("failed to create UDP socket: %v", err)
		return
	}
	defer udpConn.Close()

	tlsConf, err := tlsconfig.ServerTLSConfig()
	if err != nil{
		errCh <- fmt.Errorf("failed to fetch TLS config: %v", err)
		return
	}
	/*	This is actually what makes the udp conn into a QUIC conn
		transport is pretty central to QUIC-go
	*/
	tr := &quic.Transport{
	 	Conn: udpConn,
	}
	defer tr.Close()

	// start a quic listener
	listener, err := tr.Listen(tlsConf, nil)
	if err != nil {
		errCh <- fmt.Errorf("error starting quic listener on %v: %v", udpAddr.Port, err)
	} else{
		log.Printf("server: QUIC listener active on %d\n",port)
	}
	defer listener.Close()

	wg.Add(1)
	go helper.CaptureCancel(wg, ctx, errCh, port, listener)

	for{
		quicConn, err := listener.Accept(ctx)
		if err != nil{
			if errors.Is(err, net.ErrClosed){
				log.Println("listener closed due to context cancel")
				return
			}
			// some errors with the accepted conns are okay and can be continued past
			// some errors must cause exit. How do i know which ones are which?
			// should implement a isExitWorthy() function to properly identify the different errors and we ought to continue or exit.
			continue;
		}
		wg.Add(1)
		go handleQuicConn(quicConn, errCh, wg, ctx)
	}
}

/*
	a quic conn has multiple streams, we need to separate those streams. and act on em
*/
func handleQuicConn(conn quic.Connection, errCh chan<- error, wg *sync.WaitGroup, ctx context.Context){
	defer wg.Done()

	log.Printf("Recieved a quic conn from %v\n", conn.RemoteAddr())
	if err := conn.SendDatagram([]byte("hello from server")); err != nil {
		errCh <- fmt.Errorf("failed to send datagram to client: %v", err)
		return
	}

	for {
		connID := rand.IntN(1000)
		strCtx := context.WithValue(ctx, "parentConnId", connID)
		stream, err := conn.AcceptStream(strCtx)
		if err != nil {
			// can we survive this error and continue?
			errCh <- fmt.Errorf("failed to accept stream: %v", err)
			return
		}
		wg.Add(1)
		go handleStream(stream, wg, ctx, errCh)
	}
}

func handleStream(stream quic.Stream, wg *sync.WaitGroup, ctx context.Context, errCh chan<- error){
	defer wg.Done()
	defer stream.Close()
	// how would i get the stream's conn id? pass it via the context?
	log.Printf("hey got a stream. stream id is %v and is parent conn's id is %v\n",stream.StreamID(), ctx.Value("parentConnId"))
	// what else do i do to a stream? whats the best way to read a stream?
	// what are some general principles for reading IO?
	// what is a stream composed of? i assume we've got headers, a body, what else?
	// if the conns are supposed to be used for vpns, should streams be forwards to some other endpoint?
	
	// read from the stream... what if the stream is being continuously written to?
	b := make([]byte, 8)
	for {
		n, err := stream.Read(b)
		log.Printf("from stream. n val: %v",n)
		log.Printf("from stream. b val: %q",b[:n])
		if err != nil { // could be an error could be io.EOF
			break
		}
	}
}