package identify

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

type config struct {
	noBranches bool
	noTags     bool
	debug      bool

	targetURL *url.URL
}

type OptionFn func(b *identify) error

func Debug() (func(b *identify) error, error) {
	return func(b *identify) error {
		b.debug = true
		return nil
	}, nil
}

func NoTags() (func(b *identify) error, error) {
	return func(b *identify) error {
		b.noTags = true
		return nil
	}, nil
}

func NoBranches() (func(b *identify) error, error) {
	return func(b *identify) error {
		b.noBranches = true
		return nil
	}, nil
}

func ProxyURL(s string) (func(b *identify) error, error) {
	dialer := net.Dial

	if s == "" {
	} else if u, err := url.Parse(s); err != nil {
		return nil, err
	} else if v, err := proxy.FromURL(u, proxy.Direct); err != nil {
		return nil, err
	} else {
		dialer = v.Dial
	}

	return func(b *identify) error {
		b.client.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial:  dialer,
		}

		return nil
	}, nil
}

func CachePath(s string) (func(b *identify) error, error) {
	return func(b *identify) error {
		b.cachePath = s
		return nil
	}, nil
}

func UserAgent(s string) (func(b *identify) error, error) {
	return func(b *identify) error {
		// todo
		return nil
	}, nil
}

func TargetApplication(s string) (func(b *identify) error, error) {
	return func(b *identify) error {
		if application, ok := b.db.Application[s]; !ok {
			return fmt.Errorf("Application not found in rule set")
		} else {
			b.application = &application
		}

		return nil
	}, nil
}

func TargetURL(s string) (func(b *identify) error, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	return func(b *identify) error {
		b.targetURL = u
		return nil
	}, nil
}
