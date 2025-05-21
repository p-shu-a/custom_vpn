package tunnel

import (
	"io"
	"net"
	"sync"
)

// add wait group here too
func CreateTunnel(dst, src net.Conn){
	var once sync.Once
	closeConns := func ()  {
		dst.Close()
		src.Close()
	}

	go func(){
		io.Copy(dst, src)
		once.Do(closeConns)
	}()

	go func ()  {
		io.Copy(src, dst)
		once.Do(closeConns)
	}()
}
