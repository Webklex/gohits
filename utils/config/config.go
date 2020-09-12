package config

import (
	"../filesystem"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

func NewConfig() *Config {
	return &Config{}
}

func DefaultConfig() *Config {
	dir, _ := os.Getwd()

	return &Config{
		FastOpen:   false,
		Naggle:     false,
		ServerAddr: "localhost:8080",
		HTTP2:      true,
		HSTS:       "",

		TLSCertFile:         "cert.pem",
		TLSKeyFile:          "key.pem",
		LetsEncrypt:         false,
		LetsEncryptCacheDir: ".",
		LetsEncryptEmail:    "",
		LetsEncryptHosts:    "",

		APIPrefix:         "/",
		CORSOrigin:        "*",
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      15 * time.Second,
		RateLimitLimit:    1,
		RateLimitBurst:    3,
		LogTimestamp:      true,
		RateLimitInterval: 3 * time.Minute,

		SessionLifetime:  20 * time.Minute,
		WriteWait:        10 * time.Second,
		ReadWait:         10 * time.Second,
		MaxMessageSize:   8192,
		PongWait:         24 * time.Second,
		PingPeriod:       12 * time.Second,
		CloseGracePeriod: 6 * time.Second,

		RootDir:        dir,
		GuiDir:         "gui",
		File:           path.Join(dir, "conf", "settings.config"),
		SaveConfigFlag: false,
		Silent:         false,
	}
}

// AddFlags adds configuration flags to the given FlagSet.
func (c *Config) AddFlags(fs *flag.FlagSet) {
	defer envconfig.Process("gohits", c)

	fs.StringVar(&c.ServerAddr, "http", c.ServerAddr, "Address in form of ip:port to listen")
	fs.StringVar(&c.TLSServerAddr, "https", c.TLSServerAddr, "Address in form of ip:port to listen")

	fs.BoolVar(&c.FastOpen, "tcp-fast-open", c.FastOpen, "Enable TCP fast open")
	fs.BoolVar(&c.Naggle, "tcp-naggle", c.Naggle, "Enable TCP Nagle's algorithm (disables NO_DELAY)")
	fs.BoolVar(&c.HTTP2, "http2", c.HTTP2, "Enable HTTP/2 when TLS is enabled")

	fs.StringVar(&c.HSTS, "hsts", c.HSTS, "Set HSTS to the value provided on all responses")
	fs.StringVar(&c.TLSCertFile, "cert", c.TLSCertFile, "X.509 certificate file for HTTPS server")
	fs.StringVar(&c.TLSKeyFile, "key", c.TLSKeyFile, "X.509 key file for HTTPS server")

	fs.BoolVar(&c.LetsEncrypt, "letsencrypt", c.LetsEncrypt, "Enable automatic TLS using letsencrypt.org")
	fs.StringVar(&c.LetsEncryptEmail, "letsencrypt-email", c.LetsEncryptEmail, "Optional email to register with letsencrypt (default is anonymous)")
	fs.StringVar(&c.LetsEncryptHosts, "letsencrypt-hosts", c.LetsEncryptHosts, "Comma separated list of hosts for the certificate (required)")
	fs.StringVar(&c.LetsEncryptCacheDir, "letsencrypt-cert-dir", c.LetsEncryptCacheDir, "Letsencrypt cert dir")

	fs.StringVar(&c.APIPrefix, "api-prefix", c.APIPrefix, "API endpoint prefix")
	fs.StringVar(&c.CORSOrigin, "cors-origin", c.CORSOrigin, "Comma separated list of CORS origins endpoints")
	fs.BoolVar(&c.UseXForwardedFor, "use-x-forwarded-for", c.UseXForwardedFor, "Use the X-Forwarded-For header when available (e.g. behind proxy)")

	fs.StringVar(&c.GuiDir, "gui", c.GuiDir, "Web gui directory")

	fs.DurationVar(&c.SessionLifetime, "session-lifetime", c.SessionLifetime, "Session lifetime of an counted visitor")
	fs.DurationVar(&c.PongWait, "pong-wait", c.PongWait, "Websocket pong wait duration")
	fs.DurationVar(&c.PingPeriod, "ping-period", c.PingPeriod, "Send pings to peer with this period. Must be less than pong-wait.")

	fs.DurationVar(&c.WriteTimeout, "write-timeout", c.WriteTimeout, "Write timeout for HTTP and HTTPS client connections")
	fs.BoolVar(&c.LogToStdout, "logtostdout", c.LogToStdout, "Log to stdout instead of stderr")
	fs.StringVar(&c.LogOutputFile, "log-file", c.LogOutputFile, "Log output file")
	fs.BoolVar(&c.LogTimestamp, "logtimestamp", c.LogTimestamp, "Prefix non-access logs with timestamp")

	fs.IntVar(&c.RateLimitBurst, "quota-burst", c.RateLimitBurst, "Max requests per source IP per request burst")
	fs.DurationVar(&c.RateLimitInterval, "quota-interval", c.RateLimitInterval, "Quota expiration interval, per source IP querying the API")
	fs.IntVar(&c.RateLimitLimit, "quota-max", c.RateLimitLimit, "Max requests per source IP per interval; set 0 to turn quotas off")

	fs.DurationVar(&c.ReadTimeout, "read-timeout", c.ReadTimeout, "Read timeout for HTTP and HTTPS client connections")

	fs.BoolVar(&c.Silent, "silent", c.Silent, "Disable HTTP and HTTPS log request details")
	fs.StringVar(&c.File, "config", c.File, "Config file")
	fs.BoolVar(&c.SaveConfigFlag, "save", c.SaveConfigFlag, "Save config")
}

func NewConfigFromFile(configFile string) *Config {
	config := DefaultConfig()

	if configFile != "" {
		config.Load(configFile)
		config.File = configFile
	}

	return config
}

func (c *Config) initFile(filename string) {
	filesystem.CreateDirectory("conf")
	if len(filename) == 0 {
		dir, _ := os.Getwd()
		filename = path.Join(dir, "conf", "settings.config")

		c.Load(filename)
	}
	c.File = filename
}

func (c *Config) Load(filename string) bool {
	c.initFile(filename)

	if _, err := os.Stat(filename); err == nil {

		content, err := ioutil.ReadFile(filename)
		if err != nil {

			if !c.Silent {
				log.Printf("[error] Config file failed to load: %s", err.Error())
			}
			return false
		}

		err = json.Unmarshal(content, c)
		if err != nil {
			if !c.Silent {
				log.Printf("[error] Config file failed to load: %s", err.Error())
			}
			return false
		}

		if !c.Silent {
			log.Printf("[info] Config file loaded successfully")
		}

	} else {
		_, _ = c.Save()
	}
	return true
}

func (c *Config) Save() (bool, error) {
	if len(c.File) == 0 {
		c.initFile("")
	}

	file, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		if !c.Silent {
			fmt.Println(err)
		}
		return false, err
	}

	err = ioutil.WriteFile(c.File, file, 0644)
	if err != nil {
		panic(err)
		return false, err
	}

	if !c.Silent {
		log.Printf("[info] Config file saved under: %s", c.File)
	}

	return true, nil
}

func (c *Config) logWriter() io.Writer {
	return c.LogOutput
}

func (c *Config) ErrorLogger() *log.Logger {
	if c.LogTimestamp {
		return log.New(c.logWriter(), "[error] ", log.LstdFlags)
	}
	return log.New(c.logWriter(), "[error] ", 0)
}

func (c *Config) AccessLogger() *log.Logger {
	return log.New(c.logWriter(), "[access] ", 0)
}
