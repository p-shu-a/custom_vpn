package tunnel

import (
	"io"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
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