package quic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"custom_vpn/internal/helpers"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"

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

	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			// can we survive this error and continue?
			errCh <- fmt.Errorf("failed to accept stream: %v", err)
			return
		}
		log.Print("hey got a stream. handling it ")
		wg.Add(1)
		go handleStream(stream.Context(), stream, wg, errCh)
	}
}

func handleStream(ctx context.Context, stream quic.Stream, wg *sync.WaitGroup, errCh chan<- error){

	defer wg.Done()
	defer stream.Close()

	log.Printf("hey got a stream. stream id is %v. Conn-Id is %v", stream.StreamID(), ctx.Value(helpers.ConnId))

	header := [4]byte{}
	io.ReadFull(stream, header[:4])
	log.Printf("header from stream: %v", string(header[:]))

	// forward the stream to a dest based on the header
	// b, _ := io.ReadAll(stream)
	// log.Printf("reading from stream: %v", b)
	// log.Printf(string(b))
	proto := header[0:4]
	log.Printf("proto: %q", string(proto))

	switch strings.Replace(string(proto), "\x00", "", -1) {
	case "HTTP":
		log.Println("proto is http")
		backService := dialService("127.0.0.1", 8080, errCh)
		tunnel.QuicTcpTunnel(backService, stream)
	case "SSH":
		log.Println("proto is ssh")
		backService := dialService("127.0.0.1", 22, errCh)
		tunnel.QuicTcpTunnel(backService, stream)
	}

}

func dialService(serviceAddr string, servicePort int, errCh chan<- error) net.Conn {

	log.Printf("quic server: dialing service on %v", serviceAddr)

	targetConn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", serviceAddr, servicePort))
	if err != nil{
		errCh <- fmt.Errorf("error while connecting to ssh on server: %v", err)
		return nil
	}
	return targetConn
}