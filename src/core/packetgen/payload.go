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
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/Arriven/db1000n/src/utils"
)

func BuildPayload(c LayerConfig) (gopacket.Layer, error) {
	switch c.Type {
	case "":
		return nil, nil
	case "raw":
		var packetConfig struct {
			Payload string
		}

		if err := utils.Decode(c.Data, &packetConfig); err != nil {
			return nil, err
		}

		return gopacket.Payload([]byte(packetConfig.Payload)), nil
	case "icmpv4":
		var packetConfig ICMPV4PacketConfig
		if err := utils.Decode(c.Data, &packetConfig); err != nil {
			return nil, err
		}

		return buildICMPV4Packet(packetConfig), nil
	case "dns":
		var packetConfig DNSPacketConfig
		if err := utils.Decode(c.Data, &packetConfig); err != nil {
			return nil, err
		}

		return buildDNSPacket(packetConfig), nil
	default:
		return nil, fmt.Errorf("unsupported layer type %s", c.Type)
	}
}

type ICMPV4PacketConfig struct {
	TypeCode uint16
	ID       uint16
	Seq      uint16
}

func buildICMPV4Packet(c ICMPV4PacketConfig) *layers.ICMPv4 {
	return &layers.ICMPv4{
		TypeCode: layers.ICMPv4TypeCode(c.TypeCode),
		Id:       c.ID,
		Seq:      c.Seq,
	}
}

type DNSPacketConfig struct {
	ID      uint16
	Qr      bool
	OpCode  uint8
	QDCount uint16
}

func buildDNSPacket(c DNSPacketConfig) *layers.DNS {
	return &layers.DNS{
		ID:      c.ID,
		QR:      c.Qr,
		OpCode:  layers.DNSOpCode(c.OpCode),
		QDCount: c.QDCount,
	}
}
