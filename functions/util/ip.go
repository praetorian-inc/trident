package util

import (
	"context"
	"fmt"
	"net"
	"time"
)

func ExternalIP() (string, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * 3,
			}
			return d.DialContext(ctx, "udp", "resolver1.opendns.com:53")
		},
	}
	ip, err := r.LookupHost(context.Background(), "myip.opendns.com")
	if err != nil {
		return "", err
	}

	if len(ip) == 0 {
		return "", fmt.Errorf("could not determine external ip")
	}

	return ip[0], nil
}
