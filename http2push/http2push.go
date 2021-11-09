package http2push

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type Http2Push struct {
	next  http.Handler
	name  string
	debug bool
}

var (
	linkRegex = regexp.MustCompile(`.*<([^>]+)>;.*`)
)

func parseLink(l string) []string {
	links := strings.Split(l, ",")

	var paths []string
	for _, l := range links {
		groups := linkRegex.FindStringSubmatch(l)

		if len(groups) != 2 {
			log.Printf("🔴 HTTP2 push: invalid link: %s\n", l)
			continue
		}

		paths = append(paths, groups[1])
	}

	return paths
}

func pushLinks(p http.Pusher, linkHeaders []string, debug bool) {
	for _, h := range linkHeaders {
		links := parseLink(h)
		for _, link := range links {
			if debug {
				fmt.Printf("☀️ Link '%s' pushed!\n", link)
			}
			err := p.Push(link, nil)
			if err != nil {
				fmt.Printf("🔴 HTTP2 push: internal error - %s\n", err)
			}
		}
	}
}

// traefik part

type Config struct {
	Debug bool
}

func CreateConfig() *Config {
	return &Config{}
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &Http2Push{
		next:  next,
		name:  name,
		debug: config.Debug,
	}, nil
}

func (a *Http2Push) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	a.next.ServeHTTP(rw, req)

	pusher, isPushable := rw.(http.Pusher)

	if a.debug {
		fmt.Printf("🍀 Request '%s' pushable: %t\n", req.RequestURI, isPushable)
	}

	if isPushable {
		pushLinks(pusher, rw.Header()["Link"], a.debug)
	}
}
