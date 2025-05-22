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
`openssl req -x509 -nodes -key $HOME/.custom_vpn/ssl/ca/ca.key -sha256 -days 365 -out $HOME/.custom_vpn/ssl/ca/ca.pem -config ssl/ca.cnf`
- do not include the `-new` flag in the command above, creates a new x509 cert. we want a ca cert signed with the x509 standard 🥴
- Test the cert's purpose:
`openssl x509 -purpose -in $HOME/ssl/ca/ca.pem -inform PEM`

### Commands to generate the Server key, csr and signed cert
- all output goes to `$HOME/.custom_vpn/ssl/`
`openssl genrsa -out $HOME/.custom_vpn/ssl/server.key 2048`
`openssl req -new -key $HOME/.custom_vpn/ssl/server.key -config ssl/server.cnf -out $HOME/.custom_vpn/ssl/server.csr`
- verify generated CSR with:
`openssl req -in $HOME/.custom_vpn/ssl/server.csr -noout -text`

- touch index.txt and serial.txt
`touch $HOME/.custom_vpn/ssl/ca/serial.txt`
`touch $HOME/.custom_vpn/ssl/ca/index.txt`

- get the cert signed by the CA (generated in previous step):
`openssl ca -config ssl/ca.cnf -policy signing_policy -extensions signing_req -out $HOME/.custom_vpn/ssl/server.pem -infiles $HOME/.custom_vpn/ssl/server.csr`

- verify generated cert with:
`openssl x509 -in $HOME/.custom_vpn/ssl/server.pem -text -noout`

- this SO answer has a lot of details (was more useful than chatGPT):
    - https://stackoverflow.com/questions/21297139/how-do-you-sign-a-certificate-signing-request-with-your-certification-authority