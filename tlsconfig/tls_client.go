package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func ClientTLSConfig() (*tls.Config, error){
	// get ca cert
	caCert, err := os.ReadFile("/Users/pranshu/git/custom_vpn/certs/ca.pem")
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