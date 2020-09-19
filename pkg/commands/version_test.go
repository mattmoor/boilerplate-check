/*
Copyright 2020 Matt Moore

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	cmd := NewVersionCommand()

	Version = "foo"
	BuildDate = "2020-09-19"
	GitRevision = "deadbeef"

	output := new(bytes.Buffer)
	cmd.SetOut(output)
	err := cmd.Execute()
	if err != nil {
		t.Error("Execute() =", err)
	}
	o := output.String()

	if !strings.Contains(o, Version) {
		t.Errorf("Got: %q, wanted substring: %q", o, Version)
	}
	if !strings.Contains(o, BuildDate) {
		t.Errorf("Got: %q, wanted substring: %q", o, BuildDate)
	}
	if !strings.Contains(o, GitRevision) {
		t.Errorf("Got: %q, wanted substring: %q", o, GitRevision)
	}
}
