### ToDo- project level
- add clean-restart to server logic...if server goes down, pops back up...continues service
# CHRIST, ADD SOME TESTS ALREADY
---
#### ToDo- quic
- implement connection resumption.
    - if the client changes ip, make them do a path_challenge
- quic header logic needs a refactor:
    - other than the protocol, the backend service's destination needs to be identified
- Timeouts behaviour needs to be configured:
    - initially, you get to enjoy the bidirectional stream. after some time of inactivity, the timeout happens, and future requests fail
    - you should be able to keep sending value until one side is closed.
    - once a timeout does happen, the UDP port on the client should be come free to recieve conections again. right now, the client requires a restart
---
### ToDo- client
- the logic for which protocol to use to talk to server, should be decoubled from the listener starter
    - the client needs to expose multiple different ports for the different services (http, ssh, etc)
- maybe drop support for TCP altogether.

---
### ToDo- Tunnel
- connsolidate the two different tunnels... it VERY similar code, see if you can make the accepted params generic enough that one func can do both
    - net.conn to net.conn
    - net.conn to quic.stream
