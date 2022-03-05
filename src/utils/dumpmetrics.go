package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Arriven/db1000n/src/logs"
	"github.com/Arriven/db1000n/src/metrics"
)

func DumpMetrics(l *logs.Logger, path, name, clientID string) {
	defer PanicHandler()

	bytesPerSecond := metrics.Default.Read(name)
	if bytesPerSecond > 0 {
		l.Info("Атака проводиться успішно! Руський воєнний корабль іди нахуй!\n")
		l.Info("Attack is successful! Russian warship, go fuck yourself!\n")
		l.Info("The app is generating approximately %v bytes per second\n", bytesPerSecond)
		ReportStatistics(int64(bytesPerSecond), clientID)
	} else {
		l.Warning("The app doesn't seem to generate any traffic, please contact your admin")
	}
	if path == "" {
		return
	}
	type metricsDump struct {
		BytesPerSecond int `json:"bytes_per_second"`
	}
	dump := &metricsDump{
		BytesPerSecond: bytesPerSecond,
	}
	dumpBytes, err := json.Marshal(dump)
	if err != nil {
		l.Warning("failed marshaling metrics: %v", err)
		return
	}
	// TODO: use proper ip
	url := fmt.Sprintf("%s?id=%s", path, clientID)
	resp, err := http.Post(url, "application/json", bytes.NewReader(dumpBytes))
	if err != nil {
		l.Warning("failed sending metrics: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		l.Warning("bad response when sending metrics. code %v", resp.StatusCode)
	}
}
