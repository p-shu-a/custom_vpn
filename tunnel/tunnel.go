package tunnel

import (
	"io"
	"net"
	"sync"
)

func CreateTunnel(dst, src net.Conn){
	var wg sync.WaitGroup
	var once sync.Once
	closeConns := func ()  {
		dst.Close()
		src.Close()
	}

	// we add wait groups here so writes can finish to the opposite end even when TERMs happen
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
