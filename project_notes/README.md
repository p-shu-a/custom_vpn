#### Contexts
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

#### Channels
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


---
### Adding support for QUIC
- Using `github.com/quic-go/quic-go`
- Partly inspired by the fact that a VPN shouldn't be using a TCP tunnel
- QUIC seemed quite feature rich, all the pieces needed for TLS and all that
- TLS configs are shared from the old TLS for TCP setup

---
#### Learnings and question
- Use nmap to query the server and you'll see some very intersting behaviour
- use `nmap -sV -p <port> <host>`
- could also do `nmap -sV <host>` for it to fast scan most common ports
-  `-sV` probes open ports and service
- you get output for port 9000 (vanilla tcp) and 9001 (tls over tcp)
- however, you don't see anything for 9002 which is the UDP port for QUIC. why is that?
    - well, quic enforces a strict handshake. which is required before any data is sent back and forth
- TLS over TCP is quite lax, compared to QUIC.
    - a service which implements TLS will still send back a banner
    - which made me wonder, namp doesn't have access to my cert. however, it still showed my banner from the TLS port. why?
    - apparently, the handshaking isn't as strict. the server sends its cert, and the client has no obligation to send anything back.
    - a service like nmap can report details.

- the default number of streams per conn is 100. does that mean, the same conn can open multiple, to pick and example, http streams?
    - what if we limited the number of streams to just 2?
    - how do you differentiate streams?
- if one of the streams is closed, does it stay closed for the duration of the connection?
- whats the difference of the nature of the contents of a quic conf struct and a transport struct? 

- so, when i multiplex over quic, i have to handle each incoming stream individually.
- by default the quic-go lib allows for 100 streams per connection. 
    - i guess you can say that is the "bandwidth." 
    - in reality, the number of streams a server receives depends on how many were opened by the client.
        - If the client only opens two streams, then thats all we'll have to deal with
- each stream must be identified, and redirected based on the frame headers. 
    - Client adds headers, server reads 'em
    - Use something like: 
        - 1byte : proto
        - 8bytes: ip
        - 4bytes: port
- if i have a stream dedicated to some HTTP service, which itself utilizes TLS over TCP.
    - I think what would still work over QUIC. 
    - Quic is agnostic to stream content
        - once the stream is opened and redirected to endpoint, quic's role becomes about managing transmission. 
    - In effect what i'm doing here is TCP over QUIC.
- what happens if the packets containing the stream header are dropped? aren't they are important for redirection, wouldn't that cause HoL type issues?
    - sort of. It could cause a HoL issue in that on particular stream, but not across streams
    - also, quic handles the retransmission for you
- there is a limit to QUIC packet size, but quic handles packet sizing for you
- A quic connection can handle datagrams AND streams over the same connection.
    - datagrams are great for low latency, fire-and-forget purposes. ddos?

### questions to research
- what are some must haves in a quic config?
- what is 0-rtt and 0.5-rtt. why/when should i use them?
- specifics of a quic handshake
- is there a limit to the wait group size?