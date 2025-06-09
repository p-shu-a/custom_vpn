package helpers

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

/*
	Here are some shared functions and an interface.
	These functions are called across the QUIC and TCP package
*/

/*
	Since our definition of a CloseableListener is "an interface which implements a close() method",
	and both net.Listener and quic.listener have a close(), they are, thus, CloseableListeners
	This is "interface satisfaction". Very Cool
*/
type CloseableListener interface {
	Close() error
}

/* 
	This function blocks, waiting for a cancel signal. Upon receving a signal, it closes the passed listener
	Doesn't matter if its a TCP listener or a QUIC listener. See CloseableListener interface.
	port is the port which the lister is bound to
*/
func CaptureCancel(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, port int, listener CloseableListener){
	defer wg.Done()
	<-ctx.Done()			// block here until cancel()
	listener.Close()		// call our closeable listeners close() function
	errCh <-fmt.Errorf("listener closed on port-%d due to SIGTERM", port)
}

/*
	Helper method to return a context which will track shutdown/terminations.
	Basically, we're setting up the exit signal handling.
	The func creates a channel which gets notified when SIGINT or SIGTERM are called
	When a signal is picked up, we unblock and call the cancel() on the server-level context.
	Cancelling the context leads to ctx.Done() being called in other function. See CaptureCancel()
*/
func SetupShutdownHelper() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // block here until a SIGINT or SIGTERM is recieved
		cancel()
	}()
	return ctx
}


/*
	Sole purpore of this collector func is to print errors as they come down the err channel
*/
func ErrorCollector(errCh <-chan error, done chan struct{}){
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

/*  
	Refers to the connection ID for a QUIC connection. 
*/
type CtxKey string
const ConnId CtxKey = "ConnId"


/*
	Returns a UUID.
	RFC says uuid is 128bits (16*8) long.
*/
func GenUUID() (string, error){
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil{
		return "", err
	}
	// why do this bitmasking business?
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// header should be: Type (4byte), IP (4bytes for ipv4, 16bytes for ipv6), and Port (1byte)
type StreamHeader struct{
	Proto [4]byte
	IP 	  net.IP
	Port  uint16   // port could be anywhere from 0-5digits long int. uint16 (16bit, positive ints) work perfectly since ports only get up to 65535
}