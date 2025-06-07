package config

import "net"

var (
	ClientVIP = net.ParseIP("10.0.0.2")
	ServerVIP = net.ParseIP("10.0.0.1")
)