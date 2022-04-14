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
	"bytes"
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/valyala/fasthttp"

	"github.com/Arriven/db1000n/src/core/http"
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
	case "http":
		return buildHTTPPacket(c.Data)
	case "icmpv4":
		return buildICMPV4Packet(c.Data)
	case "dns":
		return buildDNSPacket(c.Data)
	default:
		return nil, fmt.Errorf("unsupported layer type %s", c.Type)
	}
}

type ICMPV4PacketConfig struct {
	TypeCode uint16
	ID       uint16
	Seq      uint16
}

func buildICMPV4Packet(data map[string]any) (*layers.ICMPv4, error) {
	var c ICMPV4PacketConfig
	if err := utils.Decode(data, &c); err != nil {
		return nil, err
	}

	return &layers.ICMPv4{
		TypeCode: layers.ICMPv4TypeCode(c.TypeCode),
		Id:       c.ID,
		Seq:      c.Seq,
	}, nil
}

// DNSQuestion wraps a single request (question) within a DNS query.
type DNSQuestion struct {
	Name  string
	Type  layers.DNSType
	Class layers.DNSClass
}

type DNSPacketConfig struct {
	ID     uint16
	Qr     bool
	OpCode uint8

	AA bool  // Authoritative answer
	TC bool  // Truncated
	RD bool  // Recursion desired
	RA bool  // Recursion available
	Z  uint8 // Reserved for future use

	ResponseCode uint8
	QDCount      *uint16 // Number of questions to expect

	// Entries
	Questions []DNSQuestion
}

func buildDNSPacket(data map[string]any) (*layers.DNS, error) {
	var c DNSPacketConfig
	if err := utils.Decode(data, &c); err != nil {
		return nil, err
	}

	questions := make([]layers.DNSQuestion, 0, len(c.Questions))
	for _, question := range c.Questions {
		questions = append(questions, layers.DNSQuestion{Name: []byte(question.Name), Type: question.Type, Class: question.Class})
	}

	return &layers.DNS{
		ID:     c.ID,
		QR:     c.Qr,
		OpCode: layers.DNSOpCode(c.OpCode),

		AA: c.AA,
		TC: c.TC,
		RD: c.RD,
		RA: c.RA,
		Z:  c.Z,

		QDCount:   utils.NonNilOrDefault(c.QDCount, uint16(len(c.Questions))),
		Questions: questions,
	}, nil
}

func buildHTTPPacket(data map[string]any) (gopacket.Payload, error) {
	var c http.RequestConfig

	if err := utils.Decode(data, &c); err != nil {
		return nil, err
	}

	var req fasthttp.Request

	http.InitRequest(c, &req)

	var buf bytes.Buffer

	if _, err := req.WriteTo(&buf); err != nil {
		return nil, err
	}

	return gopacket.Payload(buf.Bytes()), nil
}
