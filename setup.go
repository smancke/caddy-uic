package uic

import (
	"fmt"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"path/filepath"
	"strings"
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
			if len(args) != 1 {
				return fmt.Errorf("Wrong number of arguments for fetch: %v (%v:%v)", args, c.File(), c.Line())
			}
			url := args[0]
			if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
				if strings.HasPrefix(url, "/") {
					url = "file://" + url
				} else {
					url = "file://" + filepath.Join(httpserver.GetConfig(c).Root, url)
				}
			}
			config.AddFetch(&Fetch{
				URL: url,
			})
		case "except":
			config.Except = args
		default:
			return fmt.Errorf("Unknown option within uic: %v (%v:%v)", c.Val(), c.File(), c.Line())
		}
	}

	return nil
}
