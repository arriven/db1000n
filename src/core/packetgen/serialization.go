package packetgen

import (
	"fmt"

	"github.com/google/gopacket"
)

var opts = gopacket.SerializeOptions{
	FixLengths:       true,
	ComputeChecksums: true,
}

func SerializeLayers(payloadBuf gopacket.SerializeBuffer, layers ...gopacket.Layer) error {
	serializableLayers := make([]gopacket.SerializableLayer, 0, len(layers))

	for _, layer := range layers {
		if layer == nil {
			continue
		}

		serializableLayer, err := toSerializable(layer)
		if err != nil {
			return err
		}

		serializableLayers = append(serializableLayers, serializableLayer)
	}

	return gopacket.SerializeLayers(payloadBuf, opts, serializableLayers...)
}

func Serialize(payloadBuf gopacket.SerializeBuffer, layer gopacket.Layer) error {
	serializable, err := toSerializable(layer)
	if err != nil {
		return err
	}

	return serializable.SerializeTo(payloadBuf, opts)
}

func toSerializable(layer gopacket.Layer) (gopacket.SerializableLayer, error) {
	serializable, ok := layer.(gopacket.SerializableLayer)
	if !ok {
		return nil, fmt.Errorf("layer is not serializable: %v", layer.LayerType())
	}

	return serializable, nil
}
