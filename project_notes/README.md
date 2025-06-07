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

### Issues with QUIC streams
- My Plan was to open a different port for each service on the client device. Then have the user make requests to the different ports.
    - That is: client opens port 2023 for ssh , 2022 for http. if user is making http request, they send request to 2022
    - requests to each different port is sent to the server in a distinct stream
- There are few problems with this:
    - as you add support for more protocols, you have to setup listeners for each port. this is cumbersome. 
        - multiple similar function calls where the only diff is a different port
    - my only form of "validatation" is that I relate a port to a specific protocol. this isn't a guarantee of anything...
    - a user making a request to a port doesn't actually mean they are sending whatever type of request I expect on that port
        - ssh port 2023 could recieve HTTP data
    - This is bad UX. User has to keep track of each different port. gonna lead to mistakes
    - Right now, stream header is compose of Protocol(4b), IP(16b), Port(2b)
        - but the IP and Port values are for the server, not for the service endpoint. Their inclusion is effectively useless
        - what i need is to send the URL+Port of the endpoint service. for most of my cases, the url would stay the same since the server is running the services, but the ports are deffs different. and if we're forwarding requests beyond the server, then the IP is deffs uselless
        - I could read the request, but that is a privacy violation. and it wouldn't be possible if even the incoming request was encrypted
    - Overall, this jsut feels like a lot of work to forward requests.
- If only there was some way to forward all traffic through? HMMMM
    - enter TUN!
    - i'm not sure about the specifics, but i assume that all my data, or just data from a few select ports, or even just data headed to a specific remote address, can be transported through the new interface

### TUN
- with TUN interfaces, i'd be able to forward all (or select) data from my client machine to the server
    - state has to be managed:
        - my client has to setup TUN and route to the server
        - the server has to setup TUN and route to client. server has the additional responsibility of sending packets forward.
    - this data, going between client and server, will be encrypted. not sure how
- what my server should do is look at the packets and see what the intended end destination is, and forward data.
    - infact, this can't really happen if I use QUIC. i have to steer the packet to the TUN
#### here is how i think it'll go:
- my client creates a TUN, and adds a route saying "forward traffic going to X subnets/my server through this TUN". this makes merealize, the more selective I want to about steering data to specifc outgoing subnets, the more routes i will need to add.
- my server have to create a interface too, but to receive. basically the other end of the pipe.
    - but does the server need to add any routes? yes, back to the client.
- this needs pipe needs to be encrypted. how? is the pipe encrypted or the data within?
    - where does the app sit, in fornt of the TUN or behind?
    - so the app sits behind the TUN
        - at that point, i think, the server could just recieve on its default interface
        - if the server didn't have the respoinsibility to send data back to the client, it would not need to create a new interface and issue route commands.
- the data inside, regardless of it using TLS, gets encrypted at the client-end. 
- it is received at the server, and decrypted. 
- the server reads the what the end destination address on the packet is. so far this is the only time the destination address on thepacket has be read by the app.
- server forwards the data. 
- process repeats backwards
- use `github.com/songgao/water` for creating TUNs. not sure if there is a better lib...

### questions to research
- what are some must haves in a quic config?
- what is 0-rtt and 0.5-rtt. why/when should i use them?
- specifics of a quic handshake
- is there a limit to the wait group size?