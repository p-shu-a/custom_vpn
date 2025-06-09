package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"os"
)

// Returns a TLS config for client
// user can provide a CA certificate location
// Default is retreived from an env-var
func ClientTLSConfig(caCertLoc string) (*tls.Config, error){
	
	if caCertLoc == "" {
		caCertLoc = os.Getenv("CA_CERT_LOC")
	}
	
	caCert, err := os.ReadFile(caCertLoc)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)
	clientConfig := tls.Config{
		RootCAs: certPool,
		ServerName: "localhost", // added to beat SAN warning
		// ClientSessionChache allows for TLS session resumption
		// ClientSessionCache: tls.NewLRUClientSessionCache(100),
	}
	return &clientConfig, nil
}