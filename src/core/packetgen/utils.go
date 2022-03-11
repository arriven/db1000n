// MIT License

// Copyright (c) [2022] [Arriven (https://github.com/Arriven)]

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package packetgen [allows sending customized tcp/udp traffic. Inspired by https://github.com/bilalcaliskan/syn-flood]
package packetgen

import (
	"fmt"
	"math/rand"
	"net"
)

// RandomPayload returns a byte slice to spoof ip packets with random payload in specified length
func RandomPayload(length int) []byte {
	payload := make([]byte, length)
	rand.Read(payload)
	return payload
}

// RandomIP returns a random ip to spoof packets
func RandomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255)+1, rand.Intn(255)+1,
		rand.Intn(255)+1, rand.Intn(255)+1)
}

// RandomPort returns a random port to spoof packets
func RandomPort() int {
	const minPort = 1024
	const maxPort = 65535
	return rand.Intn(maxPort-minPort) + minPort
}

// RandomMacAddr returns a random mac address to spoof packets
func RandomMacAddr() net.HardwareAddr {
	buf := make([]byte, 6)
	rand.Read(buf)
	var addr net.HardwareAddr = buf
	return net.HardwareAddr(addr.String())
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
