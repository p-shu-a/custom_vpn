##### update me!

### PROJECT NAME
---
### PROJECT DESCRIPTION
---
### GETTING STARTED
---
### USAGE

---
### ToDo
- add clean-restart to server logic...if server goes down, pops back up...continues service
- modularize the code (always ongoing):
    - saparate the contexts. cancel context should be saparate for a stream / conn fail
    - refactor client
- cert management could certainly use another look. 

---
#### Lets walk through the architecture
- there is a client, which listens for connections on port 2022. upon receiving a conn, it dials my server on port 9000 and 9001 for TLS
- I've got a server, listening on port 9000. 
    - upon receiving a connection, the server dials to port 22, ssh. 
    - a banner `SSH-2.0-OpenSSH_9.9` is presented to the connecting user
- right now, i connect (via nc/socat) to my client. but in a "real" vpn app, this would be done through something like a desktop app.
    - login in, click "connect" and all the dialing to server is done via the app.
