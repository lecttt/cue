// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/go-quicktest/qt"
)

var checkTests = []struct {
	path    string
	version string
	ok      bool
}{
	{"rsc.io/quote@v0", "0.1.0", false},
	{"rsc io/quote", "v1.0.0", false},

	{"github.com/go-yaml/yaml@v0", "v0.8.0", true},
	{"github.com/go-yaml/yaml@v1", "v1.0.0", true},
	{"github.com/go-yaml/yaml", "v2.0.0", false},
	{"github.com/go-yaml/yaml@v1", "v2.1.5", false},
	{"github.com/go-yaml/yaml@v3.0", "v3.0.0", false},

	{"github.com/go-yaml/yaml@v2", "v1.0.0", false},
	{"github.com/go-yaml/yaml@v2", "v2.0.0", true},
	{"github.com/go-yaml/yaml@v2", "v2.1.5", true},
	{"github.com/go-yaml/yaml@v2", "v3.0.0", false},

	{"rsc.io/quote", "v17.0.0", false},
}

func TestCheck(t *testing.T) {
	for _, tt := range checkTests {
		err := Check(tt.path, tt.version)
		if tt.ok && err != nil {
			t.Errorf("Check(%q, %q) = %v, wanted nil error", tt.path, tt.version, err)
		} else if !tt.ok && err == nil {
			t.Errorf("Check(%q, %q) succeeded, wanted error", tt.path, tt.version)
		}
	}
}

var checkPathWithoutVersionTests = []struct {
	path    string
	wantErr string
}{{
	path:    "rsc io/quote",
	wantErr: `invalid char ' '`,
}, {
	path:    "foo.com@v0",
	wantErr: `module path inappropriately contains major version`,
}, {
	path: "foo.com/bar/baz",
}}

func TestCheckPathWithoutVersion(t *testing.T) {
	for _, test := range checkPathWithoutVersionTests {
		t.Run(test.path, func(t *testing.T) {
			err := CheckPathWithoutVersion(test.path)
			if test.wantErr != "" {
				qt.Assert(t, qt.ErrorMatches(err, test.wantErr))
				return
			}
			qt.Assert(t, qt.IsNil(err))
		})
	}
}

var newVersionTests = []struct {
	path, vers   string
	wantError    string
	wantPath     string
	wantBasePath string
}{{
	path:         "github.com/foo/bar@v0",
	vers:         "v0.1.2",
	wantPath:     "github.com/foo/bar@v0",
	wantBasePath: "github.com/foo/bar",
}, {
	path:         "github.com/foo/bar",
	vers:         "v3.1.2",
	wantPath:     "github.com/foo/bar@v3",
	wantBasePath: "github.com/foo/bar",
}, {
	path:         "github.com/foo/bar@v1",
	vers:         "",
	wantPath:     "github.com/foo/bar@v1",
	wantBasePath: "github.com/foo/bar",
}, {
	path:      "github.com/foo/bar@v1",
	vers:      "v3.1.2",
	wantError: `mismatched major version suffix in "github.com/foo/bar@v1" \(version v3\.1\.2\)`,
}, {
	path:      "github.com/foo/bar@v1",
	vers:      "v3.1",
	wantError: `version "v3.1" \(of module "github.com/foo/bar@v1"\) is not canonical`,
}, {
	path:      "github.com/foo/bar@v1",
	vers:      "v3.10.4+build",
	wantError: `version "v3.10.4\+build" \(of module "github.com/foo/bar@v1"\) is not canonical`,
}, {
	path:      "something/bad@v1",
	vers:      "v1.2.3",
	wantError: `malformed module path "something/bad@v1": missing dot in first path element`,
}, {
	path:      "foo.com/bar",
	vers:      "",
	wantError: `path "foo.com/bar" has no major version`,
}, {
	path:      "x.com",
	vers:      "bad",
	wantError: `version "bad" \(of module "x.com"\) is not well formed`,
}}

func TestNewVersion(t *testing.T) {
	for _, test := range newVersionTests {
		t.Run(test.path+"@"+test.vers, func(t *testing.T) {
			v, err := NewVersion(test.path, test.vers)
			if test.wantError != "" {
				qt.Assert(t, qt.ErrorMatches(err, test.wantError))
				return
			}
			qt.Assert(t, qt.IsNil(err))
			qt.Assert(t, qt.Equals(v.Path(), test.wantPath))
			qt.Assert(t, qt.Equals(v.BasePath(), test.wantBasePath))
			qt.Assert(t, qt.Equals(v.Version(), test.vers))
		})
	}
}

var parseVersionTests = []struct {
	s         string
	wantError string
}{{
	s: "github.com/foo/bar@v0.1.2",
}}

func TestParseVersion(t *testing.T) {
	for _, test := range parseVersionTests {
		t.Run(test.s, func(t *testing.T) {
			v, err := ParseVersion(test.s)
			if test.wantError != "" {
				qt.Assert(t, qt.ErrorMatches(err, test.wantError))
				return
			}
			qt.Assert(t, qt.IsNil(err))
			qt.Assert(t, qt.Equals(v.String(), test.s))
		})
	}
}

var escapeVersionTests = []struct {
	v   string
	esc string // empty means same as path
}{
	{v: "v1.2.3-alpha"},
	{v: "v3"},
	{v: "v2.3.1-ABcD", esc: "v2.3.1-!a!bc!d"},
}

func TestEscapeVersion(t *testing.T) {
	for _, tt := range escapeVersionTests {
		esc, err := EscapeVersion(tt.v)
		if err != nil {
			t.Errorf("EscapeVersion(%q): unexpected error: %v", tt.v, err)
			continue
		}
		want := tt.esc
		if want == "" {
			want = tt.v
		}
		if esc != want {
			t.Errorf("EscapeVersion(%q) = %q, want %q", tt.v, esc, want)
		}
	}
}

func TestEscapePath(t *testing.T) {
	// Check invalid paths.
	for _, tt := range checkPathWithoutVersionTests {
		if tt.wantErr != "" {
			_, err := EscapePath(tt.path)
			if err == nil {
				t.Errorf("EscapePath(%q): succeeded, want error (invalid path)", tt.path)
			}
		}
	}
	path := "foo.com/bar"
	esc, err := EscapePath(path)
	if err != nil {
		t.Fatal(err)
	}
	if esc != path {
		t.Fatalf("EscapePath(%q) = %q, want %q", path, esc, path)
	}
}

var parseImportPathTests = []struct {
	testName      string
	path          string
	want          ImportPath
	wantCanonical string
}{{
	testName: "StdlibLikeWithSlash",
	path:     "stdlib/path",
	want: ImportPath{
		Path:      "stdlib/path",
		Qualifier: "path",
	},
}, {
	testName: "StdlibLikeNoSlash",
	path:     "math",
	want: ImportPath{
		Path:      "math",
		Qualifier: "math",
	},
}, {
	testName: "StdlibLikeExplicitQualifier",
	path:     "stdlib/path:other",
	want: ImportPath{
		Path:              "stdlib/path",
		ExplicitQualifier: true,
		Qualifier:         "other",
	},
}, {
	testName: "StdlibLikeExplicitQualifierNoSlash",
	path:     "math:other",
	want: ImportPath{
		Path:              "math",
		ExplicitQualifier: true,
		Qualifier:         "other",
	},
}, {
	testName: "WithMajorVersion",
	path:     "foo.com/bar@v0",
	want: ImportPath{
		Path:      "foo.com/bar",
		Version:   "v0",
		Qualifier: "bar",
	},
}, {
	testName: "WithMajorVersionNoSlash",
	path:     "main.test@v0",
	want: ImportPath{
		Path:      "main.test",
		Version:   "v0",
		Qualifier: "main.test",
	},
}, {
	testName: "WithMajorVersionAndExplicitQualifier",
	path:     "foo.com/bar@v0:other",
	want: ImportPath{
		Path:              "foo.com/bar",
		Version:           "v0",
		ExplicitQualifier: true,
		Qualifier:         "other",
	},
}, {
	testName: "WithMajorVersionAndNoQualifier",
	path:     "foo.com/bar@v0",
	want: ImportPath{
		Path:      "foo.com/bar",
		Version:   "v0",
		Qualifier: "bar",
	},
}, {
	testName: "WithRedundantQualifier",
	path:     "foo.com/bar@v0:bar",
	want: ImportPath{
		Path:              "foo.com/bar",
		Version:           "v0",
		ExplicitQualifier: true,
		Qualifier:         "bar",
	},
	wantCanonical: "foo.com/bar@v0",
}}

func TestParseImportPath(t *testing.T) {
	for _, test := range parseImportPathTests {
		t.Run(test.testName, func(t *testing.T) {
			parts := ParseImportPath(test.path)
			qt.Assert(t, qt.DeepEquals(parts, test.want))
			qt.Assert(t, qt.Equals(parts.String(), test.path))
			if test.wantCanonical == "" {
				test.wantCanonical = test.path
			}
			qt.Assert(t, qt.Equals(parts.Canonical().String(), test.wantCanonical))
		})
	}
}
