#### Lets walk through the architecture.
- there is a client, which listens for connections on port 2022. upon receiving a conn, it dials my server on port 9000 and 9001 for TLS
- I've got a server, listening on port 9000. 
    - upon receiving a connection, the server dials to port 22, ssh. 
    - a banner `SSH-2.0-OpenSSH_9.9` is presented to the connecting user
- right now, i connect (via nc/socat) to my client. but in a "real" vpn app, this would be done through something like a desktop app.
    - login in, click "connect" and all the dialing to server is done via the app.

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
        - `openssl req -new -key homie.key -out homie.csr` add a `-config config_file.cnf` if you have a config file
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
- CN is deprecated, gotta use Subject Alternate Names
- Further Reading: https://letsencrypt.org/certificates/


#### Converting simple TCP tunnel to TLS Tunnel
- to setup a TLS tunnel, we need to specify TLS configs for both the server and client ends
    - we need specify a tls.Config{} struct for both

### TCP handshake
- TCP handshake is a way of extablishing connection from client to server
- C->S: Syn, S->C: syn-ack, C->S: ack
- then onto a TLS handshake

### ToDo:

- add clean-restart to server logic...if server goes down, pops back up...continues service
- implement flags. what are some decent flags to have?
    - flags to specify certs?
- modularize the code
- export cert locations to env-vars.

- how can i specify multiple different names in altnames of my cert?
    - do he names have to be ips, or domains or can they be just a bunch of baloney?
    - localhost is one. can "Tony's fine servers" be another?
