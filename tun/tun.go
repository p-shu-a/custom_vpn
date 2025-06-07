package tun

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"

	"github.com/songgao/water"
)

// TunDetails contains the details for a TUN interface
type TunDetails struct{
	// Reference to a TUN interface
	TunIface *water.Interface
	// holds VIP of the local
	LocalVIP net.IP
	// hold the IP of the peer
	DestIP  net.IP
}

// Called this function creates a TUN interface. Assigns a VIP to the interface. Adds a Route
// Since uTUNs are ephemeral, the route rules related to them are also lost when uTUNs disappear.
func ConfigureTUN(localVIP, peerVIP net.IP) (*TunDetails, error) {

	var details TunDetails

	iface, err := createTUN()
	if err != nil {
		return nil, err
	}

	details.TunIface = iface
	details.LocalVIP = localVIP
	details.DestIP = peerVIP

	if err = addVIP(details); err != nil{
		return nil, err
	}

	if err = addRoute(details); err != nil{
		return nil, err
	}

	return &details, nil
}

func createTUN() (*water.Interface, error) {
	// create a water config
	config := water.Config{
		DeviceType: water.TUN,
	}

	// create a new interface
	iface, err := water.New(config)
	if err != nil {
		return nil, fmt.Errorf("died creating new TUN interface: %v", err)
	}
	
	return iface, nil
}

// configuring VIPs for the given interface
func addVIP(details TunDetails) error {

	var cmd *exec.Cmd

	if runtime.GOOS == "darwin" {
		log.Print("we're on darwin (macos)")
		//                 use ifconfg,      with  TUN         , IPv4  ,          local VIP             ,            peer VIP          , activate the interface
		cmd = exec.Command("ifconfig", details.TunIface.Name(), "inet", details.LocalVIP.To4().String(), details.DestIP.To4().String(), "up")
		if err := cmd.Run(); err != nil{
			return fmt.Errorf("died assigning VIP to client: %v", err)
		}
	} else if runtime.GOOS == "linux" {
		log.Print("we're on linux")
		vip := fmt.Sprintf("%v/32", details.LocalVIP.String())
		cmd = exec.Command("ip", "addr", "add", vip, "dev", details.TunIface.Name())
		if err := cmd.Run(); err != nil{
			return fmt.Errorf("died adding addr to interface: %v", err)
		}
		cmd = exec.Command("ip", "link", "set", "dev", details.TunIface.Name(), "up")
		if err := cmd.Run(); err != nil{
			return fmt.Errorf("died activating interface: %v", err)
		} 
	}
	
	return nil
}

// Create an IP route which forwards all data destined for our Peer interface through the passed interface
// Given that i'm shelling out and issuing commands, this is not very interopeable between OSs.
func addRoute(details TunDetails) error {
	
	var cmd *exec.Cmd
	switch runtime.GOOS  {
	case "darwin":
		cmd = exec.Command("route", "add", details.DestIP.To4().String(), "-interface", details.TunIface.Name())
	case "linux":
		cmd = exec.Command("route", "add", details.DestIP.To4().String(), "dev", details.TunIface.Name())
	}
	
	if err := cmd.Run(); err != nil{
		return fmt.Errorf("died adding route to server: %v", err)
	}

	return nil
}