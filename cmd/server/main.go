package main

import (
	"log"
	"sync"
	
	"custom_vpn/internal/helpers"
	"custom_vpn/internal/tcp"
	"custom_vpn/internal/quic"
)


func main(){

	// The returned returned context is a WithCancel() context
	// Its purpose it to shutdown the entire server upon a closing signal
	// It should not be used as the context per-stream or per-connection...i'd think
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

	// since we call both servers in go-routines
	wg.Add(1)
	go tcp.ListenAndServeNoTLS(cancelCtx, errCh, &wg, 9000)

	wg.Add(1)
	go tcp.ListenAndServeWithTLS(cancelCtx, errCh, &wg, 9001)

	wg.Add(1)
	go quic.QuicServer(cancelCtx, errCh, &wg, 9002)

	wg.Wait()
	close(errCh)
	/*
		Wait for the done channel to close (after the errorCollector go-routine is closed)
		This way, the errors eminating from the server functions are all logged in order(!) before we exit
	*/
	<-done
	log.Println("server: All servers closed. Exiting...")
}

