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
	"net"

	"github.com/google/gopacket"
)

type Packet struct {
	Link      gopacket.LinkLayer
	Network   gopacket.NetworkLayer
	Transport gopacket.TransportLayer
	Payload   gopacket.Layer
}

type LayerConfig struct {
	Type string
	Data map[string]interface{}
}

type PacketConfig struct {
	Link      LayerConfig
	Network   LayerConfig
	Transport LayerConfig
	Payload   LayerConfig
}

func (c PacketConfig) Build() (result Packet, err error) {
	if result.Link, err = BuildLinkLayer(c.Link); err != nil {
		return Packet{}, err
	}

	if result.Network, err = BuildNetworkLayer(c.Network); err != nil {
		return Packet{}, err
	}

	if result.Transport, err = BuildTransportLayer(c.Transport, result.Network); err != nil {
		return Packet{}, err
	}

	if result.Payload, err = BuildPayload(c.Payload); err != nil {
		return Packet{}, err
	}

	return result, nil
}

func (p Packet) Serialize(payloadBuf gopacket.SerializeBuffer) (err error) {
	return SerializeLayers(payloadBuf, p.Link, p.Network, p.Transport, p.Payload)
}

func (p Packet) IP() net.IP {
	return p.Network.NetworkFlow().Dst().Raw()
}
