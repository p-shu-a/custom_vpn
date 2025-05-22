### wtf
- You'll need to generate:
    - private keys
    - public certs
    - crs
    - cnf

- Run the `env_setup.sh` file first

### What private keys need to be generated?
- server private key. called server.key
- ca private key. called ca.key

### What certs need to be generated?
- cert for the server. called server.pem
    - You'll need to have the cert signed by a CA
- cert for the ca. called ca.pem

### Commands to generate the CA key-pair (.key and .pem)
- You'll be generating a self-signed cert here.
- all output goes to `$HOME/.custom_vpn/ssl/ca/`
`openssl genrsa -out $HOME/.custom_vpn/ssl/ca/ca.key 2048`
`openssl req -x509 -new -nodes -key $HOME/.custom_vpn/ssl/ca/ca.key -sha256 -days 365 -out $HOME/.custom_vpn/ssl/ca/ca.pem -config ssl/ca.cnf`

### Commands to generate the Server key, csr and signed cert
- all output goes to `$HOME/.custom_vpn/ssl/`
`openssl genrsa -out $HOME/.custom_vpn/ssl/server.key 2048`
`openssl req -new -key $HOME/.custom_vpn/ssl/server.key -config ssl/server.cnf -out $HOME/.custom_vpn/ssl/server.csr`
- verify genereate CSR with:
`openssl req -in $HOME/.custom_vpn/ssl/server.csr -noout -text`
- get the cert signed by the CA (generated in previous step):
`openssl x509 -req -in $HOME/.custom_vpn/ssl/server.csr -CA $HOME/.custom_vpn/ssl/ca/ca.pem -CAkey $HOME/.custom_vpn/ssl/ca/ca.key -CAcreateserial -days 365 -sha256 -out $HOME/.custom_vpn/ssl/server.pem`
- verify generated cert with:
`openssl x509 -in $HOME/.custom_vpn/ssl/server.pem -text -noout`