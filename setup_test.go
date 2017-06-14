package uic

import (
	"testing"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/stretchr/testify/assert"
	"time"
)

var headerDef = &Fetch{
	URL:     "http://example.com/header.html",
	Timeout: 10 * time.Second,
}

var footerDef = &Fetch{
	URL:     "file:///footer.html",
	Timeout: 10 * time.Second,
}

func TestSetup(t *testing.T) {

	for j, test := range []struct {
		input          string
		shouldErr      bool
		expectedConfig *Config
	}{
		{input: "uic", shouldErr: true},
		{input: "uic / / xx", shouldErr: true},
		{"uic / {\n  fetch http://example.com/header.html\n  fetch /footer.html\n}",
			false,
			&Config{
				Path:           "/",
				Upstream:       "file://.",
				DefaultTimeout: time.Second * 10,
				FetchRules:     []*Fetch{headerDef, footerDef},
				Except:         []string{},
			},
		},
		{"uic / {\n except /foo /bar \n}",
			false,
			&Config{
				Path:           "/",
				Upstream:       "file://.",
				DefaultTimeout: time.Second * 10,
				FetchRules:     []*Fetch{},
				Except:         []string{"/foo", "/bar"},
			},
		},
		{"uic / {\n  fetch http://example.com/header.html\n  fetch footer footer.html\n}", // footer as relative path to root
			false,
			&Config{
				Path:           "/",
				Upstream:       "file://.",
				DefaultTimeout: time.Second * 10,
				FetchRules:     []*Fetch{headerDef, &Fetch{URL: "file://footer.html", Name: "footer", Timeout: 10 * time.Second}},
				Except:         []string{},
			},
		},
		{"uic /somePath http://example.com/ {\n  fetch http://example.com/header.html\n  fetch /footer.html\n}",
			false,
			&Config{
				Path:           "/somePath",
				Upstream:       "http://example.com/",
				DefaultTimeout: time.Second * 10,
				FetchRules:     []*Fetch{headerDef, footerDef},
				Except:         []string{},
			},
		},
		{"uic /somePath http://example.com/ {\n  fetch header http://example.com/header.html 5000ms\n  fetch /footer.html\n}",
			false,
			&Config{
				Path:           "/somePath",
				Upstream:       "http://example.com/",
				DefaultTimeout: time.Second * 10,
				FetchRules:     []*Fetch{{Name: "header", URL: "http://example.com/header.html", Timeout: 5000 * time.Millisecond}, footerDef},
				Except:         []string{},
			},
		},
		{"uic /somePath http://example.com/ {\n default_timeout 20s \nfetch header http://example.com/header.html \n  fetch /footer.html\n}",
			false,
			&Config{
				Path:           "/somePath",
				Upstream:       "http://example.com/",
				DefaultTimeout: time.Second * 20,
				FetchRules: []*Fetch{{Name: "header", URL: "http://example.com/header.html", Timeout: 20 * time.Second}, {
					URL:     "file:///footer.html",
					Timeout: 20 * time.Second}},
				Except: []string{},
			},
		},
		{"uic /somePath http://example.com/ {\n default_timeout 20s \nfetch header http://example.com/header.html 50s \n  fetch /footer.html\n}",
			false,
			&Config{
				Path:           "/somePath",
				Upstream:       "http://example.com/",
				DefaultTimeout: time.Second * 20,
				FetchRules: []*Fetch{{Name: "header", URL: "http://example.com/header.html", Timeout: 50 * time.Second}, {
					URL:     "file:///footer.html",
					Timeout: 20 * time.Second}},
				Except: []string{},
			},
		},
		{input: "uic /somePath http://example.com/ {\n default_timeout \nfetch header http://example.com/header.html \n  fetch /footer.html\n}", shouldErr: true},
		{input: "uic /somePath http://example.com/ {\n default_timeout 20sa \nfetch header http://example.com/header.html \n  fetch /footer.html\n}", shouldErr: true},
		{input: "uic /somePath http://example.com/ {\n  fetch header http://example.com/header.html abc\n  fetch /footer.html\n}", shouldErr: true},
	} {
		c := caddy.NewTestController("http", test.input)
		err := setup(c)
		if err != nil && !test.shouldErr {
			t.Errorf("Test case #%d received an error of %v", j, err)
		} else if test.shouldErr {
			continue
		}
		mids := httpserver.GetConfig(c).Middleware()
		if len(mids) == 0 {
			t.Errorf("no middlewares created in test #%v", j)
			continue
		}
		middleware := mids[len(mids)-1](nil).(*Uic)
		assert.Equal(t, test.expectedConfig, middleware.config)
	}
}

func TestSetupMultipleMiddlewares(t *testing.T) {

	cfg := "uic /foo {\n\n} \n\n uic /bar {\n\n}"
	c := caddy.NewTestController("http", cfg)

	initialMiddlwareCount := len(httpserver.GetConfig(c).Middleware())
	err := setup(c)
	assert.NoError(t, err)

	if err != nil {
		t.Errorf("Error seting up multiple middlewares: %v", err)
	}

	mids := httpserver.GetConfig(c).Middleware()

	assert.Equal(t, initialMiddlwareCount+2, len(mids))

	middleware1 := mids[len(mids)-2](nil).(*Uic)
	middleware2 := mids[len(mids)-1](nil).(*Uic)

	assert.Equal(t, "/foo", middleware1.config.Path)
	assert.Equal(t, "/bar", middleware2.config.Path)
}
