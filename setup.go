package uic

import (
	"fmt"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/tarent/lib-compose/composition"
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

		fetchRules, except, err := parseConfig(c)
		if err != nil {
			return err
		}

		httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
			return NewUic(next, args[0], upstream, fetchRules, except)
		})
	}

	return nil
}

func parseConfig(c *caddy.Controller) ([]Fetch, []string, error) {
	var fetchs []Fetch
	var except = []string{}

	for c.NextBlock() {
		value := c.Val()
		args := c.RemainingArgs()
		switch value {
		case "fetch":
			if len(args) != 1 {
				return fetchs, except, fmt.Errorf("Wrong number of arguments for fetch: %v (%v:%v)", args, c.File(), c.Line())
			}
			url := args[0]
			if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
				if strings.HasPrefix(url, "/") {
					url = "file://" + url
				} else {
					url = "file://" + filepath.Join(httpserver.GetConfig(c).Root, url)
				}
			}
			fetch := Fetch{FetchDefinition: composition.NewFetchDefinition(url)}
			fetchs = append(fetchs, fetch)
		case "except":
			except = args
		default:
			return fetchs, except, fmt.Errorf("Unknown option within uic: %v (%v:%v)", c.Val(), c.File(), c.Line())
		}
	}

	return fetchs, except, nil
}
