# boilerplate-check

This is still under active development, but this tool is intended to flag any
mismatched file headers against a provided boilerplate file.  It ignores the
year, so that old files don't become false positives.


## Install

You can install `boilerplate-check` with:

```
go get github.com/mattmoor/boilerplate-check
```

## Running

You can run `boilerplate-check` like so:

```
boilerplate-check check \
  `# This is our boilerplate file` \
  --boilerplate ./hack/boilerplate/boilerplate.go.txt \
  `# This is the file extension to check` \
  --file-extension go \
  `# This is a regular expression of paths to exclude.` \
  --exclude "(vendor|third_party)/"
```

### Example errors

Here some sample errors from our testdata directory:

```
# boilerplate-check check --boilerplate ./pkg/commands/testdata/boilerplate.mm.txt --file-extension mm
pkg/commands/testdata/missing.bad.mm:1: missing boilerplate:
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
pkg/commands/testdata/short.bad.mm:1: incomplete boilerplate
pkg/commands/testdata/typo.bad.mm:1: {[]string}[1]:
        -: "Copyright YYYY Matt Moore"
        +: "Copyright YYYY Matt More"
```

These errors are designed to be used in conjunction with
[reviewdog](https://github.com/reviewdog/reviewdog), more examples of this
information will be forthcoming.
