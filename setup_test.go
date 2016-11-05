package uic

import (
	"testing"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/stretchr/testify/assert"
	"github.com/tarent/lib-compose/composition"
	"time"
)

var headerDef = Fetch{
	FetchDefinition: composition.FetchDefinition{
		URL:             "file:///header.html",
		Timeout:         time.Second * 10,
		FollowRedirects: true,
		Required:        true,
		Method:          "GET",
	},
}

var footerDef = Fetch{
	FetchDefinition: composition.FetchDefinition{
		URL:             "file:///footer.html",
		Timeout:         time.Second * 10,
		FollowRedirects: true,
		Required:        true,
		Method:          "GET",
	},
}

func TestSetup(t *testing.T) {

	for j, test := range []struct {
		input           string
		shouldErr       bool
		path            string
		upstream        string
		expectedFetches []Fetch
	}{
		{input: "uic", shouldErr: true},
		{input: "uic / / xx", shouldErr: true},
		{"uic / {\n  fetch /header.html\n  fetch /footer.html\n}",
			false,
			"/",
			"file://.",
			[]Fetch{headerDef, footerDef}},
		{"uic / {\n  fetch /header.html\n  fetch footer.html\n}", // footer as relative path to root
			false,
			"/",
			"file://.",
			[]Fetch{headerDef, Fetch{
				FetchDefinition: composition.FetchDefinition{
					URL:             "file://footer.html",
					Timeout:         time.Second * 10,
					FollowRedirects: true,
					Required:        true,
					Method:          "GET",
				},
			}}},
		{"uic /somePath http://example.com/ {\n  fetch /header.html\n  fetch /footer.html\n}",
			false,
			"/somePath",
			"http://example.com/",
			[]Fetch{headerDef, footerDef}},
	} {
		c := caddy.NewTestController("http", test.input)
		err := setup(c)
		if err != nil && !test.shouldErr {
			t.Errorf("Test case #%d received an error of %v", j, err)
		} else if test.shouldErr {
			continue
		}
		mids := httpserver.GetConfig(c).Middleware()
		middleware := mids[len(mids)-1](nil).(*Uic)
		assert.Equal(t, test.path, middleware.path)
		assert.Equal(t, test.upstream, middleware.upstream)
		assert.Equal(t, test.expectedFetches, middleware.fetchRules)
	}

}
