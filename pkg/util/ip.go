package util

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func ExternalIP() (string, error) {
	resp, err := http.Get("https://checkip.amazonaws.com")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(buf)), nil
}
