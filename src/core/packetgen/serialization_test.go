package packetgen

import (
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func TestSerialize(t *testing.T) {
	t.Parallel()

	configTpl := map[string]interface{}{
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
	}

	logger := zap.NewExample()

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

	logger.Debug("serialized packet", zap.Any("layers", buf.Layers()))

	packetData := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeIPv6, gopacket.Default)
	if packetData.ErrorLayer() != nil {
		t.Fatal("error deserializing packet", string(packetData.ErrorLayer().LayerPayload()))
	}

	logger.Debug("deserialized packet data", zap.Any("packet", packetData.String()))

	if string(packetData.ApplicationLayer().Payload()) != "test" {
		t.Fatal("bad payload", string(packetData.ApplicationLayer().Payload()))
	}
}
