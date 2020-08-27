package okta

import (
	"testing"

	"git.praetorianlabs.com/mars/trident/functions/nozzle"
)

func TestNozzleOpen(t *testing.T) {
	_, err := nozzle.Open("okta", map[string]string{
		"domain": "praetorianlabs",
	})
	if err != nil {
		t.Fatalf("unable to open nozzle: %s", err)
	}
}

func TestInvalidLogin(t *testing.T) {
	t.Error("expected invalid login")
}

func TestValidLogin(t *testing.T) {
	t.Error("expected invalid login")
}

func TestLockout(t *testing.T) {
	t.Error("expected lockout")
}
