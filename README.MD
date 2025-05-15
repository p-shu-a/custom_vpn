UDP vs TCP
- differences?
- uses?

- what is NORDVPN's architecture?

- to start sshd on macos, you gotta enable remote login. do this via settings panel or :
sudo systemsetup -setremotelogin on

user@user ~ $ lsof -Pn -i6 | grep "^server\|^client"
server    30645 user    3u  IPv6 0xf18cdf112a105407      0t0  TCP *:9000 (LISTEN)
server    30645 user    5u  IPv6 0x677e1030533d680b      0t0  TCP [::1]:9000->[::1]:54632 (ESTABLISHED)
server    30645 user    6u  IPv6 0x489670b6cc14f73b      0t0  TCP [::1]:54633->[::1]:22 (ESTABLISHED)
client    30646 user    3u  IPv6 0xc37b1a3b2f44f182      0t0  TCP *:2022 (LISTEN)
client    30646 user    5u  IPv6 0x5a5707c7b5ce2824      0t0  TCP [::1]:2022->[::1]:54631 (ESTABLISHED)
client    30646 user    6u  IPv6 0xec586c6d9524a865      0t0  TCP [::1]:54632->[::1]:9000 (ESTABLISHED)

#### Lets walk through the architecture.
- there is a client, which listens for connections on port 2022. upon receiving a conn, it dials my server on port 9000.
- I've got a server, listening on port 9000. upon receiving a connection, the server dials to port 22, ssh. a banner `SSH-2.0-OpenSSH_9.9` is presented to the connecting user
- right now, i connect (via nc/socat) to my client. but in a "real" vpn app, this would be done through something like a desktop app.
    - login in, click "connect" and all the dialing to server is done via the app.


#### Useful commands:

`sudo tcpdump -i lo0 -w session.pcap tcp port 2022`: monitor action on port 2022, on the specified interface (lo0), use ifconfig to get list of interfaces
`lsof -Pn -i[46]`: list all files, iwhtout port names or host names. files should be ipv4[or 6, but not both]
`netstat`: is also useful, but its not the same as linux


### TLS
- TLS is a protocol used to provide auth between parties, and encryption and integrity of data shared between them.
- in TLS, for the handshake between two parites, asymetric encryption is used
- for the encrypt/decrypt-ing, and maintaining integrity (to detect/prohibit tampering) they use symetric encryption
    - Diffie-helman is a way to generate a shared symetric key over public channel
- There are mutual and one-sided TLS handshakes. For one-sided:
    - a client connects to a server
    - server sends its public cert
    - the client verifies the server cert againt the CA's cert
        - CA certs are in a device's trust store
        - if a cert is self-signed, that can also be detected (Issuer == Subject)
    - Client and server do Diffie-Helman to a symetric key
    - With a mutual TLS, both client and server, verifiy each other.
- There are many Components to a TLS handshake. Wikipage is decent, goodish diagrams

#### generating keys and certs
- a key is private
    - generate a key by doing:
        - `openssl genrsa -out homie.key 2048`
- a cert is public facing
    - there are self signed certs, and certs issues by a CA, certificate authority, (like letsencrypt)
    - generate a cert by doing:
        - `openssl req -x509 -new -nodes -key homie.key -sha256 -days 365 -out homie.pem`
        - above we created a self-signed cert. if you want to get a cert from a CA, you generate a CSR...
- A CSR is a request to sign a certificate.
    - lets look at the command to generate a CSR first
        - `openssl req -new -key homie.key -out homie.csr`
    - the requesting entity's private key is used to generate a CSR
        - the private key is used to "prove ownership of a corresponding public key"
            - which means, the public key is derived from the private key
        - a CSR contains the CSR issuing entity's public key
        - terminology matters. a csr is "generated", not "signed"
    - you send this request to a CA.
    - The CA verifies the CSR issuing entity's identity
    - CA sends back a signed cert. this cert is signed using the ca's private key
        - the signed cert is a verification of the entity's identity
- Command to generate a signed certificate for a CSR:
    - `openssl x509 -req -in homie.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out homie.pem -days 365 -sha256`
    - the `-CAcreateserial` generats a .srl file which contains a serial number to used in the cert signing
- NOTE: openssl is sensitive to the positioning of flags in the command:
    - `openssl x509 -req...` vs `openssl req -x509...`
    - the `-help` flag for openssl is pretty decent
- If you want to see the metadata of a cert, do:
    `openssl x509 -in homie.pem -text -noout`
- Further Reading: https://letsencrypt.org/certificates/


#### Converting simple TCP tunnel to TLS Tunnel
- to setup a TLS tunnel, we need to specify TLS configs for both the server and client ends
    - we need specify a tls.Config{} struct for both








### Questions:
- whats a context?
- how does go handle go-routines which spawn go-routines?