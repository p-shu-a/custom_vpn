package main

import (
	"context"
	"math/rand/v2"
	"crypto/tls"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/quic-go/quic-go"
)

/*
	Helper method to return a context which will track shutdown/terms
	Basically, we're setting up the signal handling.
	Create a channel which gets notified when SIGINT or SIGTERM are called
	When a signal is picked up, we unblock and call the cancel(), cancelling the context
	Cancelling the context leads to ctx.Done() being called in other places (the server funcs)
*/
func setupShutdownHelper() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // block here until there is something is recieved
		cancel()
	}()
	return ctx
}


func main(){

	ctx := setupShutdownHelper()
	var wg sync.WaitGroup
	
	/* 
		Shifted strategy: since we have go-routines called by go-routines, this leads to a bastardized mix of error handling.
		Some places I was returning errors, others I was writing to an error channel. nothing worse than a mix.
		So, I'll be writing all errors to a channel.
		(using errgroup package is another approach, but not gonna look into that righ now...)
	*/
	errCh := make(chan error)
	done := make(chan struct{})		// this done channel was created to ensure the ordering of the logs
	go errorCollector(errCh, done)

	// since we call both servers in go-routines
	wg.Add(1)
	go listenAndServeWithTLS(9001, ctx, errCh, &wg)

	wg.Add(1)
	go listenAndServeNoTLS(9000, ctx, errCh, &wg)

	wg.Add(1)
	go quicServer(errCh, ctx, 9002, &wg)

	wg.Wait()
	close(errCh)
	/*
		Wait for the done channel to close (after the errorCollector go-routine is closed)
		This way, the errors eminating from the server functions are all logged in order(!) before we exit
	*/
	<-done
	log.Println("server: All servers closed. Exiting...")
}

/*
	Sole purpore of this collector func is to print errors as they come down the err channel
*/
func errorCollector(errCh <-chan error, done chan struct{}){
	/*
		re:defer close(done)
		since this function blocks waiting for things to come down the errCh, it exists until errCh closes
		This has the effect where "done" channel close is defered until after the errCh is closed (in main)
	*/
	defer close(done)
	for err := range errCh{
		log.Printf("ERROR: %v\n", err)
	}
}

// listen and server, with transport layer scurity
func listenAndServeWithTLS(port int, ctx context.Context, errCh chan<- error, wg *sync.WaitGroup) {
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
	go captureCancel(wg, ctx, errCh, port, listener)

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

func listenAndServeNoTLS(port int, ctx context.Context, errCh chan<- error, wg *sync.WaitGroup) {
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
	go captureCancel(wg, ctx, errCh, port, listener)

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

/*
	We're interesting in "interface satisfaction"
	Perhaps its not accurate to use the "parent-child" terminology here, but indulge me.
	net.Listener is a Closeable interface. Why? because it implements the Close() method.
	And since our definition of a closeablelistener is "a closeable interface implements as close() method" and 
	net.Listener implments close(), thus it satisfies.
	Very Cool
*/
type CloseableListener interface {
	Close() error
}

/* 
	This function closes a listener, doesn't matter if its a TCP listener or a QUIC listener.
	That is why we defined the CloseableListener interface.
*/
func captureCancel(wg *sync.WaitGroup, ctx context.Context, errCh chan<- error, port int, listener CloseableListener){
	defer wg.Done()
	<-ctx.Done()			// block here until cancel()
	listener.Close()		// call our closeable listeners close() function
	errCh <-fmt.Errorf("%v: listener closed on port-%d due to SIGTERM", ctx.Err(), port)
}

/*
	Shifting to QUIC. Things will be handled differently:
	- you have create a UDP socket and get a UDP conn
	- convert that udp conn as a quic conn (using quic.transport)
	- start a transport listener which will accept connections
	- handle connections, and since quic conns have streams...
	- handle streams
*/

 // start a QUIC listener on a port
func quicServer(errCh chan<- error, ctx context.Context, port int, wg *sync.WaitGroup){ // a tls conf (so go back and modularize that)
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
	go captureCancel(wg, ctx, errCh, port, listener)

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