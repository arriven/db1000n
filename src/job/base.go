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

// Package job [contains all the attack types db1000n can simulate]
package job

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/job/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

// GlobalConfig passes commandline arguments to every job.
type GlobalConfig struct {
	ClientID string

	ProxyURLs           string
	proxyPath           string
	SkipEncrypted       bool
	EnablePrimitiveJobs bool
	ScaleFactor         int
	MinInterval         time.Duration
	Backoff             utils.BackoffConfig
}

func (config *GlobalConfig) GetProxiesFromPath() {

	if config.proxyPath != "" {
		proxies, err := templates.GetURLContent(config.proxyPath)
		if err != nil {
			log.Println("Could not obtain proxies from given proxy path", err.Error())
			return
		}

		if len(proxies) == 0 {
			log.Println("Given proxy path returns nothing")
			return
		}

		config.ProxyURLs = strings.Join(strings.Split(proxies, "\n"), ",")
	}
}

// NewGlobalConfigWithFlags returns a GlobalConfig initialized with command line flags.
func NewGlobalConfigWithFlags() *GlobalConfig {
	res := GlobalConfig{
		ClientID: uuid.NewString(),
	}

	flag.StringVar(&res.ProxyURLs, "proxy", utils.GetEnvStringDefault("SYSTEM_PROXY", ""),
		"system proxy to set by default (can be a comma-separated list or a template)")
	flag.StringVar(&res.proxyPath, "proxy-path", utils.GetEnvStringDefault("SYTEM_PROXY_PATH", ""),
		"url to a list of proxies")
	flag.BoolVar(&res.SkipEncrypted, "skip-encrypted", utils.GetEnvBoolDefault("SKIP_ENCRYPTED", false),
		"set to true if you want to only run plaintext jobs from the config for security considerations")
	flag.BoolVar(&res.EnablePrimitiveJobs, "enable-primitive", utils.GetEnvBoolDefault("ENABLE_PRIMITIVE", true),
		"set to true if you want to run primitive jobs that are less resource-efficient")
	flag.IntVar(&res.ScaleFactor, "scale", utils.GetEnvIntDefault("SCALE_FACTOR", 1),
		"used to scale the amount of jobs being launched, effect is similar to launching multiple instances at once")
	flag.DurationVar(&res.MinInterval, "min-interval", utils.GetEnvDurationDefault("MIN_INTERVAL", 0),
		"minimum interval between job iterations")

	flag.IntVar(&res.Backoff.Limit, "backoff-limit", utils.GetEnvIntDefault("BACKOFF_LIMIT", utils.DefaultBackoffConfig().Limit),
		"how much exponential backoff can be scaled")
	flag.IntVar(&res.Backoff.Multiplier, "backoff-multiplier", utils.GetEnvIntDefault("BACKOFF_MULTIPLIER", utils.DefaultBackoffConfig().Multiplier),
		"how much exponential backoff is scaled with each new error")
	flag.DurationVar(&res.Backoff.Timeout, "backoff-timeout", utils.GetEnvDurationDefault("BACKOFF_TIMEOUT", utils.DefaultBackoffConfig().Timeout),
		"initial exponential backoff timeout")

	return &res
}

// Job comment for linter
type Job = func(ctx context.Context, logger *zap.Logger, globalConfig *GlobalConfig, args config.Args) (data any, err error)

// Get job by type name
//nolint:cyclop // The string map alternative is orders of magnitude slower
func Get(t string) Job {
	switch t {
	case "http", "http-flood":
		return fastHTTPJob
	case "http-request":
		return singleRequestJob
	case "tcp":
		return tcpJob
	case "udp":
		return udpJob
	case "slow-loris":
		return slowLorisJob
	case "packetgen":
		return packetgenJob
	case "dns-blast":
		return dnsBlastJob
	case "sequence":
		return sequenceJob
	case "parallel":
		return parallelJob
	case "log":
		return logJob
	case "set-value":
		return setVarJob
	case "check":
		return checkJob
	case "loop":
		return loopJob
	case "encrypted":
		return encryptedJob
	default:
		return nil
	}
}

type Config interface {
	FromGlobal(GlobalConfig)
}

func ParseConfig(c Config, args config.Args, global GlobalConfig) error {
	if err := utils.Decode(args, c); err != nil {
		return err
	}

	c.FromGlobal(global)

	return nil
}

// BasicJobConfig comment for linter
type BasicJobConfig struct {
	IntervalMs int
	Interval   *time.Duration
	utils.Counter
	Backoff *utils.BackoffConfig
}

func (c *BasicJobConfig) FromGlobal(global GlobalConfig) {
	if c.GetInterval() < global.MinInterval {
		c.Interval = &global.MinInterval
	}

	if c.Backoff == nil {
		c.Backoff = &global.Backoff
	}
}

func (c BasicJobConfig) GetInterval() time.Duration {
	return utils.NonNilOrDefault(c.Interval, time.Duration(c.IntervalMs)*time.Millisecond)
}

// Next comment for linter
func (c *BasicJobConfig) Next(ctx context.Context) bool {
	return utils.Sleep(ctx, c.GetInterval()) && c.Counter.Next()
}
