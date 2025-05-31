### ToDo- project level
- add clean-restart to server logic...if server goes down, pops back up...continues service
- modularize the code (always ongoing):
    - saparate the contexts. cancel context should be saparate for a stream / conn fail. implement proper context chaning
- cert management could certainly use another look.
# CHRIST, ADD SOME TESTS ALREADY

#### ToDo- quic
- do something meaning full with QUIC
- add headers to opened streams
- adding headers to streams. Type, IP, and Port
- two possible directions:
    - specific streams. 
    - generic tunnel. how do we figure out the protocol employed by the underlying? (for the header type)
- get jwt_auth working
- THERE IS AN ISSUE WITH THE TLS CERT FETCHING LOGIC
- properly impement connetion id
    - CONNDB SHOULDN'T BE JUST A MAP
- implement connection resumption.
    - if the client changes ip, make them do a path_challenge
- ERROR HANDLING IN CONN, STREAMS ACCEPT LOOPS NEED ATTENTION


### ToDO- client
- there should be a client level errCH
    - clean up error messages in ageneral
- client has been long neglected. needs attention. especially how listeners are started.

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


