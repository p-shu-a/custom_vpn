package config

import (
	"net"
	"time"
)


var (
	RawTcpServerPort = 9000
	TcpTlsServerPort = 9001
	QuicServerPort	 = 9002
	TimeOutDuration  = time.Second * 15
)

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