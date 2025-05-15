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
	serverKeyLoc := "../priv_keys/server.key"
	serverPemLoc := "../certs/server.pem"
	serverCert, err := tls.LoadX509KeyPair(serverPemLoc,serverKeyLoc)
	if err != nil {
		return nil, fmt.Errorf("tls_server: fatal error loading KeyPair")
	}
	serverConfig := tls.Config{
		MinVersion: tls.VersionTLS13,
		Certificates: []tls.Certificate{serverCert}, // this key takes a slice of certs
		Time: nil,
	}

	return &serverConfig, nil
}