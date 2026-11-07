package nacos

import (
	"errors"
	"fmt"
	"time"
)

const (
	DefaultPollTime = 30 * time.Second
	DefaultTimeout  = 15 * time.Second
)

// Config holds Nacos connection and target config item settings.
type Config struct {
	Address     string
	Port        uint64
	Username    string
	Password    string
	NamespaceID string
	DataID      string
	Group       string
	PollTime    time.Duration
	Timeout     time.Duration
	Scheme      string
}

// DefaultConfig returns library defaults for optional timing fields.
func DefaultConfig() Config {
	return Config{
		PollTime: DefaultPollTime,
		Timeout:  DefaultTimeout,
		Scheme:   "http",
	}
}

// ConfigDefault is an alias for DefaultConfig (naming consistency across go-* libs).
func ConfigDefault() Config { return DefaultConfig() }

// Validate checks required fields and applies defaults.
func (c *Config) Validate() error {
	if c.Address == "" || c.Port == 0 || c.Username == "" || c.Password == "" || c.NamespaceID == "" {
		return errors.New("nacos: address, port, username, password, and namespace_id are required")
	}
	if c.DataID == "" || c.Group == "" {
		return errors.New("nacos: data_id and group are required")
	}
	if c.PollTime <= 0 {
		c.PollTime = DefaultPollTime
	}
	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}
	if c.Scheme == "" {
		c.Scheme = "http"
	}
	return nil
}

func (c Config) String() string {
	return fmt.Sprintf("nacos.Config{Address:%q NamespaceID:%q DataID:%q Group:%q}", c.Address, c.NamespaceID, c.DataID, c.Group)
}
