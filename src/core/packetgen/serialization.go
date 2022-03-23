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
