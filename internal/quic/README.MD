### QUIC Shift

- Using `github.com/quic-go/quic-go`
- what are some must haves in a quic config?
- One recurring issue I keep encountering is, how do i know what parent interface another interface is composed of?
- with my server, i'm using the same ctx every where. might not be a good idea. should dig deeper
    - one potential error is that if a stream calls a context done, my whole server would exit...no?
    - whats best practice for scoping contexts?
    - update: definitly should not be using the same context
- i'm also using the same waitgroup. whats the limit for wait group's size?

- the default number of streams per conn is 100. does that mean, the same conn can open multiple, to pick and example, http streams?
    - what if we limited the number of streams to just 2?
    - how do you differentiate streams?
- if one of the streams is closed, does it stay closed for the duration of the connection?
- whats the difference of the nature of the contents of a quic conf struct and a transport struct? 

---

so, when i multiplex over quic, i have to:
- handle each incoming stream individually.
- by default the quic-go lib allows for 100 streams per connection. 
    - i guess you can say that is the "bandwidth." 
    - in reality, the number of streams a server receives depends on how many were opened by the client.
        - If the client only opens two streams, then thats all we'll have to deal with
- each stream must be identified, and redirected based on the frame headers
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

### ToDo

### learnings

### questions

- what is 0-rtt and 0.5-rtt. what should i use them?


