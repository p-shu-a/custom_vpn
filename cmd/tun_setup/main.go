package main

import (
	"custom_vpn/internal/helpers"
	"custom_vpn/tun"
	"log"
	"os/exec"
)

/// This module really is to create uTUN interfaces for client and server
/// and add IP routes.
/// Since uTUNs are ephemeral, the route rules related to them are also lost when uTUNs disappear
/// i'm almost certian this is not the final implmentation, but it is to help me get going for now
/// it was getting difficult mimic flows with localhost, the only differentiator being Port nums
/// need to assign VIPs

func main(){

	ctx := helpers.SetupShutdownHelper()

	clientIfce, err := tun.CreateTUN()
	if err != nil{
		log.Fatalf("failed to create client TUN device: %v", err)
	}

	serverIfce, err := tun.CreateTUN()
	if err != nil{
		log.Fatalf("failed to create server TUN device: %v", err)
	}

	serverVIP := "10.0.0.1"
	clientVIP := "10.0.0.2"

	// configuring VIPs. 
	// add a vip to the tun interface. this is for client to server
	//          use ifconfg, with client TUN  ,  IPv4 , local VIP, peer VIP , activate the interface
	cmd := exec.Command("ifconfig", clientIfce.Name(), "inet", clientVIP, serverVIP, "up")
	if err = cmd.Run(); err != nil{
		log.Fatalf("died assigning VIP to client: %v", err)
	}
	// route all data destined to server to flow through client interface     
	cmd = exec.Command("route", "add", serverVIP, "-interface", clientIfce.Name())
	if err = cmd.Run(); err != nil{
		log.Fatalf("died adding route to server: %v", err)
	}
	
	// same setup for server
	cmd = exec.Command("ifconfig", serverIfce.Name(), "inet", serverVIP, clientVIP, "up")
	if err = cmd.Run(); err != nil{
		log.Fatalf("died assigning VIP to server: %v", err)
	}
	cmd = exec.Command("route", "add", clientVIP, "-interface", serverIfce.Name())
	if err = cmd.Run(); err != nil{
		log.Fatalf("died adding route to client: %v", err)
	}
	
	<-ctx.Done()

}



