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
	configTpl := map[string]interface{}{
		"network": map[string]interface{}{
			"type": "ipv4",
			"data": map[string]interface{}{
				"src_ip": "{{ local_ip }}",
				"dst_ip": "{{ local_ip }}",
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
		"application": map[string]interface{}{
			"type": "raw",
			"data": map[string]interface{}{
				"payload": "test",
			},
		},
	}

	logger := zap.NewExample()
	defer logger.Sync()

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

	packetData := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeIPv4, gopacket.Default)
	logger.Debug("deserialized packet data", zap.Any("packet", packetData.String()))
	if string(packetData.ApplicationLayer().Payload()) != "test" {
		t.Fatal("bad payload", string(packetData.ApplicationLayer().Payload()))
	}
}
