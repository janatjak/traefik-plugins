package prefixed_basic_auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Config struct {
	User           string
	Password       string
	PublicPrefixes string
	PublicSuffixes string
}

func CreateConfig() *Config {
	return &Config{}
}

type PrefixedBasicAuth struct {
	next           http.Handler
	name           string
	user           string
	password       string
	publicPrefixes []string
	publicSuffixes []string
}

// New created a new PrefixedBasicAuth plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.User == "" {
		return nil, fmt.Errorf("user is empty")
	}

	if config.Password == "" {
		return nil, fmt.Errorf("password is empty")
	}

	publicPrefixes := strings.Split(config.PublicPrefixes, ",")
	if len(publicPrefixes) == 0 || publicPrefixes[0] == "" {
		return nil, fmt.Errorf("publicPrefixes is empty")
	}

	publicSuffixes := strings.Split(config.PublicSuffixes, ",")
	if config.PublicSuffixes == "" {
		publicSuffixes = []string{}
	}

	return &PrefixedBasicAuth{
		next:           next,
		name:           name,
		user:           config.User,
		password:       config.Password,
		publicPrefixes: publicPrefixes,
		publicSuffixes: publicSuffixes,
	}, nil
}

func (a *PrefixedBasicAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match := false
	for _, prefix := range a.publicPrefixes {
		if strings.HasPrefix(req.RequestURI, "/"+prefix) {
			match = true
			break
		}
	}

	for _, suffix := range a.publicSuffixes {
		if strings.HasSuffix(req.URL.Path, suffix) {
			match = true
			break
		}
	}

	if !match {
		user, password, ok := req.BasicAuth()

		if !ok || user != a.user || password != a.password {
			rw.Header().Set("WWW-Authenticate", `Basic realm="wow"`)
			rw.WriteHeader(401)
			rw.Write([]byte("Unauthorised.\n"))
			return
		}
	}

	a.next.ServeHTTP(rw, req)
}
