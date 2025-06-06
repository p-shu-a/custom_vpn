package tun

import (
	"github.com/songgao/water"
)


func CreateTUN() (*water.Interface, error){
	// create a water config
	config := water.Config{
		DeviceType: water.TUN,
	}

	// create a new interface
	iface, err := water.New(config)
	if err != nil {
		return nil, err
	}
	
	return iface, nil
}

// how are QUIC Conns and a TUN different?
// i mean, they ARE different (interface, vs conn) but can quic do the work of an interface?
// quic routes based on a per-port basis, 
// is wireguard solving this problem?