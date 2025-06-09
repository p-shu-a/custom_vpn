package config

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

// Server Specific
var (
	// Port for raw TCP listener
	RawTcpServerPort = 9000
	// Port for TCP+TLS listener
	TcpTlsServerPort = 9001
	// Port for QUIC listener
	QuicServerPort	 = 9002
	// Used in QUIC configs to adjust connection timeouts
	TimeOutDuration  = time.Second * 15
)

// QUIC config for server
var	ServerQuicConf = quic.Config{
	EnableDatagrams: true,
	MaxIdleTimeout: TimeOutDuration,
	Allow0RTT: true,
}

// QUIC config for client
var ClientQuicConfig = quic.Config{
	EnableDatagrams: true,
}

// Server Endpoint Services
var (
	HTTPEndpointService = net.TCPAddr{
		IP: net.ParseIP("127.0.0.1"),
		Port: 8080,
	}
	SSHEndpointService = net.TCPAddr{
		IP: net.ParseIP("127.0.0.1"),
		Port: 22,
	}
)

// Client Specifc
var (
	// Port on which the client app recieves requests
	ClientListnerPort = 2022
)