### QUIC Shift
- Using `github.com/quic-go/quic-go`
- Partly inspired by the fact that a VPN shouldn't be using a TCP tunnel
- QUIC seemed quite feature rich, all the pieces needed for TLS and all that
- TLS configs are shared from the old TLS for TCP setup


#### ToDo
- do something meaning full with QUIC
- add headers to opened streams
- get jwt_auth working
- adding headers to streams. Type, IP, and Port
- two possible directions:
    - specific streams. 
    - generic tunnel. how do we figure out the protocol employed by the underlying? (for the header type)
- The context chain needs to be look at
- THERE IS AN ISSUE WITH THE TLS CERT FETCHING LOGIC
- properly impement connetion id
- implement connection resumption.
    - if the client changes ip, make them do a path_challenge

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



### questions
- what are some must haves in a quic config?
- what is 0-rtt and 0.5-rtt. why/when should i use them?
- specifics of a quic handshake
- is there a limit to the wait group size? 

- With my server, i'm using the same ctx every where. might not be a good idea. should dig deeper
    - one potential error is that if a stream calls a context done, my whole server would exit...no?
    - whats best practice for chaining contexts?
    - update: definitly should not be using the same context
- i'm also using the same waitgroup. should the wait groups be scoped differntly? per connection? server wide?


