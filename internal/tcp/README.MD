#### Converting simple TCP tunnel to TLS Tunnel
- to setup a TLS tunnel, we need to specify TLS configs for both the server and client ends
    - we need specify a tls.Config{} struct for both

### TCP handshake
- TCP handshake is a way of extablishing connection from client to server
- C->S: Syn, S->C: syn-ack, C->S: ack
- then onto a TLS handshake



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