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

package packetgen

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"golang.org/x/net/ipv6"

	"github.com/Arriven/db1000n/src/utils"
)

// ConnectionConfig describes which network to use when sending packets
type ConnectionConfig struct {
	Type string
	Args map[string]any
}

func OpenConnection(ctx context.Context, c ConnectionConfig) (Connection, error) {
	switch c.Type {
	case "raw":
		var cfg rawConnConfig
		if err := utils.Decode(c.Args, &cfg); err != nil {
			return nil, fmt.Errorf("error decoding connection config: %w", err)
		}

		return openRawConn(cfg)
	case "net":
		var cfg netConnConfig
		if err := utils.Decode(c.Args, &cfg); err != nil {
			return nil, fmt.Errorf("error decoding connection config: %w", err)
		}

		return openNetConn(ctx, cfg)
	default:
		return nil, fmt.Errorf("unknown connection type: %v", c.Type)
	}
}

type Connection interface {
	Write(Packet) (int, error)
	Close() error
	Target() string
}

// raw ipv4/ipv6 connection
type rawConnConfig struct {
	Name    string
	Address string
}

type rawConn struct {
	*ipv6.PacketConn
	buf gopacket.SerializeBuffer

	target string
}

// openRawConn opens a raw ip network connection based on the provided config
// use ipv6 as it also supports ipv4
func openRawConn(c rawConnConfig) (*rawConn, error) {
	packetConn, err := net.ListenPacket(c.Name, c.Address)
	if err != nil {
		return nil, err
	}

	return &rawConn{
		PacketConn: ipv6.NewPacketConn(packetConn),
		buf:        gopacket.NewSerializeBuffer(),
		target:     c.Name + "://" + c.Address,
	}, nil
}

func (conn *rawConn) Write(packet Packet) (n int, err error) {
	if err := packet.Serialize(conn.buf); err != nil {
		return 0, fmt.Errorf("error serializing packet: %w", err)
	}

	return conn.PacketConn.WriteTo(conn.buf.Bytes(), nil, &net.IPAddr{IP: packet.IP()})
}

func (conn *rawConn) Close() error {
	return conn.PacketConn.Close()
}

func (conn *rawConn) Target() string { return conn.target }

type netConnConfig struct {
	Protocol        string
	Address         string
	Timeout         time.Duration
	ProxyURLs       string
	LocalAddr       string
	Reader          *netReaderConfig
	TLSClientConfig *tls.Config
}

type netReaderConfig struct {
	Interval time.Duration
}

type netConn struct {
	net.Conn
	buf gopacket.SerializeBuffer

	target string
}

func readStub(ctx context.Context, conn net.Conn, c *netReaderConfig) {
	const bufSize = 1024
	buf := make([]byte, bufSize)
	ticker := time.NewTicker(c.Interval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := conn.Read(buf)
			if err != nil {
				return
			}
		}
	}
}

func openNetConn(ctx context.Context, c netConnConfig) (*netConn, error) {
	conn, err := utils.GetProxyFunc(c.ProxyURLs, utils.ResolveAddr(c.Protocol, c.LocalAddr), c.Timeout, false)(c.Protocol, c.Address)

	switch {
	case err != nil:
		return nil, err
	case c.TLSClientConfig == nil:
		if c.Reader != nil {
			go readStub(ctx, conn, c.Reader)
		}

		return &netConn{Conn: conn, buf: gopacket.NewSerializeBuffer(), target: c.Protocol + "://" + c.Address}, nil
	}

	tlsConn := tls.Client(conn, c.TLSClientConfig)
	if err = tlsConn.Handshake(); err != nil {
		tlsConn.Close()

		return nil, err
	}

	if c.Reader != nil {
		go readStub(ctx, conn, c.Reader)
	}

	return &netConn{Conn: tlsConn, buf: gopacket.NewSerializeBuffer(), target: c.Protocol + "://" + c.Address}, nil
}

func (conn *netConn) Write(packet Packet) (n int, err error) {
	if err = packet.Serialize(conn.buf); err != nil {
		return 0, fmt.Errorf("error serializing packet: %w", err)
	}

	return conn.Conn.Write(conn.buf.Bytes())
}

func (conn *netConn) Close() error {
	return conn.Conn.Close()
}

func (conn *netConn) Target() string { return conn.target }
