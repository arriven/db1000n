package templates

import (
	"fmt"
	"math/rand"
	"net"
)

// RandomPayload returns a byte slice to spoof ip packets with random payload in specified length
func RandomPayload(length int) string {
	payload := make([]byte, length)
	rand.Read(payload) //nolint:gosec // Cryptographically secure random not required

	return string(payload)
}

// RandomIP returns a random ip to spoof packets
func RandomIP() string {
	const maxByte = 255

	//nolint:gosec // Cryptographically secure random not required
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(maxByte)+1, rand.Intn(maxByte)+1,
		rand.Intn(maxByte)+1, rand.Intn(maxByte)+1)
}

// RandomPort returns a random port to spoof packets
func RandomPort() int {
	const (
		minPort = 1024
		maxPort = 65535
	)

	return rand.Intn(maxPort-minPort) + minPort //nolint:gosec // Cryptographically secure random not required
}

// RandomMacAddr returns a random mac address to spoof packets
func RandomMacAddr() net.HardwareAddr {
	const macSizeBytes = 6

	addr := make(net.HardwareAddr, macSizeBytes)
	rand.Read(addr) //nolint:gosec // Cryptographically secure random not required

	return addr
}

// LocalMacAddres returns first valid mac address
func LocalMacAddres() string {
	ifas, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			return a
		}
	}

	return ""
}

// LocalIP returns the first non loopback local IP of the host
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

// ResolveHost function gets a string and returns the ip address
func ResolveHost(host string) (string, error) {
	addrs, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if addr.To4() != nil {
			return addr.String(), nil
		}
	}

	return "", fmt.Errorf("no addrs found for host %v", host)
}
