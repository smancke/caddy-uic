package uic

import "time"

const DefaultTimeout = time.Second * 10

type Config struct {
	Path           string
	Upstream       string
	DefaultTimeout time.Duration
	FetchRules     []*Fetch
	Except         []string
}

type Fetch struct {
	URL     string
	Name    string
	Timeout time.Duration
}

func NewConfig(path, upstream string) *Config {
	return &Config{
		Path:           path,
		Upstream:       upstream,
		DefaultTimeout: DefaultTimeout,
		FetchRules:     []*Fetch{},
		Except:         []string{},
	}
}

func (c *Config) AddFetch(f *Fetch) {
	c.FetchRules = append(c.FetchRules, f)
}
