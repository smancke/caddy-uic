package uic

import (
	"fmt"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	caddy.RegisterPlugin("uic", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
	httpserver.RegisterDevDirective("uic", "proxy")
}

// setup configures a new Proxy middleware instance.
func setup(c *caddy.Controller) error {

	for c.Next() {
		// default upstream is the local file root
		args := c.RemainingArgs()

		if len(args) < 1 {
			return fmt.Errorf("Missing path argument for uic directive (%v:%v)", c.File(), c.Line())
		}

		var upstream string
		if len(args) == 2 {
			upstream = args[1]
		} else {
			upstream = "file://" + httpserver.GetConfig(c).Root
		}

		if len(args) > 2 {
			return fmt.Errorf("To many arguments for uic directive %q (%v:%v)", args, c.File(), c.Line())
		}

		config := NewConfig(args[0], upstream)

		err := parseConfig(c, config)
		if err != nil {
			return err
		}

		httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
			return NewUic(next, config)
		})
	}

	return nil
}

func parseConfig(c *caddy.Controller, config *Config) error {
	for c.NextBlock() {
		value := c.Val()
		args := c.RemainingArgs()
		switch value {
		case "fetch":
			f := &Fetch{
				Timeout: DefaultTimeout,
			}

			switch len(args) {
			case 1:
				f.URL = args[0]
			case 2:
				f.Name = args[0]
				f.URL = args[1]
			case 3:
				var err error
				f.Name = args[0]
				f.URL = args[1]
				f.Timeout, err = time.ParseDuration(args[2])
				if err != nil {
					return c.Err(fmt.Sprintf("Error parsing timeout: %v", err.Error()))
				}
			default:
				return c.ArgErr()
			}
			if !(strings.HasPrefix(f.URL, "http://") || strings.HasPrefix(f.URL, "https://")) {
				if strings.HasPrefix(f.URL, "/") {
					f.URL = "file://" + f.URL
				} else {
					f.URL = "file://" + filepath.Join(httpserver.GetConfig(c).Root, f.URL)
				}
			}
			config.AddFetch(f)
		case "except":
			config.Except = args
		default:
			return c.Err("Unknown option within uic")
		}
	}

	return nil
}
