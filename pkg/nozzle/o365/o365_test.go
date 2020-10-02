// Copyright 2020 Praetorian Security, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package o365

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/praetorian-inc/trident/pkg/nozzle"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UTC().UnixNano())
	v := m.Run()
	os.Exit(v)
}

type testcase struct {
	desc     string
	username string
	password string
	valid    bool
	mfa      bool
	locked   bool
}

func TestNozzle(t *testing.T) {
	noz, err := nozzle.Open("o365", map[string]string{
		"domain": "login.micrsoft.com",
	})
	if err != nil {
		t.Fatalf("unable to open nozzle: %s", err)
	}
	num := rand.Intn(4)
	username := fmt.Sprintf("test%d@tridentcontoso.onmicrosoft.com", num)
	password := fmt.Sprintf("ItsAutumn2020%d!", num)

	usernameMFA := "test5@tridentcontoso.onmicrosoft.com"
	passwordMFA := "ItsAutumn20205!"

	var testcases = []testcase{
		{
			desc:     "invalid login",
			username: username,
			password: "Invalid1!",
			valid:    false,
			mfa:      false,
			locked:   false,
		},
		{
			desc:     "valid login",
			username: username,
			password: password,
			valid:    true,
			mfa:      false,
			locked:   false,
		},
		{
			desc:     "valid login with mfa",
			username: usernameMFA,
			password: passwordMFA,
			valid:    true,
			mfa:      true,
			locked:   false,
		},
	}

	for _, test := range testcases {
		res, err := noz.Login(test.username, test.password)
		if err != nil {
			t.Errorf("error in login: %s", err)
			continue
		}
		if res.Valid != test.valid {
			t.Errorf("[%s] noz.valid was %t, expected %t", test.desc, res.Valid, test.valid)
		}
		if res.MFA != test.mfa {
			t.Errorf("[%s] noz.mfa %t, expected %t", test.desc, res.MFA, test.mfa)
		}
		if res.Locked != test.locked {
			t.Errorf("[%s] noz.locked %t, expected %t", test.desc, res.Locked, test.locked)
		}
	}

}
