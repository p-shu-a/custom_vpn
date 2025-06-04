package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"custom_vpn/internal/helpers"
	"custom_vpn/internal/quic"
	"custom_vpn/internal/tcp"
)

/*
	A client needs to be more explicit about what it needs to do.
	A server makes services available, a client has to be decisive about what it needs to do.
	That's why there are a lot more options in the client code.
	Any of the default actions are me being lazy, because I don't want to spend the time to make the behaviour opt in
*/

func main(){

	clientListenerPort := flag.Int("p", 2022, "Port used to connect to client (via nc, socat)")
	serverAddress := flag.String("addr", "localhost", "Server IP Address")
	useRawTcp := flag.Bool("raw", false, "Use Raw TCP")
	useTls := flag.Bool("tls", false, "Use TLS")
	useQuic := flag.Bool("quic", true, "Use QUIC protocol (default). If false, TCP+TLS will be used")
	caCertLoc := flag.String("ca", "", "specify a custom CA cert")
	flag.Parse()

	errCh := make(chan error, 1)
	done := make(chan struct{})
	go helpers.ErrorCollector(errCh, done)

	var wg sync.WaitGroup

	wg.Add(1)
	// add local listener calls to multiple ports here
	// also add context to start listner, just like with server, to kill client if sigterm is sent
	startLocalListener(errCh, &wg, *clientListenerPort, *serverAddress, *caCertLoc, *useQuic, *useTls, *useRawTcp)

	wg.Wait()
	
	<-done
	close(errCh)

	log.Print("all listeners stopped...exiting client")
}


/*
	Creates a tcp net.conn on the clientListnerPort
	The user can establish multiple connections to this port. but why?
	based on user's selection of TLS or not (transSec)
	tries to establish remote conn with or without tls
	We need to be able to start multiple listeners (HTTP, SSH, etc...)
*/
func startLocalListener(errCh chan<-error, wg *sync.WaitGroup, clientListenerPort int, serverAddr, caCertLoc string, useQuic, useTls, useRawTcp bool) {
	defer wg.Done()

	// Start a local listener on the provided port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", clientListenerPort))
	if err != nil{
		errCh <- fmt.Errorf("error creating listener: %v", err)
		return
	} else {
		log.Printf("client: listener started successfully on %d", clientListenerPort)
	}
	defer listener.Close()
	

	for {
		conn, err := listener.Accept()
		if err != nil {
			_ , match := err.(net.Error) 
			if match {
				continue
			}
			errCh <- fmt.Errorf("failed to accept connection: %v", err)
		}

		log.Printf("client: recieved connection from: %v\n", conn.RemoteAddr().String())
	
		// this is NFG. why? the default, quic, does not need to be coupled to my TCP listener. hell, if its quic, we don't
		// even need to do the TCP listener setup. my intuitons are struggling here.
		// quic and tcp conns are fundamentally different. since QUIC can have multiple streams per conn.
		// once a quic conns is established, all incomming requests to the local listner need to be handled as streams
		// that just ain't so for tcp.
		// every time a request comes in, it shouldn't just create a new quic conn (like it does now)
		switch {
		case useTls:
			wg.Add(1)
			go tcp.ConnectRemoteSecure(wg, errCh, conn, 9001, serverAddr, caCertLoc)
		case useRawTcp:
			wg.Add(1)
			go tcp.ConnectRemoteUnsec(wg, errCh, 9000, conn, serverAddr)
		default:
			wg.Add(1)
			go quic.ConnectRemoteQuic(wg, errCh, 9002, serverAddr, caCertLoc, conn)
		}

	}

}