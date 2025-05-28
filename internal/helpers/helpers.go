package helper

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

/*
	whats this puppy for?

*/

/*
	We're interesting in "interface satisfaction"
	Perhaps its not accurate to use the "parent-child" terminology here, but indulge me.
	net.Listener is a Closeable interface. Why? because it implements the Close() method.
	And since our definition of a closeablelistener is "a closeable interface implements as close() method" and
	net.Listener implments close(), thus it satisfies.
	Very Cool
*/
type CloseableListener interface {
	Close() error
}

/* 
	This function closes a listener, doesn't matter if its a TCP listener or a QUIC listener.
	That is why we defined the CloseableListener interface.
*/
func CaptureCancel(wg *sync.WaitGroup, ctx context.Context, errCh chan<- error, port int, listener CloseableListener){
	defer wg.Done()
	<-ctx.Done()			// block here until cancel()
	listener.Close()		// call our closeable listeners close() function
	errCh <-fmt.Errorf("%v: listener closed on port-%d due to SIGTERM", ctx.Err(), port)
}


/*
	Helper method to return a context which will track shutdown/terms
	Basically, we're setting up the signal handling.
	Create a channel which gets notified when SIGINT or SIGTERM are called
	When a signal is picked up, we unblock and call the cancel(), cancelling the context
	Cancelling the context leads to ctx.Done() being called in other places (the server funcs)
*/
func SetupShutdownHelper() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // block here until there is something is recieved
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