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
	"errors"
	"fmt"
	"testing"
)

func TestCheckPreRunE(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr error
	}{{
		name:    "no args",
		wantErr: ErrBoilerplateRequired,
	}, {
		name: "boilerplate not found",
		args: []string{
			"--boilerplate", "testdata/not-found.mm.txt",
		},
		wantErr: errors.New(`error reading --boilerplate file "testdata/not-found.mm.txt": open testdata/not-found.mm.txt: no such file or directory`),
	}, {
		name: "empty boilerplate",
		args: []string{
			"--boilerplate", "testdata/empty.txt",
		},
		wantErr: errors.New(`--boilerplate file "testdata/empty.txt" is empty`),
	}, {
		name: "just boilerplate",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
		},
		wantErr: ErrFileExtensionRequired,
	}, {
		name: "with a dot",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", ".mm",
		},
		wantErr: errors.New(`--file-extension ".mm" may not contain '.'`),
	}, {
		name: "bad regexp",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", "mm",
			"--exclude", ")(",
		},
		wantErr: fmt.Errorf("error compiling --exclude pattern %q: error parsing regexp: unexpected ): `)(`", ")("),
	}, {
		name: "no errors, with mmod regexp",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", "mm",
			"--exclude", ".*.bad.mm",
		},
		wantErr: nil,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := NewCheckCommand()
			output := new(bytes.Buffer)
			cmd.SetOut(output)

			cmd.SetArgs(test.args)

			gotErr := cmd.Execute()
			if (test.wantErr != nil) != (gotErr != nil) {
				t.Errorf("Execute() = %v, wanted %v", gotErr, test.wantErr)
			} else if (test.wantErr != nil) && (gotErr != nil) {
				got, want := gotErr.Error(), test.wantErr.Error()
				if got != want {
					t.Errorf("Execute() = %s, wanted %s", got, want)
				}
			}
		})
	}
}

func TestCheckRunE(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{{
		name: "with mismatch error",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", "mm",
			"--exclude", "[^o].bad.mm",
		},
		want: `testdata/typo.bad.mm:1: {[]string}[1]:
	-: "Copyright YYYY Matt Moore"
	+: "Copyright YYYY Matt More"
`,
	}, {
		name: "with too short error",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", "mm",
			"--exclude", "[^t].bad.mm",
		},
		want: `testdata/short.bad.mm:1: incomplete boilerplate
`,
	}, {
		name: "with no header error",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", "mm",
			"--exclude", "[^g].bad.mm",
		},
		want: `testdata/missing.bad.mm:1: missing boilerplate:
/*
Copyright YYYY Matt Moore

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
`,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := NewCheckCommand()
			output := new(bytes.Buffer)
			cmd.SetOut(output)

			cmd.SetArgs(test.args)

			err := cmd.Execute()
			if err != nil {
				t.Errorf("Execute() = %v", err)
			}

			got := output.String()
			if test.want != got {
				t.Errorf("Execute() = %s, wanted %s", got, test.want)
			}
		})
	}
}
