### ToDo- project level
- add clean-restart to server logic...if server goes down, pops back up...continues service
# CHRIST, ADD SOME TESTS ALREADY
---
#### ToDo- quic
- add headers to opened streams
    - adding headers to streams. Type, IP, and Port
- get jwt_auth working (sort of working, hardwired for now)
- implement connection resumption.
    - if the client changes ip, make them do a path_challenge

---
### ToDo- client
- the logic for which protocol to use, should be decoubled from the listener starter
    - the client needs to expose multiple different ports for the different services (http, ssh, etc)

---
### ToDo- Tunnel
- connsolidate the two different tunnels... it VERY similar code, see if you can make the accepted params generic enough that one func can do both
    - net.conn to net.conn
    - net.conn to quic.stream
---
### questions to research
- what are some must haves in a quic config?
- what is 0-rtt and 0.5-rtt. why/when should i use them?
- specifics of a quic handshake
- is there a limit to the wait group size? 