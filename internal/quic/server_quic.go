package quic

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"custom_vpn/config"
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

	// Local binding. Bind on provided port
	localAddr := net.UDPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: port,
	}

	// Create a UPD conn on specified address 
	udpConn, err := net.ListenUDP("udp", &localAddr)
	if err != nil{
		errCh <- fmt.Errorf("QUIC server: %v", err)
		return
	}
	defer udpConn.Close()

	tlsConf, err := tlsconfig.ServerTLSConfig()
	if err != nil{
		errCh <- fmt.Errorf("QUIC server: %v", err)
		return
	}

	/*
		Transport is pretty central to QUIC-go.	
		This is actually what "makes" the UDP Conn into a QUIC Conn.
		The ConnContext function is whats used to assign a connId to a connection
		The parent context is passed to the ConnContext func via the Lister.Accept() func
	*/
	tr := &quic.Transport{
	 	Conn: udpConn,
		ConnContext: func(ctx context.Context, ci *quic.ClientInfo) (context.Context, error) {
			connId, _ := helpers.GenUUID()
			return context.WithValue(ctx, helpers.ConnId, connId), nil
		},
	}
	defer tr.Close()

	// start a QUIC listener
	listener, err := tr.Listen(tlsConf, &config.ServerQuicConf)
	if err != nil {
		errCh <- fmt.Errorf("QUIC server: failed to start listener on %v: %v", localAddr.Port, err)
		return
	}else{
		log.Printf("QUIC Server: listening on port %v", localAddr.Port)
	}
	defer listener.Close()

	wg.Add(1)
	go helpers.CaptureCancel(cancelCtx, wg, errCh, localAddr.Port, listener)

	// still not happy with the error handling on following
	for{
		quicConn, err := listener.Accept(cancelCtx)
		if err != nil{
			if errors.Is(err, net.ErrClosed) {
				break	// just because the listener is closed, doesn't mean that we should return. there could be active streams
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
			var idleErr *quic.IdleTimeoutError
			if errors.As(err, &idleErr) || errors.Is(err, net.ErrClosed){
				errCh <- fmt.Errorf("QUIC server: failed to accept stream: %v", err)
				return
			}
			continue
		}
		wg.Add(1)
		go handleStream(stream.Context(), stream, wg, errCh)
	}
}


// Reads the stream header and dials the appropriate backend service
func handleStream(ctx context.Context, stream quic.Stream, wg *sync.WaitGroup, errCh chan<- error){
	defer wg.Done()
	defer stream.Close()
	log.Printf("Recieved Stream. stream-id: %v. Conn-Id: %v", stream.StreamID(), ctx.Value(helpers.ConnId))

	// buffers for reading from stream
	var protoBuff [4]byte
	var ipBuff [16]byte
	var port uint16

	// Deserialize by reading into buffers first
	io.ReadFull(stream, protoBuff[:])
	io.ReadFull(stream, ipBuff[:])
	binary.Read(stream, binary.BigEndian, &port)
	// read from stream into var-port. since port is a uint16, two bytes will be read

	streamHeader := helpers.StreamHeader{
		Proto: protoBuff,
		IP: net.IP(ipBuff[:]),
		Port: port,
	}

	log.Printf("from stream header. Proto (%v), IP (%v), Port (%v)",
		string(streamHeader.Proto[:]),
		streamHeader.IP.String(),
		streamHeader.Port)

	switch string(streamHeader.Proto[:]) {
	case "HTTP":
		backService := dialService(config.HTTPEndpointService, errCh)
		tunnel.QuicTcpTunnel(backService, stream)
	case "SSH":
		backService := dialService(config.SSHEndpointService, errCh)
		tunnel.QuicTcpTunnel(backService, stream)
	default:
		errCh <- fmt.Errorf("failed to identify stream protocol")
	}

}

// This functions dials some endpoint service and returns a net.conn
func dialService(endpointService net.TCPAddr, errCh chan<- error) net.Conn {

	targetConn, err := net.Dial("tcp", endpointService.String())
	if err != nil{
		errCh <- fmt.Errorf("error while connecting to ssh on server: %v", err)
		return nil
	}
	return targetConn
}