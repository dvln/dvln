// Copyright © 2015 Erik Brady <brady@dvln.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmds

import (
	"fmt"
	"testing"

	cfg "github.com/spf13/viper"
)

func TestVersion(t *testing.T) {
	cfg.Set("long", true)
	fmt.Println("Pretending to test")
}

/*
func TestFixUrl(t *testing.T) {
	type data struct {
		TestName   string
		CliBaseUrl string
		CfgBaseUrl string
		AppendPort bool
		Port       int
		Result     string
	}
	tests := []data{
		{"Basic http localhost", "", "http://foo.com", true, 1313, "http://localhost:1313/"},
		{"Basic https production, http localhost", "", "https://foo.com", true, 1313, "http://localhost:1313/"},
		{"Basic subdir", "", "http://foo.com/bar", true, 1313, "http://localhost:1313/bar/"},
		{"Basic production", "http://foo.com", "http://foo.com", false, 80, "http://foo.com/"},
		{"Production subdir", "http://foo.com/bar", "http://foo.com/bar", false, 80, "http://foo.com/bar/"},
		{"No http", "", "foo.com", true, 1313, "http://localhost:1313/"},
		{"Override configured port", "", "foo.com:2020", true, 1313, "http://localhost:1313/"},
		{"No http production", "foo.com", "foo.com", false, 80, "http://foo.com/"},
		{"No http production with port", "foo.com", "foo.com", true, 2020, "http://foo.com:2020/"},
	}

	for i, test := range tests {
		BaseUrl = test.CliBaseUrl
		cfg.Set("BaseUrl", test.CfgBaseUrl)
		serverAppend = test.AppendPort
		serverPort = test.Port
		result, err := fixUrl(BaseUrl)
		if err != nil {
			t.Errorf("Test #%d %s: unexpected error %s", err)
		}
		if result != test.Result {
			t.Errorf("Test #%d %s: expected %q, got %q", i, test.TestName, test.Result, result)
		}
	}
}
*/