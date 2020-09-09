package util

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// ExternalIP returns the external IP address via the default route to the // Internet.
func ExternalIP() (string, error) {
	resp, err := http.Get("https://checkip.amazonaws.com")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() // nolint:errcheck

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(buf)), nil
}
