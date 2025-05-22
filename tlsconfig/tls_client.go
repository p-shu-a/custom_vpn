package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
)

func ClientTLSConfig(caCertLoc string) (*tls.Config, error){
	
	// get ca cert
	if caCertLoc == "" {
		log.Println("user provided no CA cert, fetching default")
		caCertLoc = os.Getenv("CA_CERT_LOC")
		log.Printf("fetched CA cert from: %v\n", caCertLoc)
	}
	
	if caCertLoc == "" {
		return nil, fmt.Errorf("failed to find CA cert")
	}

	caCert, err := os.ReadFile(caCertLoc)
	if err != nil {
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