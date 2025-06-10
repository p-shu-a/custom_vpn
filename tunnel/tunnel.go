package tunnel

import (
	"io"
	"log"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
	"github.com/songgao/water"
)

// Use for copy between two net.conns
func CreateTunnel(dst, src net.Conn){
	var wg sync.WaitGroup
	var once sync.Once
	closeConns := func ()  {
		dst.Close()
		src.Close()
	}

	wg.Add(1)
	go func(){
		defer wg.Done()
		io.Copy(dst, src)
		once.Do(closeConns)
	}()

	wg.Add(1)
	go func ()  {
		defer wg.Done()
		io.Copy(src, dst)
		once.Do(closeConns)
	}()
	wg.Wait()
}

// Use for copy between QUIC Stream and net.conn
func QuicTcpTunnel(conn net.Conn, stream quic.Stream){
	var wg sync.WaitGroup
	var once sync.Once
	close := func(){
		stream.Close()
		conn.Close()
	}
	
	wg.Add(1)
	go func(){
		defer wg.Done()
		io.Copy(stream, conn)
		once.Do(close)
	}()

	wg.Add(1)
	go func(){
		defer wg.Done()
		io.Copy(conn, stream)
		once.Do(close)
	}()

	wg.Wait()
}

func QuicTunTunnel(stream quic.Stream, ifce *water.Interface){
	var wg sync.WaitGroup
	var once sync.Once
	close := func(){
		log.Print("QuicTunTunnel: doing close")
		ifce.Close()
		stream.Close()
	}

	wg.Add(1)
	go func(){
		defer wg.Done()
		_, err := io.Copy(ifce, stream)
		log.Printf("QuicTunTunnel: ioCopy (ifce, str) ended: %v", err)
		once.Do(close)
	}()

	wg.Add(1)
	go func(){
		defer wg.Done()
		_, err := io.Copy(stream, ifce)
		log.Printf("QuicTunTunnel: ioCopy (str, ifce) ended: %v", err)
		once.Do(close)
	}()

	wg.Wait()
}