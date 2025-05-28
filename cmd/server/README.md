### Adding Contexts
- What is a context for?
    - a context carries deadlines, cancellation signals and metadata across go-routines
    - specifically, you can catch terminations and gracefully shutdown/clean up a go-routine.

- Contexts employ select{}. Select is like a switch statement, except:
    - there is not conditional
    - it blocks until one of the case clauses can be executed
    - re:blocking, thats why select{} blocks indefinitly. though, if you use a default, it executes that operation.

- When do you use a context? do you create a context or accept it?
    - worker/child/called functions, should accept a context
    - top-level functions should create context
    - eg. If main() calls foo(), then main() should create context, and foo() should accept it

- What are some core context functions?
    - context.Background() is used to creating a context
    - Context is an interface with the following functions:
        - Done(), Deadline(), Err(), Value()
    - You can also create "derived contexts", which are created with parent contexts.
        - Four: WithCancel(), WithTimeout(), WithDeadline(), WithValue()
        - `context.WithCancel(parentContext)` this returns a context and a CancelFunc
            - the cancelFunc must be called to release resources.
        - Derived contexts can be layered. A context can have a timeout, a value, deadline, etc

- A Note on `WithCancel(Context) (Context, CancelFunc)`
    - while every context has a .Done(), it is execute by calling the cancelFunc which is returned by WithCancel
    - without `WithCancel()` you never get back a cancelfunc which can be called.
    - ctx.Done() returns a channel. 
        - in the context interface its defined as `Done() <-chan struct{}`
        - what does that mean? well...s

### Channels
- Instantiate a channel using `ch := make(chan Type)`
- two types (sort of...):
    - bi-directional: can send to and recieve from.
    - unidirectional: can either only send or recieve. not both.
- Every channel created is bi-directional. But you constrain its usage (say in a function) to only send or recieve from that channel
- when you see `Done() <-chan struct{}` more clearly written as `Done() (<-chan struct{})`
    - its saying "Done() returns a receive only channel of type struct{}"
    - we pass an empty struct to just send a signal, since it takes-up no memory.
- Re:Send-only and Receive-only, a better description of each would be:
    - send-only     : "you can only SEND TO that channel"
    - recieve-only  : "you can only RECIEVE FROM that channel"
- The `<-` notation, and its placement, indicates the direction of operation:
    - `ch <- 5` sends to channel
    - `foo := <-ch` recieves from that channel
- By default channels are unbuffered, that is they only accept a send if a recieve is ready.
    - Channels are blocking
    - they force syncronization
    - VERY interesting behaviour!
- Buffered Channles on the other hand, allow for values to accomulate and then consumed as needed.
- We want channels because they allow for communication and coordination between go-routines
    - go routines can't return values. thus, they must use channels 


### Managing Shutdowns and Terminations
- ------