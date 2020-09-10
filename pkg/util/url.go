package util

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateURLSuffix will parse a provided url, extract the hostname, and compare it
// to the provided domain suffix. If the extracted hostname does not have the
// expected suffix, an error is returned.
func ValidateURLSuffix(rawurl, domainSuffix string) error {
	if len(domainSuffix) == 0 || domainSuffix[0] != '.' {
		return fmt.Errorf("domain suffix must begin with '.'")
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	if u.Scheme != "https" {
		return fmt.Errorf("url must use https instead of %s", u.Scheme)
	}
	if !strings.HasSuffix(u.Host, domainSuffix) {
		return fmt.Errorf("host suffix mismatch: %s must end with %s", u.Host, domainSuffix)
	}
	return nil
}
