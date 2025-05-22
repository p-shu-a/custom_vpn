### What goes here?
- This folder should hold your public certs, crs, cnf and private keys (*gasp*)

### What certs need to be generated?
- you need a cert for the server (and a matching private key)
- You'll need to have the cert signed by a CA

### What private keys need to be generated?
- You'll need a server private key (and a matching server cert). it should be called server.key


export SERVER_PEM=/Users/pranshu/git/custom_vpn/certs/server.pem
export SERVER_KEY=/Users/pranshu/git/custom_vpn/priv_keys/server.key
export CA_CERT_LOC=/Users/pranshu/git/custom_vpn/certs/ca.pem