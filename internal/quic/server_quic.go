package quic

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
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

	// A hash which keeps track of conn ids
	connDb := make(map[int]bool, 100)
	/*
		This is actually what makes the udp conn into a QUIC conn.
		transport is pretty central to QUIC-go.
		should define connection id here.
	*/
	tr := &quic.Transport{
	 	Conn: udpConn,
		ConnContext: func(ctx context.Context, ci *quic.ClientInfo) (context.Context, error) {
			connId := generateConnId(connDb)
			return context.WithValue(ctx, helpers.ParentConnId, connId), nil
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
	go helpers.CaptureCancel(wg, cancelCtx, errCh, port, listener)

	for{
		quicConn, err := listener.Accept(context.Background()) /// or cancelCtx /// this is the context which gets passed to the transport's ConnContext func:ctx param

		//log.Printf("conn id after being set (from the conn):: %v", quicConn.Context().Value(helpers.ParentConnId))

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
		go handleQuicConn(quicConn, wg, errCh)
	}
}

/*
	a quic conn has multiple streams, we need to separate those streams. and act on em
*/
func handleQuicConn(conn quic.Connection, wg *sync.WaitGroup, errCh chan<- error){
	defer wg.Done()

	log.Printf("Recieved a quic conn from %v\n", conn.RemoteAddr())
	if err := conn.SendDatagram([]byte("hello from server")); err != nil {
		errCh <- fmt.Errorf("failed to send datagram to client: %v", err)
		return
	}

	log.Printf("handleQuicConn. conn id via context:: %v\n", conn.Context().Value(helpers.ParentConnId))

	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			// can we survive this error and continue?
			errCh <- fmt.Errorf("failed to accept stream: %v", err)
			return
		}
		wg.Add(1)
		go handleStream(stream, wg, errCh)
	}
}

func handleStream(stream quic.Stream, wg *sync.WaitGroup, errCh chan<- error){
	defer wg.Done()
	defer stream.Close()
	// how would i get the stream's conn id? pass it via the context?
	log.Printf("hey got a stream. stream id is %v and is parent conn's id is %v\n",
		stream.StreamID(),
		stream.Context().Value(helpers.ParentConnId))

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
	stream.Close()
	log.Printf("stream closed goodbye")

}

/*
	
*/
func generateConnId(connDb map[int]bool) int{
	for {
		connId := rand.IntN(1000)
		if connDb[connId] {
			continue
		} else {
			connDb[connId] = true
			return connId
		}
	}
}