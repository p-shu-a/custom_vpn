package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"custom_vpn/config"
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

	clientListenerPort := flag.Int("p", config.ClientListnerPort, "Port used to connect to client (via socat, postman, ssh, etc.)")
	remoteServerAddress := flag.String("addr", "127.0.0.1", "Server IP Address")
	mode := flag.String("mode", "quic", "Connection mode. options are: \"tcp\", \"tls\", and \"quic\"")
	caCertLoc := flag.String("ca", "", "specify a custom CA cert")
	flag.Parse()

	errCh := make(chan error, 1)
	done := make(chan struct{})
	go helpers.ErrorCollector(errCh, done)

	ctx := helpers.SetupShutdownHelper()
	var wg sync.WaitGroup

	wg.Add(1)
	// add local listener calls to multiple ports here
	// also add context to start listner, just like with server, to kill client if sigterm is sent
	go startLocalListener(ctx, errCh, &wg, *clientListenerPort, *remoteServerAddress, *caCertLoc, *mode)

	wg.Wait()
	close(errCh)

	<-done
	log.Print("all listeners stopped...exiting client")
}


/*
	Creates a tcp net.conn on the clientListnerPort
	The user can establish multiple connections to this port. but why?
	based on user's selection of TLS or not (transSec)
	tries to establish remote conn with or without tls
	We need to be able to start multiple listeners (HTTP, SSH, etc...)
*/
func startLocalListener(ctx context.Context, errCh chan<-error, wg *sync.WaitGroup, clientListenerPort int, remoteServerAddr, caCertLoc string, mode string) {
	defer wg.Done()

	localAddr := net.TCPAddr{
		IP: net.ParseIP("0.0.0.0"),
		Port: clientListenerPort,
	}

	// Start a local listener...what if this was UDP?
	localListener, err := net.Listen("tcp", localAddr.String())
	if err != nil{
		errCh <- fmt.Errorf("error creating listener: %v", err)
		return
	} else {
		log.Printf("Client: listener started on %d", localAddr.Port)
	}
	defer localListener.Close()
	
	wg.Add(1)
	go helpers.CaptureCancel(ctx, wg, errCh, localAddr.Port, localListener)

	for {
		conn, err := localListener.Accept()
		if err != nil {
			//_ , match := err.(net.Error) 
			if errors.Is(err, net.ErrClosed){
				errCh <- err
				return
			}
			continue
		}

		log.Printf("client: recieved client request from: %v\n", conn.RemoteAddr().String())
	
		switch mode{
		case "tls":
			remoteAddr := net.TCPAddr{
				IP: net.ParseIP(remoteServerAddr),
				Port: config.TcpTlsServerPort,
			}
			wg.Add(1)
			go tcp.ConnectRemoteSecure(wg, errCh, conn, caCertLoc, &remoteAddr)
		case "tcp":
			remoteAddr := net.TCPAddr{
				IP: net.ParseIP(remoteServerAddr),
				Port: config.RawTcpServerPort,
			}
			wg.Add(1)
			go tcp.ConnectRemoteUnsec(wg, errCh, conn, &remoteAddr)
		default:
			wg.Add(1)
			remoteAddr := net.UDPAddr{
				IP: net.ParseIP(remoteServerAddr),
				Port: config.QuicServerPort,
			}
			go quic.ConnectRemoteQuic(ctx, wg, errCh, &remoteAddr, caCertLoc, conn)
		}
	}
}