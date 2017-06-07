package uic

import (
	"fmt"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"path/filepath"
	"strings"
	"strconv"
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
			var url, name string
			var timeoutInMillis int64 = 10000
			var err error
			switch len(args) {
			case 1:
				url = args[0]
			case 2:
				name = args[0]
				url = args[1]
			case 3:
				name = args[0]
				url = args[1]
				timeoutInMillis, err = strconv.ParseInt(args[2], 10, 64)
				if err != nil {
					return c.Err(fmt.Sprintf("Error parsing timeout: %v", err.Error()))
				}
			default:
				return c.ArgErr()
			}
			if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
				if strings.HasPrefix(url, "/") {
					url = "file://" + url
				} else {
					url = "file://" + filepath.Join(httpserver.GetConfig(c).Root, url)
				}
			}
			f := &Fetch{
				URL:  url,
				Name: name,
				Timeout: time.Duration(timeoutInMillis) * time.Millisecond,
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
