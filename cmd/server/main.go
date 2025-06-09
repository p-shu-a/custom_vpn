package main

import (
	"custom_vpn/config"
	"custom_vpn/internal/helpers"
	"custom_vpn/internal/quic"
	"custom_vpn/internal/tcp"
	"log"
	"sync"
)


func main(){

	// The returned returned context is a WithCancel() context
	// Its purpose it to shutdown the entire server upon a closing signal
	cancelCtx := helpers.SetupShutdownHelper()
	var wg sync.WaitGroup
	
	/* 
		Shifted strategy: since we have go-routines called by go-routines, this leads to a bastardized mix of error handling.
		Some places I was returning errors, others I was writing to an error channel. nothing worse than a mix.
		So, I'll be writing all errors to a channel.
		(using errgroup package is another approach, but not gonna look into that righ now...)
	*/
	errCh := make(chan error)
	done := make(chan struct{})		// this done channel was created to ensure the ordering of the logs
	go helpers.ErrorCollector(errCh, done)

	wg.Add(1)
	go tcp.ListenAndServeNoTLS(cancelCtx, errCh, &wg, config.RawTcpServerPort, config.SSHEndpointService)

	wg.Add(1)
	go tcp.ListenAndServeWithTLS(cancelCtx, errCh, &wg, config.TcpTlsServerPort, config.SSHEndpointService)

	wg.Add(1)
	go quic.QuicServer(cancelCtx, errCh, &wg, config.QuicServerPort)

	wg.Wait()
	close(errCh)
	/*
		Wait for the done channel to close (after the errorCollector go-routine is closed)
		This way, the errors eminating from the server functions are all logged in order(!) before we exit
	*/
	<-done
	log.Println("server: All servers closed. Exiting...")
}

