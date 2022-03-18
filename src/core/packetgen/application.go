package packetgen

import (
	"fmt"

	"github.com/google/gopacket"

	"github.com/Arriven/db1000n/src/utils"
)

func BuildApplicationLayer(c LayerConfig) (gopacket.ApplicationLayer, error) {
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
	default:
		return nil, fmt.Errorf("unsupported application layer type %s", c.Type)
	}
}
