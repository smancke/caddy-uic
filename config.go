package uic

type Config struct {
	Path       string
	Upstream   string
	FetchRules []*Fetch
	Except     []string
}

type Fetch struct {
	URL string
}

func NewConfig(path, upstream string) *Config {
	return &Config{
		Path:       path,
		Upstream:   upstream,
		FetchRules: []*Fetch{},
		Except:     []string{},
	}
}

func (c *Config) AddFetch(f *Fetch) {
	c.FetchRules = append(c.FetchRules, f)
}
