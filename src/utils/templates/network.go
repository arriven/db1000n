// MIT License

// Copyright (c) [2022] [Bohdan Ivashko (https://github.com/Arriven)]

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

package templates

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
)

// RandomPayload returns a string to spoof ip packets with random payload in specified length.
func RandomPayload(length int) string {
	return string(RandomPayloadByte(length))
}

// RandomPayloadByte returns a byte slice to spoof ip packets with random payload in specified length.
func RandomPayloadByte(length int) []byte {
	payload := make([]byte, length)
	rand.Read(payload) //nolint:gosec // Cryptographically secure random not required

	return payload
}

// RandomIP returns a random ip to spoof packets.
func RandomIP() string {
	const maxByte = 255

	//nolint:gosec // Cryptographically secure random not required
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(maxByte)+1, rand.Intn(maxByte)+1,
		rand.Intn(maxByte)+1, rand.Intn(maxByte)+1)
}

// RandomPort returns a random port to spoof packets.
func RandomPort() int {
	const (
		minPort = 1024
		maxPort = 65535
	)

	return rand.Intn(maxPort-minPort) + minPort //nolint:gosec // Cryptographically secure random not required
}

// RandomMacAddr returns a random MAC address to spoof packets.
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

func localIP(filter func(ip net.IP) bool) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if filter(ipnet.IP) {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

func resolveHost(host string, filter func(ip net.IP) bool) (string, error) {
	addrs, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if filter(addr) {
			return addr.String(), nil
		}
	}

	return "", fmt.Errorf("no addrs found for host %v", host)
}

// LocalIPV4 returns the first non loopback local ipv4 of the host
func LocalIPV4() string {
	return localIP(func(ip net.IP) bool { return ip.To4() != nil })
}

// ResolveHostIPV4 function gets a string and returns the ipv4 address
func ResolveHostIPV4(host string) (string, error) {
	return resolveHost(host, func(ip net.IP) bool { return ip.To4() != nil })
}

// LocalIPV6 returns the first non loopback local ipv4 of the host
func LocalIPV6() string {
	return localIP(func(ip net.IP) bool { return ip.To4() == nil })
}

// ResolveHostIPV6 function gets a string and returns the ipv4 address
func ResolveHostIPV6(host string) (string, error) {
	return resolveHost(host, func(ip net.IP) bool { return ip.To4() == nil })
}

func cookieString(cookies map[string]string) string {
	s := ""
	for key, value := range cookies {
		s = fmt.Sprintf("%s %s=%s;", s, key, value)
	}

	return strings.Trim(strings.TrimSpace(s), ";")
}
