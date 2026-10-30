package xhttp

import (
	"errors"
	"net/url"
	"strings"
)

func ValidateProduction(mode string, secure bool, origins string) error {
	if mode != "pro" {
		return nil
	}
	if !secure {
		return errors.New("production requires Secure authentication cookies")
	}
	for _, raw := range strings.Split(origins, ",") {
		u, err := url.Parse(strings.TrimSpace(raw))
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return errors.New("production requires explicit HTTPS browser origins")
		}
	}
	return nil
}
