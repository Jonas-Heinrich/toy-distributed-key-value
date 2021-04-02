package kv

import "net"

var BASE_IP_ADDRESS = net.IPv4(172, 22, 0, 0)
var LEADER_IP_ADDRESS = net.IPv4(172, 22, 0, 2)
var PORT string = ":8080"

func GetIPAdress(s byte) net.IP {
	ipAddress := make(net.IP, len(BASE_IP_ADDRESS))
	copy(ipAddress, BASE_IP_ADDRESS)
	ipAddress[len(ipAddress)-1] = s
	return ipAddress
}

func GetURL(ip net.IP, path string) string {
	return "http://" + ip.String() + PORT + path
}
