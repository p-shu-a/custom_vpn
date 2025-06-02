### Setting up the certs and key
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
    - `openssl genrsa -out $HOME/.custom_vpn/ssl/ca/ca.key 2048`
    - `openssl req -x509 -nodes -key $HOME/.custom_vpn/ssl/ca/ca.key -sha256 -days 365 -out $HOME/.custom_vpn/ssl/ca/ca.pem -config ssl/ca.cnf`
- do not include the `-new` flag in the command above, creates a new x509 cert. we want a ca cert signed with the x509 standard 🥴
- Test the cert's purpose:
    - `openssl x509 -purpose -in $HOME/ssl/ca/ca.pem -inform PEM`

### Commands to generate the Server key, csr and signed cert
- all output goes to `$HOME/.custom_vpn/ssl/`
    - `openssl genrsa -out $HOME/.custom_vpn/ssl/server.key 2048`
    - `openssl req -new -key $HOME/.custom_vpn/ssl/server.key -config ssl/server.cnf -out $HOME/.custom_vpn/ssl/server.csr`
- verify generated CSR with:
    - `openssl req -in $HOME/.custom_vpn/ssl/server.csr -noout -text`

- touch index.txt and serial.txt
    - `touch $HOME/.custom_vpn/ssl/ca/serial.txt`
    - `touch $HOME/.custom_vpn/ssl/ca/index.txt`

- get the cert signed by the CA (generated in previous step):
    - `openssl ca -config ssl/ca.cnf -policy signing_policy -extensions signing_req -out $HOME/.custom_vpn/ssl/server.pem -infiles $HOME/.custom_vpn/ssl/server.csr`

- verify generated cert with:
    - `openssl x509 -in $HOME/.custom_vpn/ssl/server.pem -text -noout`

- this SO answer has a lot of details (was more useful than chatGPT):
    - https://stackoverflow.com/questions/21297139/how-do-you-sign-a-certificate-signing-request-with-your-certification-authority

- verify if a server and a ca cert match:
    - `openssl verify -CAfile ca.pem server.pem`

---
#### generating keys and certs DUMP FROM THE OLD README -- REFACTOR ME
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
- Further Reading: 
    - https://letsencrypt.org/certificates/
    - https://stackoverflow.com/questions/21297139/how-do-you-sign-a-certificate-signing-request-with-your-certification-authority