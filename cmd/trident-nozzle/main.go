package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"trident/pkg/nozzle"

	_ "trident/pkg/nozzle/okta"
)

var (
	flagProvider     string
	flagProviderMeta string
	flagUsernames    string
	flagPasswords    string
)

func main() {
	flag.StringVar(&flagProvider, "provider", "okta", "authentication provider")
	flag.StringVar(&flagProviderMeta, "metadata", "{}", "configuration data for auth provider")
	flag.StringVar(&flagUsernames, "usernames", "-", "path to username list (or '-' for stdin)")
	flag.StringVar(&flagPasswords, "passwords", "passwords.txt", "path to password list")
	flag.Parse()

	var metadata map[string]string
	err := json.Unmarshal([]byte(flagProviderMeta), &metadata)
	if err != nil {
		log.Fatalf("error parsing provider metadata: %s", err)
	}

	if flagUsernames == "-" {
		flagUsernames = "/dev/stdin"
	}

	usernames, err := os.Open(flagUsernames) // nolint:gosec
	if err != nil {
		log.Fatalf("error reading usernames: %s", err)
	}
	defer usernames.Close() // nolint:errcheck,gosec

	content, err := ioutil.ReadFile(flagPasswords) // nolint:gosec
	if err != nil {
		log.Fatalf("error reading passwords: %s", err)
	}
	passwords := strings.Split(string(content), "\n")

	noz, err := nozzle.Open(flagProvider, metadata)
	if err != nil {
		log.Fatalf("error opening nozzle: %s", err)
	}

	var wg sync.WaitGroup

	scanner := bufio.NewScanner(usernames)
	for scanner.Scan() {
		username := scanner.Text()
		if username == "" {
			continue
		}

		for _, password := range passwords {
			if password == "" {
				continue
			}

			wg.Add(1)
			go func(username, password string) {
				res, err := noz.Login(username, password)
				if err != nil {
					log.Fatalf("error logging in: %s", err)
				}

				if res.Valid {
					fmt.Printf("%s:%s\n", username, password)
				} else {
					log.Infof("invalid credential (%s, %s)", username, password)
				}
				wg.Done()
			}(username, password)
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scanner error: %s", err)
	}

	wg.Wait()
}
