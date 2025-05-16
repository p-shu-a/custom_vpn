package tlsconfig

import (
	"crypto/tls"
	"fmt"
)

/*
	- initially figured that pem, and key would be params.
	- but since we don't have multiple server certs, decided against it.
*/
func ServerTLSConfig() (*tls.Config, error) {
	serverKeyLoc := "/Users/pranshu/git/custom_vpn/priv_keys/server.key"
	serverPemLoc := "/Users/pranshu/git/custom_vpn/certs/server.pem"
	serverCert, err := tls.LoadX509KeyPair(serverPemLoc,serverKeyLoc)
	if err != nil {
		return nil, fmt.Errorf("tls_server: fatal error loading KeyPair: %v", err)
	}
	serverConfig := tls.Config{
		MinVersion: tls.VersionTLS13,
		// this key takes a slice of certs, you'd add more here if there were multiple domains. smart enough to do SNI for you
		Certificates: []tls.Certificate{serverCert},
		Time: nil,
	}

	return &serverConfig, nil
}