package quic

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"custom_vpn/internal/helpers"
	"custom_vpn/tlsconfig"

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

// start a QUIC listener on specified port
func QuicServer(cancelCtx context.Context, errCh chan<- error, wg *sync.WaitGroup, port int){
	defer wg.Done()

	udpAddr := net.UDPAddr{Port: 9002, IP: net.IPv4(0,0,0,0)}

	// Unlike TCP's listen, we need to explicitly create a UDP socket. with tcp listen (the lib handles that for you)
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

	uuid, err := helpers.GenUUID()
	if err != nil {
		errCh <- fmt.Errorf("failed to genereate UUID: %v", err)
		return
	}
	/*
		This is actually what makes the udp conn into a QUIC conn.
		transport is pretty central to QUIC-go.
		The ConnContext function is whats used to assign a connId to a connection
		the parent context is passed to the ConnContext func via the Lister.Accept() func
	*/
	tr := &quic.Transport{
	 	Conn: udpConn,
		ConnContext: func(ctx context.Context, ci *quic.ClientInfo) (context.Context, error) {
			connId := uuid
			return context.WithValue(ctx, helpers.ConnId, connId), nil
		},
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
	go helpers.CaptureCancel(cancelCtx, wg, errCh, port, listener)

	for{
		quicConn, err := listener.Accept(cancelCtx)

		if err != nil{
			if errors.Is(err, net.ErrClosed) {
				log.Println("listener closed")
				return	// just because the listener is closed, doesn't mean that we should return. there could be active streams
			}
			// some errors with the accepted conns are okay and can be continued past
			// some errors must cause exit. How do i know which ones are which?
			// should implement a isExitWorthy() function to properly identify the different errors and we ought to continue or exit.
			continue
		}
		wg.Add(1)
		go handleQuicConn(quicConn.Context(), quicConn, wg, errCh)
	}
}

/*
	a quic conn has multiple streams, we need to separate those streams. and act on em
*/
func handleQuicConn(ctx context.Context, conn quic.Connection, wg *sync.WaitGroup, errCh chan<- error){
	defer wg.Done()

	log.Printf("Recieved a quic conn from %v\n", conn.RemoteAddr())
	if err := conn.SendDatagram([]byte("hello from server")); err != nil {
		errCh <- fmt.Errorf("failed to send datagram to client: %v", err)
		return
	}

	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			// can we survive this error and continue?
			errCh <- fmt.Errorf("failed to accept stream: %v", err)
			return
		}
		wg.Add(1)
		go handleStream(stream.Context(), stream, wg, errCh)
	}
}

func handleStream(ctx context.Context, stream quic.Stream, wg *sync.WaitGroup, errCh chan<- error){

	defer wg.Done()
	defer stream.Close()

	log.Printf("hey got a stream. stream id is %v. Conn-Id is %v", stream.StreamID(), ctx.Value(helpers.ConnId))

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