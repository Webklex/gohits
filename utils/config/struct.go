package config

import (
	"os"
	"time"
)

type Build struct {
	Number  string `json:"number"`
	Version string `json:"version"`
}

type Config struct {
	FastOpen   bool   `json:"TCP_FAST_OPEN"`
	Naggle     bool   `json:"TCP_NAGGLE"`
	ServerAddr string `json:"HTTP"`
	HTTP2      bool   `json:"HTTP2"`
	HSTS       string `json:"HSTS"`

	TLSServerAddr string `json:"HTTPS"`
	TLSCertFile   string `json:"CERT"`
	TLSKeyFile    string `json:"KEY"`

	LetsEncrypt         bool   `json:"LETSENCRYPT"`
	LetsEncryptCacheDir string `json:"LETSENCRYPT_CERT_DIR"`
	LetsEncryptEmail    string `json:"LETSENCRYPT_EMAIL"`
	LetsEncryptHosts    string `json:"LETSENCRYPT_HOSTS"`

	APIPrefix        string        `json:"API_PREFIX"`
	CORSOrigin       string        `json:"CORS_ORIGIN"`
	ReadTimeout      time.Duration `json:"READ_TIMEOUT"`
	WriteTimeout     time.Duration `json:"WRITE_TIMEOUT"`
	UseXForwardedFor bool          `json:"USE_X_FORWARDED_FOR"`
	Silent           bool          `json:"SILENT"`
	LogToStdout      bool          `json:"LOG_STDOUT"`
	LogOutputFile    string        `json:"LOG_FILE"`
	LogTimestamp     bool          `json:"LOG_TIMESTAMP"`

	RateLimitInterval time.Duration `json:"QUOTA_INTERVAL"`
	RateLimitLimit    int           `json:"QUOTA_MAX"`
	RateLimitBurst    int           `json:"QUOTA_BURST"`

	// Time allowed to write a message to the peer.
	WriteWait time.Duration `json:"WRITE_WAIT"`
	ReadWait  time.Duration `json:"READ_WAIT"`

	SessionLifetime time.Duration `json:"SESSION_LIFETIME"`

	// Maximum message size allowed from peer.
	MaxMessageSize int64 `json:"MAX_MESSAGE_SIZE"`
	// Time allowed to read the next pong message from the peer.
	PongWait time.Duration `json:"PONG_WAIT"`
	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod time.Duration `json:"PING_PERIOD"`
	// Time to wait before force close on connection.
	CloseGracePeriod time.Duration `json:"CLOSE_GRACE_PERIOD"`

	GuiDir string `json:"GUI"`

	File           string `json:"-"`
	RootDir        string `json:"-"`
	SaveConfigFlag bool   `json:"-"`
	RunSetupFlag   bool   `json:"-"`
	Build          Build  `json:"-"`

	LogOutput *os.File `json:"-"`
}
