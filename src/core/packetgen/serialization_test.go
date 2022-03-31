package packetgen

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func TestSerialize(t *testing.T) {
	t.Parallel()

	configTpls := []map[string]interface{}{
		{
			"network": map[string]interface{}{
				"type": "ipv6",
				"data": map[string]interface{}{
					"src_ip": "{{ local_ipv6 }}",
					"dst_ip": "{{ local_ipv6 }}",
				},
			},
			"transport": map[string]interface{}{
				"type": "tcp",
				"data": map[string]interface{}{
					"src_port": "{{ random_port }}",
					"dst_port": "{{ random_port }}",
					"flags": map[string]interface{}{
						"syn": true,
					},
				},
			},
			"payload": map[string]interface{}{
				"type": "raw",
				"data": map[string]interface{}{
					"payload": "test",
				},
			},
			"expected_result": "test",
		},
		{
			"network": map[string]interface{}{
				"type": "ipv4",
				"data": map[string]interface{}{
					"src_ip": "{{ local_ipv4 }}",
					"dst_ip": "{{ local_ipv4 }}",
				},
			},
			"transport": map[string]interface{}{
				"type": "tcp",
				"data": map[string]interface{}{
					"src_port": "{{ random_port }}",
					"dst_port": "{{ random_port }}",
				},
			},
			"payload": map[string]interface{}{
				"type": "icmpv4",
				"data": map[string]interface{}{
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

func extractPayload(p gopacket.Packet, c map[string]interface{}) error {
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
