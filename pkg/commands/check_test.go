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
	"os"
	"path/filepath"
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
		name: "no errors, with good regexp",
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
		name: "with typo mismatch error",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", "mm",
			"--exclude", "[^o].bad.mm",
		},
		want: denormalize(`testdata/typo.bad.mm:2: found mismatched boilerplate lines:
{[]string}[0]:
	-: "Copyright YYYY Matt Moore"
	+: "Copyright YYYY Matt More"
`),
	}, {
		name: "with whitespace mismatch error",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", "mm",
			"--exclude", "[^d].bad.mm",
		},
		want: `testdata/trimmed.bad.mm:3: found mismatched boilerplate lines:
{[]string}[0->?]:
	-: ""
	+: <non-existent>
{[]string}[4->?]:
	-: ""
	+: <non-existent>
{[]string}[6->?]:
	-: ""
	+: <non-existent>
{[]string}[?->10]:
	-: <non-existent>
	+: ""
{[]string}[?->11]:
	-: <non-existent>
	+: "// Package foo builds widgets"
{[]string}[?->12]:
	-: <non-existent>
	+: "package foo"
`,
	}, {
		name: "with http[s] mismatch error",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", "mm",
			"--exclude", "[^s].bad.mm",
		},
		want: `testdata/https.bad.mm:8: found mismatched boilerplate lines:
{[]string}[0]:
	-: "    http://www.apache.org/licenses/LICENSE-2.0"
	+: "    https://www.apache.org/licenses/LICENSE-2.0"
`,
	}, {
		name: "with tab/space mismatch error",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", "mm",
			"--exclude", "[^b].bad.mm",
		},
		want: `testdata/tab.bad.mm:8: found mismatched boilerplate lines:
{[]string}[0]:
	-: "    http://www.apache.org/licenses/LICENSE-2.0"
	+: "\thttp://www.apache.org/licenses/LICENSE-2.0"
`,
	}, {
		name: "with too short error",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", "mm",
			"--exclude", "[^t].bad.mm",
		},
		want: `testdata/short.bad.mm:1: incomplete boilerplate, missing:
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
`,
	}, {
		name: "with no header error",
		args: []string{
			"--boilerplate", "testdata/boilerplate.mm.txt",
			"--file-extension", "mm",
			"--exclude", "[^g].bad.mm",
		},
		want: denormalize(`testdata/missing.bad.mm:1: missing boilerplate:
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
`),
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

func TestCheckFix(t *testing.T) {
	tests := []struct {
		desc          string
		inputFile     string
		wantExitError bool
		wantNoChanges bool
	}{{
		desc:          "typo in copyright name",
		inputFile:     "testdata/typo.bad.mm",
		wantExitError: true,
	}, {
		desc:          "incomplete boilerplate",
		inputFile:     "testdata/short.bad.mm",
		wantExitError: true,
	}, {
		desc:          "missing boilerplate",
		inputFile:     "testdata/missing.bad.mm",
		wantExitError: true,
	}, {
		desc:          "https instead of http",
		inputFile:     "testdata/https.bad.mm",
		wantExitError: true,
	}, {
		desc:          "tab instead of spaces",
		inputFile:     "testdata/tab.bad.mm",
		wantExitError: true,
	}, {
		desc:          "correct boilerplate with old year",
		inputFile:     "testdata/old.good.mm",
		wantNoChanges: true,
	}}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Copy the test file to temp directory
			inputContent, err := os.ReadFile(tt.inputFile)
			if err != nil {
				t.Fatalf("Failed to read input file: %v", err)
			}

			testFile := filepath.Join(tmpDir, "test.mm")
			if err := os.WriteFile(testFile, inputContent, 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			originalWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get working directory: %v", err)
			}
			defer os.Chdir(originalWd)

			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}

			cmd := NewCheckCommand()
			output := new(bytes.Buffer)
			cmd.SetOut(output)
			cmd.SetArgs([]string{
				"--boilerplate", filepath.Join(originalWd, "testdata/boilerplate.mm.txt"),
				"--file-extension", "mm",
				"--fix",
			})

			err = cmd.Execute()
			if tt.wantExitError {
				if err == nil {
					t.Errorf("Execute() succeeded, wanted error. Output: %s", output.String())
				}
			} else {
				if err != nil {
					t.Errorf("Execute() failed: %v", err)
				}
			}

			// Read the fixed file
			fixedContent, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read fixed file: %v", err)
			}

			if tt.wantNoChanges {
				if string(fixedContent) != string(inputContent) {
					t.Errorf("File was modified when no changes were expected")
				}
			} else {
				// Verify the fixed file now passes validation
				cmd2 := NewCheckCommand()
				output2 := new(bytes.Buffer)
				cmd2.SetOut(output2)
				cmd2.SetArgs([]string{
					"--boilerplate", filepath.Join(originalWd, "testdata/boilerplate.mm.txt"),
					"--file-extension", "mm",
				})

				if err := cmd2.Execute(); err != nil {
					t.Errorf("Fixed file failed validation: %v\nOutput: %s", err, output2.String())
				}

				if output2.String() != "" {
					t.Errorf("Fixed file has validation errors: %s", output2.String())
				}
			}
		})
	}
}
