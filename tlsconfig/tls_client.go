package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func ClientTLSConfig() (*tls.Config, error){
	
	// get ca cert
	caCertLoc := os.Getenv("CA_CERT_LOC")

	if caCertLoc == "" {
		return nil, fmt.Errorf("failed to find CA cert")
	}

	caCert, err := os.ReadFile(caCertLoc)
	if err != nil{
		 return nil, fmt.Errorf("error while reading ca cert: %v", err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)
	clientConfig := tls.Config{
		RootCAs: certPool,
		ServerName: "localhost", // This ought to be a variable. added to beat SAN warning
	}
	return &clientConfig, nil
}