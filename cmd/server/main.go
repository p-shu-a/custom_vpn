package main

import (
	"context"
	"crypto/tls"
	"custom_vpn/tlsconfig"
	"custom_vpn/tunnel"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// helper method to return a context which will track shutdown/terms
// basically, we're setting up the signal handling
func setupShutdownHelper() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // block here until there is something in the channel
		cancel()
	}()
	return ctx
} // this is pretty fuckin smart


func main(){

	ctx := setupShutdownHelper()
	var wg sync.WaitGroup
	
	// Shifted strategy: since we have go-routines called by go-routines, there leads to a bastardized mix of error handling
	// some places i was returning errors, others i was writing to an error channel. nothing worse than a mix.
	// so, i'll be writing all errors to a channel.
	// using errgroup package is another approach, but not gonna look into that righ now...
	errCh := make(chan error)
	go errorCollector(errCh)

	// since we call both servers in go-routines
	wg.Add(1)
	go listenAndServeWithTLS(9001, ctx, errCh, &wg)

	wg.Add(1)
	go listenAndServeNoTLS(9000, ctx, errCh, &wg)

	wg.Wait()
	log.Println("All servers closed. Exiting...")
	close(errCh)
}


func errorCollector(errCh <-chan error){
	for err := range errCh{
		log.Printf("ERROR: %v\n", err)
	}
}

// Listen for incoming connections
func listenAndServeWithTLS(port int, ctx context.Context, errCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	serverConfig, err := tlsconfig.ServerTLSConfig()
	if err != nil {
		errCh <- fmt.Errorf("error getting server config: %v", err)
		return
	}

	listener, err := tls.Listen("tcp6", fmt.Sprintf(":%d", port), serverConfig)
	if err != nil {
		errCh <- fmt.Errorf("error while starting listener on %d:%v", port, err)
		return
	}
	defer listener.Close()
	
	// a for without a condition is go-lang's "while". this will continue until a break, or a return
	for {
		select {
		case <- ctx.Done():
			return
		default:
			clientConn, err := listener.Accept()
			if err != nil {
				if _, ok := err.(net.Error); ok {		// Type assert that the error is a net.error type
					continue
				}
				errCh <- fmt.Errorf("unable to accept connection: %v", err)		// if it ain't, return error
				return
			}
			go handleClientConn(clientConn, errCh)
		}
	}
}

func listenAndServeNoTLS(port int, ctx context.Context, errCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	listener, err := net.Listen("tcp6", fmt.Sprintf(":%d",port))
	if err != nil{
		errCh <- fmt.Errorf("server: error starting listener (on-tls) on %d: %v", port, err)
		return
	}
	defer listener.Close()

	for {
		select {
		case <- ctx.Done():
			return
		default:
			clientConn, err := listener.Accept()
			// this error handling needs to be looked at
			if err != nil{
				errCh <- fmt.Errorf("server: unable to accept connection on %d: %v", port, err)
				return
			}
			go handleClientConn(clientConn, errCh)
		}
		
	}
}

func handleClientConn(clientConn net.Conn, errCh chan<- error) {

	log.Printf("Recieved a conn on %v from %v\n", clientConn.LocalAddr(), clientConn.RemoteAddr())

	targetConn, err := net.Dial("tcp", "127.0.0.1:22")
	if err != nil{
		errCh <- fmt.Errorf("error while connecting to ssh on server: %v", err)
		return
	}
	tunnel.CreateTunnel(targetConn, clientConn)
}