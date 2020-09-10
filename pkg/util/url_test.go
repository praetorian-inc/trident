package util

import (
	"fmt"
	"testing"
)

type testcase struct {
	input     string
	domain    string
	expecterr bool
}

func TestValidation(t *testing.T) {
	var testcases = []testcase{
		{"example", ".okta.com", false},
		{"example", "okta.com", true},
		{"example.com/", ".okta.com", true},
		{"example", ".example.com", true},
	}
	for _, test := range testcases {
		url := fmt.Sprintf("https://%s.okta.com/api/v1/authn", test.input)
		err := ValidateURLSuffix(url, test.domain)
		if test.expecterr && (err == nil) {
			t.Errorf("expected validation error for (%s, %s)", test.input, test.domain)
		}
		if !test.expecterr && (err != nil) {
			t.Errorf("unexpected validation error for (%s, %s): %s", test.input, test.domain, err)
		}
	}

	testcases = []testcase{
		{"https://example.okta.com", ".okta.com", false},
		{"file:///example.okta.com", ".okta.com", true},
		{"http://example.okta.com", "okta.com", true},
	}
	for _, test := range testcases {
		err := ValidateURLSuffix(test.input, test.domain)
		if test.expecterr && (err == nil) {
			t.Errorf("expected validation error for (%s, %s)", test.input, test.domain)
		}
		if !test.expecterr && (err != nil) {
			t.Errorf("unexpected validation error for (%s, %s): %s", test.input, test.domain, err)
		}
	}

}
