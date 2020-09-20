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
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
)

var (
	ErrBoilerplateRequired   = errors.New("--boilerplate is a required flag.")
	ErrFileExtensionRequired = errors.New("--file-extension is a required flag.")
)

// NewCheckCommand implements the `check` sub-command
func NewCheckCommand() *cobra.Command {
	co := &checkOptions{}

	cmd := &cobra.Command{
		Use:     "check",
		Short:   "Checks that file headers match boilerplate files.",
		PreRunE: co.PreRunE,
		RunE:    co.RunE,
	}
	co.AddFlags(cmd)
	cmd.SetOut(os.Stdout)

	return cmd
}

type checkOptions struct {
	BoilerplateFile string
	FileExtension   string
	ExcludePattern  string

	boilerplateLines []string
	exclude          *regexp.Regexp
}

func (co *checkOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&co.BoilerplateFile, "boilerplate", "", "",
		"The path to the required boilerplate file.")
	cmd.Flags().StringVarP(&co.FileExtension, "file-extension", "", "",
		"The extension of files that should match this boilerplate.")
	cmd.Flags().StringVarP(&co.ExcludePattern, "exclude", "", "",
		"A pattern of files to exclude from consideration.")
}

func (co *checkOptions) PreRunE(cmd *cobra.Command, args []string) error {
	if co.BoilerplateFile == "" {
		return ErrBoilerplateRequired
	}
	bts, err := ioutil.ReadFile(co.BoilerplateFile)
	if err != nil {
		return fmt.Errorf("error reading --boilerplate file %q: %v", co.BoilerplateFile, err)
	}
	if string(bts) == "" {
		return fmt.Errorf("--boilerplate file %q is empty", co.BoilerplateFile)
	}
	raw := strings.Split(string(bts), "\n")
	co.boilerplateLines = make([]string, 0, len(raw))
	for _, rl := range raw {
		co.boilerplateLines = append(co.boilerplateLines, normalize(rl))
	}

	if co.FileExtension == "" {
		return ErrFileExtensionRequired
	}
	if strings.Contains(co.FileExtension, ".") {
		return fmt.Errorf("--file-extension %q may not contain '.'", co.FileExtension)
	}
	// filepath.Ext returns the leading "."
	co.FileExtension = "." + co.FileExtension

	if co.ExcludePattern != "" {
		co.exclude, err = regexp.Compile(co.ExcludePattern)
		if err != nil {
			return fmt.Errorf("error compiling --exclude pattern %q: %v", co.ExcludePattern, err)
		}
	}
	return nil
}

func (co *checkOptions) match(path string) bool {
	// Check whether the file extension matches.
	if ext := filepath.Ext(path); ext != co.FileExtension {
		return false
	}

	// Check whether the file is excluded by a pattern.
	if co.exclude != nil {
		if co.exclude.MatchString(path) {
			return false
		}
	}
	return true
}

func (co *checkOptions) RunE(cmd *cobra.Command, args []string) error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if !co.match(path) {
			return nil
		}

		// Open the file to copy it into the tarball.
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		// Find the first matching line of the file.
		idx, found := 1, false
		// TODO(mattmoor): Consider making the number of lines to scan a flag.
		for ; idx <= 10; idx++ {
			if !scanner.Scan() {
				break
			}
			line := normalize(scanner.Text())
			if line == co.boilerplateLines[0] {
				found = true
				break
			}
		}
		if !found {
			cmd.Printf("%s:%d: missing boilerplate:\n%s",
				path, 1, denormalize(strings.Join(co.boilerplateLines, "\n")))
			return nil
		}

		lines := make([]string, 0, len(co.boilerplateLines))
		lines = append(lines, co.boilerplateLines[0])

		for range co.boilerplateLines[1:] {
			if !scanner.Scan() {
				cmd.Printf("%s:%d: incomplete boilerplate, missing:\n%s", path, idx,
					denormalize(strings.Join(co.boilerplateLines[len(lines):], "\n")))
				return nil
			}

			lines = append(lines, normalize(scanner.Text()))
		}

		// We comment on the first bad line instead of the first line of the comment
		// because if the error is a change, and the first line of the comment block
		// isn't part of the diff, then reviewdog will filter the error.
		for i := range lines {
			if co.boilerplateLines[i] != lines[i] {
				cmd.Printf("%s:%d: found mismatched boilerplate lines:\n%s",
					path, idx+i, denormalize(cmp.Diff(co.boilerplateLines[i:], lines[i:])))
				break
			}
		}
		return nil
	})
}

// TODO(mattmoor): Fix this y10k bug.
var matchYear = regexp.MustCompile("[0-9][0-9][0-9][0-9]")

// normalize strips year-like strings out in favor of YYYY,
// so that we do not complain about older files with otherwise
// fine headers.
func normalize(line string) string {
	return matchYear.ReplaceAllString(line, "YYYY")
}

// denormalize replaces YYYY with the current year.
func denormalize(line string) string {
	return strings.ReplaceAll(line, "YYYY", fmt.Sprint(time.Now().Year()))
}
