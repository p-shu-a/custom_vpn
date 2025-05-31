### Custom VPN
- The project isn't quite a vpn, but its almost a vpn. Title is asprirational
- But the project is definitely a tunnel for securely communicating data from client to server

---
#### Why does this project exist?
- The genesis of this project was a question, "what does the world see when I use a vpn?"
- Figured it would teach me about VPN architecture, networking and allow me to explore Go in more depth
- Ultimately, idea is that you'd be able to fire this client up from anywhere, and access the server running somewhere (device at home, some box in the sky, etc)
    - And you'd be able to do so without port forwarding on your home router

##### Where we've been:
- Initially, the project was simply a TCP tunnel which conencted a client and server (both running on localhost).
    - Port 2022 on the client. 9000 on the server
    - It didn't connect anything sexy either, just dialed to port 22 on the back end. presented a ssh banner.
- Layered on some transport security. Few steps required:
    - needed to generate key, certs and related requests for CA (self-signed) and server
    - Fetch configs when the listener starts
    - server recieves TLS connections on 9003
    - Spent time inspecting traffic using wireshark
    - Wireshark is a phenominal piece of software. You can find some packet captures are in the `pcaps` dir
- Adding some graceful shutdowns to server
    - this was to make sure that, if the server died, things would still get copied to the other end-point
    - it was this point where I really started to use GO's concurreny features
        - waitgroups, contexts (cancel), channels for (error messaging) and go-routines
        - I find the blocking behaviour of channels just sensational. pairing it with go-routines leads to some very cool possibilities
    - Even at this point, the project still just connected to ssh on the back end. it is using TLS now.
- Moved server to Raspberry Pi
    - Finally decided to put the server on a server
    - Server Specs
        - RaspberryPi (Model-4b)
        - Ubuntu server. 24.04.02 LST
        - 228GiB SSD replaced the SD card storage. ChatGPT was in dispensible for getting all this setup.
    - had to regen the certs again. properly this time
        - The CA cert had to be handled a bit differently. Had to explicity be declared the CA cert
        - encounterd an issue where my server cert didn't mention the SAN (ip or domain), thus the TLS handshake was failing
        - added a shell script to mod the ~/.bashrc to export env vars and create dirs for the cert and keys to reside in
    - Required updates to the client to accept config options via flags

##### Where we are:
- The project now implements communication via QUIC. We're using the `github.com/quic-go/quic-go` library
- At this point:
    - the server (recieving on port 9002) sends the client a datagram
    - the client opens a stream and send the server an acknoledgement hello
    - This is multiplexing. A connections and some stream in it.
- See `internal/quic/README.MD` for details
- Spent some time modularizing and re-organizing the code
- Immediately working on:
    - adding a header per stream to direct to a particular endpoint on the server
    - adding a tunnel for generic data
    - adding the `https://github.com/p-shu-a/jwt-auth` project as an auth service

##### Where we're going:
- External access between peers is pretty important
- Have multiple peers
- Most likely will implement wireguard. a vpn shouldn't be a TCP tunnel
- Could make the auth service more robust and useful
- Get certs from LetsEncrypt.

---
### GETTING STARTED
- It is assumed that you have go installed. This project was build using Go v.1.24.2
- You'd have to do a few things:
    - run the `ssh/env_setup.sh`
    - Generate the cert and keys for the CA & Server. Detailed instructions can be found in `ssl/README.MD`
        - a `server.cnf` and `ca.cnf` have been provided for you convenience
    - This is a manual process, you have to make sure:
        - the right dirs are exported (feel free to make changes to the locations)
        - the certs generated are placed in the right location


---
### USAGE... in progress
- use socat to connect to the client port
- 
- 
---