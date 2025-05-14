UDP vs TCP
- differences?
- uses?

- what is NORDVPN's architecture?

- to start sshd on macos, you gotta enable remote login. do this via settings panel or :
sudo systemsetup -setremotelogin on

- how does go handle go-routines which spawn go-routines?

user@user ~ $ lsof -Pn -i6 | grep "^server\|^client"
server    30645 user    3u  IPv6 0xf18cdf112a105407      0t0  TCP *:9000 (LISTEN)
server    30645 user    5u  IPv6 0x677e1030533d680b      0t0  TCP [::1]:9000->[::1]:54632 (ESTABLISHED)
server    30645 user    6u  IPv6 0x489670b6cc14f73b      0t0  TCP [::1]:54633->[::1]:22 (ESTABLISHED)
client    30646 user    3u  IPv6 0xc37b1a3b2f44f182      0t0  TCP *:2022 (LISTEN)
client    30646 user    5u  IPv6 0x5a5707c7b5ce2824      0t0  TCP [::1]:2022->[::1]:54631 (ESTABLISHED)
client    30646 user    6u  IPv6 0xec586c6d9524a865      0t0  TCP [::1]:54632->[::1]:9000 (ESTABLISHED)

#### Lets walk through the architecture.
- I've got a server, listening on port 9000. upon receiving a connection, the server dials to port 22, ssh. why ssh? not sure, perhaps this will become clear as the project progresses.
- there is a client, which listens for connections on port 2022. upon receiving a conn, it dials my server on port 9000. 
- right now, i connect (via nc) to my client. but in a "real" vpn app, this would be done through something like a desktop app. login in, click "connect" and all the dialing to server is done via the app. - in my current, simple architecure, I could nc straight to the server, but we're trying to emulate the UX of the real deal. 


Useful commands:

`sudo tcpdump -i lo -w session.pcap tcp port 2022`: monitor action on port 2022, on the specified interface (lo)
`lsof -Pn -i[46]`: list all files, iwhtout port names or host names. files should be ipv4[or 6, but not both]
`netstat`: is also useful, but its not the same as linux