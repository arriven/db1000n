package packetgen

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/miekg/dns"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func TestSerialize(t *testing.T) {
	t.Parallel()

	configTpls := []map[string]any{
		{
			"network": map[string]any{
				"type": "ipv6",
				"data": map[string]any{
					"src_ip": "{{ local_ipv6 }}",
					"dst_ip": "{{ local_ipv6 }}",
				},
			},
			"transport": map[string]any{
				"type": "tcp",
				"data": map[string]any{
					"src_port": "{{ random_port }}",
					"dst_port": "{{ random_port }}",
					"flags": map[string]any{
						"syn": true,
					},
				},
			},
			"payload": map[string]any{
				"type": "raw",
				"data": map[string]any{
					"payload": "test",
				},
			},
			"expected_result": "test",
		},
		{
			"network": map[string]any{
				"type": "ipv4",
				"data": map[string]any{
					"src_ip": "{{ local_ipv4 }}",
					"dst_ip": "{{ local_ipv4 }}",
				},
			},
			"transport": map[string]any{
				"type": "tcp",
				"data": map[string]any{
					"src_port": "{{ random_port }}",
					"dst_port": "{{ random_port }}",
				},
			},
			"payload": map[string]any{
				"type": "icmpv4",
				"data": map[string]any{
					"type_code": 130,
					"seq":       1231231,
					"id":        1231231231,
				},
			},
			"expected_result": 130,
		},
	}

	logger := zap.NewExample()

	for _, configTpl := range configTpls {
		config := templates.ParseAndExecuteMapStruct(logger, configTpl, nil)

		var packetConfig PacketConfig
		if err := utils.Decode(config, &packetConfig); err != nil {
			t.Fatal(err)
		}

		logger.Debug("deserialized packet config", zap.Any("packet", packetConfig))

		packet, err := packetConfig.Build()
		if err != nil {
			t.Fatal(err)
		}

		buf := gopacket.NewSerializeBuffer()
		if err = packet.Serialize(buf); err != nil {
			t.Fatal(err)
		}

		var packetData gopacket.Packet

		logger.Debug("serialized packet", zap.Any("layers", buf.Layers()))

		switch packetConfig.Network.Type {
		case "ipv4":
			packetData = gopacket.NewPacket(buf.Bytes(), layers.LayerTypeIPv4, gopacket.Default)
			if packetData.ErrorLayer() != nil {
				t.Fatal("error deserializing packet", string(packetData.ErrorLayer().LayerPayload()))
			}
		case "ipv6":
			packetData = gopacket.NewPacket(buf.Bytes(), layers.LayerTypeIPv6, gopacket.Default)
			if packetData.ErrorLayer() != nil {
				t.Fatal("error deserializing packet", string(packetData.ErrorLayer().LayerPayload()))
			}
		}

		logger.Debug("deserialized packet data", zap.Any("packet", packetData.String()))

		err = extractPayload(packetData, configTpl)

		if err != nil {
			t.Fatal(err)
		}
	}
}

func extractPayload(p gopacket.Packet, c map[string]any) error {
	switch res := c["expected_result"].(type) {
	case int:
		if binary.BigEndian.Uint16(p.ApplicationLayer().Payload()) != uint16(res) {
			return fmt.Errorf("bad int payload - %v", p.ApplicationLayer().Payload())
		}
	case string:
		if string(p.ApplicationLayer().Payload()) != res {
			return fmt.Errorf("bad payload - %v", string(p.ApplicationLayer().Payload()))
		}
	}

	return nil
}

func TestDNS(t *testing.T) {
	t.Parallel()

	configTPL := map[string]any{
		"payload": map[string]any{
			"type": "dns",
			"data": map[string]any{
				"id":      1234,
				"op_code": 0,
				"rd":      true,
				"questions": []map[string]any{
					{
						"name":  "google.com",
						"type":  "1",
						"class": "1",
					},
				},
			},
		},
	}

	logger := zap.NewExample()
	config := templates.ParseAndExecuteMapStruct(logger, configTPL, nil)

	var packetConfig PacketConfig
	if err := utils.Decode(config, &packetConfig); err != nil {
		t.Fatal(err)
	}

	logger.Debug("deserialized packet config", zap.Any("packet", packetConfig))

	packet, err := packetConfig.Build()
	if err != nil {
		t.Fatal(err)
	}

	buf := gopacket.NewSerializeBuffer()
	if err = packet.Serialize(buf); err != nil {
		t.Fatal(err)
	}

	question := new(dns.Msg).
		SetQuestion(dns.Fqdn("google.com"), 1)
	question.Id = 1234

	out, err := question.Pack()
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(buf.Bytes(), out) {
		t.Fatal(fmt.Errorf("not equal %v to %v", buf.Bytes(), out))
	}
}
