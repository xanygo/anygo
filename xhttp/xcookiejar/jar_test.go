// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xcookiejar

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/xanygo/anygo/xt"
)

// 说明：
// 测试代码来源于 go 标准库的，当前测试文件依赖 main_test.go
// 测试用例只是调整格式，
// 将 query.want 由 string -> []string, 最终比对时 cookie<name=value> 的顺序不要求一致

// tNow is the synthetic current time used as now during testing.
// 和 synctest 的初始时间保持一致
var tNow = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// testPSL implements PublicSuffixList with just two rules: "co.uk"
// and the default rule "*".
// The implementation has two intentional bugs:
//
//	PublicSuffix("www.buggy.psl") == "xy"
//	PublicSuffix("www2.buggy.psl") == "com"
type testPSL struct{}

func (testPSL) String() string {
	return "testPSL"
}

func (testPSL) PublicSuffix(d string) string {
	if d == "co.uk" || strings.HasSuffix(d, ".co.uk") {
		return "co.uk"
	}
	if d == "www.buggy.psl" {
		return "xy"
	}
	if d == "www2.buggy.psl" {
		return "com"
	}
	return d[strings.LastIndex(d, ".")+1:]
}

// newTestJar creates an empty Jar with testPSL as the public suffix list.
func newTestJar() *Jar {
	jar, err := New(&Options{PublicSuffixList: testPSL{}})
	if err != nil {
		panic(err)
	}
	return jar
}

var hasDotSuffixTests = [...]struct {
	s, suffix string
}{
	{s: "", suffix: ""},
	{s: "", suffix: "."},
	{s: "", suffix: "x"},
	{s: ".", suffix: ""},
	{s: ".", suffix: "."},
	{s: ".", suffix: ".."},
	{s: ".", suffix: "x"},
	{s: ".", suffix: "x."},
	{s: ".", suffix: ".x"},
	{s: ".", suffix: ".x."},
	{s: "x", suffix: ""},
	{s: "x", suffix: "."},
	{s: "x", suffix: ".."},
	{s: "x", suffix: "x"},
	{s: "x", suffix: "x."},
	{s: "x", suffix: ".x"},
	{s: "x", suffix: ".x."},
	{s: ".x", suffix: ""},
	{s: ".x", suffix: "."},
	{s: ".x", suffix: ".."},
	{s: ".x", suffix: "x"},
	{s: ".x", suffix: "x."},
	{s: ".x", suffix: ".x"},
	{s: ".x", suffix: ".x."},
	{s: "x.", suffix: ""},
	{s: "x.", suffix: "."},
	{s: "x.", suffix: ".."},
	{s: "x.", suffix: "x"},
	{s: "x.", suffix: "x."},
	{s: "x.", suffix: ".x"},
	{s: "x.", suffix: ".x."},
	{s: "com", suffix: ""},
	{s: "com", suffix: "m"},
	{s: "com", suffix: "om"},
	{s: "com", suffix: "com"},
	{s: "com", suffix: ".com"},
	{s: "com", suffix: "x.com"},
	{s: "com", suffix: "xcom"},
	{s: "com", suffix: "xorg"},
	{s: "com", suffix: "org"},
	{s: "com", suffix: "rg"},
	{s: "foo.com", suffix: ""},
	{s: "foo.com", suffix: "m"},
	{s: "foo.com", suffix: "om"},
	{s: "foo.com", suffix: "com"},
	{s: "foo.com", suffix: ".com"},
	{s: "foo.com", suffix: "o.com"},
	{s: "foo.com", suffix: "oo.com"},
	{s: "foo.com", suffix: "foo.com"},
	{s: "foo.com", suffix: ".foo.com"},
	{s: "foo.com", suffix: "x.foo.com"},
	{s: "foo.com", suffix: "xfoo.com"},
	{s: "foo.com", suffix: "xfoo.org"},
	{s: "foo.com", suffix: "foo.org"},
	{s: "foo.com", suffix: "oo.org"},
	{s: "foo.com", suffix: "o.org"},
	{s: "foo.com", suffix: ".org"},
	{s: "foo.com", suffix: "org"},
	{s: "foo.com", suffix: "rg"},
}

func TestHasDotSuffix(t *testing.T) {
	for _, tc := range hasDotSuffixTests {
		got := hasDotSuffix(tc.s, tc.suffix)
		want := strings.HasSuffix(tc.s, "."+tc.suffix)
		if got != want {
			t.Errorf("s=%q, suffix=%q: got %v, want %v", tc.s, tc.suffix, got, want)
		}
	}
}

var canonicalHostTests = map[string]string{
	"www.example.com":         "www.example.com",
	"WWW.EXAMPLE.COM":         "www.example.com",
	"wWw.eXAmple.CoM":         "www.example.com",
	"www.example.com:80":      "www.example.com",
	"192.168.0.10":            "192.168.0.10",
	"192.168.0.5:8080":        "192.168.0.5",
	"2001:4860:0:2001::68":    "2001:4860:0:2001::68",
	"[2001:4860:0:::68]:8080": "2001:4860:0:::68",
	"www.bücher.de":           "www.xn--bcher-kva.de",
	"www.example.com.":        "www.example.com",
	// TODO: Fix canonicalHost so that all of the following malformed
	// domain names trigger an error. (This list is not exhaustive, e.g.
	// malformed internationalized domain names are missing.)
	".":                       "",
	"..":                      ".",
	"...":                     "..",
	".net":                    ".net",
	".net.":                   ".net",
	"a..":                     "a.",
	"b.a..":                   "b.a.",
	"weird.stuff...":          "weird.stuff..",
	"[bad.unmatched.bracket:": "error",
}

func TestCanonicalHost(t *testing.T) {
	for h, want := range canonicalHostTests {
		got, err := canonicalHost(h)
		if want == "error" {
			if err == nil {
				t.Errorf("%q: got %q and nil error, want non-nil", h, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q: %v", h, err)
			continue
		}
		if got != want {
			t.Errorf("%q: got %q, want %q", h, got, want)
			continue
		}
	}
}

var hasPortTests = map[string]bool{
	"www.example.com":      false,
	"www.example.com:80":   true,
	"127.0.0.1":            false,
	"127.0.0.1:8080":       true,
	"2001:4860:0:2001::68": false,
	"[2001::0:::68]:80":    true,
}

func TestHasPort(t *testing.T) {
	for host, want := range hasPortTests {
		if got := hasPort(host); got != want {
			t.Errorf("%q: got %t, want %t", host, got, want)
		}
	}
}

var jarKeyTests = map[string]string{
	"foo.www.example.com": "example.com",
	"www.example.com":     "example.com",
	"example.com":         "example.com",
	"com":                 "com",
	"foo.www.bbc.co.uk":   "bbc.co.uk",
	"www.bbc.co.uk":       "bbc.co.uk",
	"bbc.co.uk":           "bbc.co.uk",
	"co.uk":               "co.uk",
	"uk":                  "uk",
	"192.168.0.5":         "192.168.0.5",
	"www.buggy.psl":       "www.buggy.psl",
	"www2.buggy.psl":      "buggy.psl",
	// The following are actual outputs of canonicalHost for
	// malformed inputs to canonicalHost (see above).
	"":              "",
	".":             ".",
	"..":            ".",
	".net":          ".net",
	"a.":            "a.",
	"b.a.":          "a.",
	"weird.stuff..": ".",
}

func TestJarKey(t *testing.T) {
	for host, want := range jarKeyTests {
		if got := jarKey(host, testPSL{}); got != want {
			t.Errorf("%q: got %q, want %q", host, got, want)
		}
	}
}

var jarKeyNilPSLTests = map[string]string{
	"foo.www.example.com": "example.com",
	"www.example.com":     "example.com",
	"example.com":         "example.com",
	"com":                 "com",
	"foo.www.bbc.co.uk":   "co.uk",
	"www.bbc.co.uk":       "co.uk",
	"bbc.co.uk":           "co.uk",
	"co.uk":               "co.uk",
	"uk":                  "uk",
	"192.168.0.5":         "192.168.0.5",
	// The following are actual outputs of canonicalHost for
	// malformed inputs to canonicalHost.
	"":              "",
	".":             ".",
	"..":            "..",
	".net":          ".net",
	"a.":            "a.",
	"b.a.":          "a.",
	"weird.stuff..": "stuff..",
}

func TestJarKeyNilPSL(t *testing.T) {
	for host, want := range jarKeyNilPSLTests {
		if got := jarKey(host, nil); got != want {
			t.Errorf("%q: got %q, want %q", host, got, want)
		}
	}
}

var isIPTests = map[string]bool{
	"127.0.0.1":            true,
	"1.2.3.4":              true,
	"2001:4860:0:2001::68": true,
	"::1%zone":             true,
	"example.com":          false,
	"1.1.1.300":            false,
	"www.foo.bar.net":      false,
	"123.foo.bar.net":      false,
}

func TestIsIP(t *testing.T) {
	for host, want := range isIPTests {
		if got := isIP(host); got != want {
			t.Errorf("%q: got %t, want %t", host, got, want)
		}
	}
}

var defaultPathTests = map[string]string{
	"/":           "/",
	"/abc":        "/",
	"/abc/":       "/abc",
	"/abc/xyz":    "/abc",
	"/abc/xyz/":   "/abc/xyz",
	"/a/b/c.html": "/a/b",
	"":            "/",
	"strange":     "/",
	"//":          "/",
	"/a//b":       "/a/",
	"/a/./b":      "/a/.",
	"/a/../b":     "/a/..",
}

func TestDefaultPath(t *testing.T) {
	for path, want := range defaultPathTests {
		if got := defaultPath(path); got != want {
			t.Errorf("%q: got %q, want %q", path, got, want)
		}
	}
}

var domainAndTypeTests = [...]struct {
	host         string // host Set-Cookie header was received from
	domain       string // domain attribute in Set-Cookie header
	wantDomain   string // expected domain of cookie
	wantHostOnly bool   // expected host-cookie flag
	wantErr      error  // expected error
}{
	{host: "www.example.com", domain: "", wantDomain: "www.example.com", wantHostOnly: true, wantErr: nil},
	{host: "127.0.0.1", domain: "", wantDomain: "127.0.0.1", wantHostOnly: true, wantErr: nil},
	{host: "2001:4860:0:2001::68", domain: "", wantDomain: "2001:4860:0:2001::68", wantHostOnly: true, wantErr: nil},
	{host: "www.example.com", domain: "example.com", wantDomain: "example.com", wantHostOnly: false, wantErr: nil},
	{host: "www.example.com", domain: ".example.com", wantDomain: "example.com", wantHostOnly: false, wantErr: nil},
	{host: "www.example.com", domain: "www.example.com", wantDomain: "www.example.com", wantHostOnly: false, wantErr: nil},
	{host: "www.example.com", domain: ".www.example.com", wantDomain: "www.example.com", wantHostOnly: false, wantErr: nil},
	{host: "foo.sso.example.com", domain: "sso.example.com", wantDomain: "sso.example.com", wantHostOnly: false, wantErr: nil},
	{host: "bar.co.uk", domain: "bar.co.uk", wantDomain: "bar.co.uk", wantHostOnly: false, wantErr: nil},
	{host: "foo.bar.co.uk", domain: ".bar.co.uk", wantDomain: "bar.co.uk", wantHostOnly: false, wantErr: nil},
	{host: "127.0.0.1", domain: "127.0.0.1", wantDomain: "127.0.0.1", wantHostOnly: true, wantErr: nil},
	{host: "2001:4860:0:2001::68", domain: "2001:4860:0:2001::68", wantDomain: "2001:4860:0:2001::68", wantHostOnly: true, wantErr: nil},
	{host: "www.example.com", domain: ".", wantDomain: "", wantHostOnly: false, wantErr: errMalformedDomain},
	{host: "www.example.com", domain: "..", wantDomain: "", wantHostOnly: false, wantErr: errMalformedDomain},
	{host: "www.example.com", domain: "other.com", wantDomain: "", wantHostOnly: false, wantErr: errIllegalDomain},
	{host: "www.example.com", domain: "com", wantDomain: "", wantHostOnly: false, wantErr: errIllegalDomain},
	{host: "www.example.com", domain: ".com", wantDomain: "", wantHostOnly: false, wantErr: errIllegalDomain},
	{host: "foo.bar.co.uk", domain: ".co.uk", wantDomain: "", wantHostOnly: false, wantErr: errIllegalDomain},
	{host: "127.www.0.0.1", domain: "127.0.0.1", wantDomain: "", wantHostOnly: false, wantErr: errIllegalDomain},
	{host: "com", domain: "", wantDomain: "com", wantHostOnly: true, wantErr: nil},
	{host: "com", domain: "com", wantDomain: "com", wantHostOnly: true, wantErr: nil},
	{host: "com", domain: ".com", wantDomain: "com", wantHostOnly: true, wantErr: nil},
	{host: "co.uk", domain: "", wantDomain: "co.uk", wantHostOnly: true, wantErr: nil},
	{host: "co.uk", domain: "co.uk", wantDomain: "co.uk", wantHostOnly: true, wantErr: nil},
	{host: "co.uk", domain: ".co.uk", wantDomain: "co.uk", wantHostOnly: true, wantErr: nil},
}

func TestDomainAndType(t *testing.T) {
	jar := newTestJar()
	for _, tc := range domainAndTypeTests {
		domain, hostOnly, err := jar.domainAndType(tc.host, tc.domain)
		if err != tc.wantErr {
			t.Errorf("%q/%q: got %q error, want %v",
				tc.host, tc.domain, err, tc.wantErr)
			continue
		}
		if err != nil {
			continue
		}
		if domain != tc.wantDomain || hostOnly != tc.wantHostOnly {
			t.Errorf("%q/%q: got %q/%t want %q/%t",
				tc.host, tc.domain, domain, hostOnly,
				tc.wantDomain, tc.wantHostOnly)
		}
	}
}

// expiresIn creates an expires attribute delta seconds from tNow.
func expiresIn(delta int) string {
	t := tNow.Add(time.Duration(delta) * time.Second)
	return "expires=" + t.Format(time.RFC1123)
}

// mustParseURL parses s to a URL and panics on error.
func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		panic(fmt.Sprintf("Unable to parse URL %s.", s))
	}
	return u
}

// jarTest encapsulates the following actions on a jar:
//  1. Perform SetCookies with fromURL and the cookies from setCookies.
//     (Done at time tNow + 0 ms.)
//  2. Check that the entries in the jar matches content.
//     (Done at time tNow + 1001 ms.)
//  3. For each query in tests: Check that Cookies with toURL yields the
//     cookies in want.
//     (Query n done at tNow + (n+2)*1001 ms.)
type jarTest struct {
	description string   // The description of what this test is supposed to test
	fromURL     string   // The full URL of the request from which Set-Cookie headers where received
	setCookies  []string // All the cookies received from fromURL
	content     string   // The whole (non-expired) content of the jar
	queries     []query  // Queries to test the Jar.Cookies method
}

// query contains one test of the cookies returned from Jar.Cookies.
type query struct {
	toURL string   // the URL in the Cookies call
	want  []string // the expected list of cookies
}

// run runs the jarTest.
func (test jarTest) run(t *testing.T, jar *Jar) {
	synctest.Test(t, func(t *testing.T) {
		test.runInSync(t, jar)
	})
}

func (test jarTest) runInSync(t *testing.T, jar *Jar) {
	// Populate jar with cookies.
	setCookies := make([]*http.Cookie, len(test.setCookies))
	for i, cs := range test.setCookies {
		cookies := (&http.Response{Header: http.Header{"Set-Cookie": {cs}}}).Cookies()
		if len(cookies) != 1 {
			panic(fmt.Sprintf("Wrong cookie line %q: %#v", cs, cookies))
		}
		setCookies[i] = cookies[0]
	}
	jar.setCookies(t.Context(), mustParseURL(test.fromURL), setCookies)
	// now = now.Add(1001 * time.Millisecond)
	time.Sleep(1001 * time.Millisecond)

	// Serialize non-expired entries in the form "name1=val1 name2=val2".
	// var cs []string
	// for _, submap := range jar.entries {
	// 	for _, cookie := range submap {
	// 		if !cookie.Expires.After(now) {
	// 			continue
	// 		}

	// 		v := cookie.Value
	// 		if strings.ContainsAny(v, " ,") || cookie.Quoted {
	// 			v = `"` + v + `"`
	// 		}
	// 		cs = append(cs, cookie.Name+"="+v)
	// 	}
	// }
	// slices.Sort(cs)
	// got := strings.Join(cs, " ")

	// // Make sure jar content matches our expectations.
	// if got != test.content {
	// 	t.Errorf("Test %q Content\ngot  %q\nwant %q",
	// 		test.description, got, test.content)
	// }

	// Test different calls to Cookies.
	for _, query := range test.queries {
		// now = now.Add(1001 * time.Millisecond)
		time.Sleep(1001 * time.Millisecond)
		var s []string
		cookies, _ := jar.cookies(t.Context(), mustParseURL(query.toURL))
		for _, c := range cookies {
			s = append(s, c.String())
		}
		xt.SortEqual(t, query.want, s, test.description)
	}
}

// basicsTests contains fundamental tests. Each jarTest has to be performed on
// a fresh, empty Jar.
var basicsTests = [...]jarTest{
	{
		description: "Retrieval of a plain host cookie.",
		fromURL:     "http://www.host.test/",
		setCookies:  []string{"A=a"},
		content:     "A=a",
		queries: []query{
			{toURL: "http://www.host.test", want: []string{"A=a"}},
			{toURL: "http://www.host.test/", want: []string{"A=a"}},
			{toURL: "http://www.host.test/some/path", want: []string{"A=a"}},
			{toURL: "https://www.host.test", want: []string{"A=a"}},
			{toURL: "https://www.host.test/", want: []string{"A=a"}},
			{toURL: "https://www.host.test/some/path", want: []string{"A=a"}},
			{toURL: "ftp://www.host.test", want: nil},
			{toURL: "ftp://www.host.test/", want: nil},
			{toURL: "ftp://www.host.test/some/path", want: nil},
			{toURL: "http://www.other.org", want: nil},
			{toURL: "http://sibling.host.test", want: nil},
			{toURL: "http://deep.www.host.test", want: nil},
		},
	},
	{
		description: "Secure cookies are not returned to http.",
		fromURL:     "http://www.host.test/",
		setCookies:  []string{"A=a; secure"},
		content:     "A=a",
		queries: []query{
			{toURL: "http://www.host.test", want: nil},
			{toURL: "http://www.host.test/", want: nil},
			{toURL: "http://www.host.test/some/path", want: nil},
			{toURL: "https://www.host.test", want: []string{"A=a"}},
			{toURL: "https://www.host.test/", want: []string{"A=a"}},
			{toURL: "https://www.host.test/some/path", want: []string{"A=a"}},
		},
	},
	{
		description: "Secure cookies are sent for localhost",
		fromURL:     "http://localhost:8910/",
		setCookies:  []string{"A=a; secure"},
		content:     "A=a",
		queries: []query{
			{toURL: "http://localhost:8910", want: []string{"A=a"}},
			{toURL: "http://localhost:8910/", want: []string{"A=a"}},
			{toURL: "http://localhost:8910/some/path", want: []string{"A=a"}},
			{toURL: "https://localhost:8910", want: []string{"A=a"}},
			{toURL: "https://localhost:8910/", want: []string{"A=a"}},
			{toURL: "https://localhost:8910/some/path", want: []string{"A=a"}},
		},
	},
	{
		description: "Secure cookies are sent for localhost (tld)",
		fromURL:     "http://example.LOCALHOST:8910/",
		setCookies:  []string{"A=a; secure"},
		content:     "A=a",
		queries: []query{
			{toURL: "http://example.LOCALHOST:8910", want: []string{"A=a"}},
			{toURL: "http://example.LOCALHOST:8910/", want: []string{"A=a"}},
			{toURL: "http://example.LOCALHOST:8910/some/path", want: []string{"A=a"}},
			{toURL: "https://example.LOCALHOST:8910", want: []string{"A=a"}},
			{toURL: "https://example.LOCALHOST:8910/", want: []string{"A=a"}},
			{toURL: "https://example.LOCALHOST:8910/some/path", want: []string{"A=a"}},
		},
	},
	{
		description: "Secure cookies are sent for localhost (ipv6)",
		fromURL:     "http://[::1]:8910/",
		setCookies:  []string{"A=a; secure"},
		content:     "A=a",
		queries: []query{
			{toURL: "http://[::1]:8910", want: []string{"A=a"}},
			{toURL: "http://[::1]:8910/", want: []string{"A=a"}},
			{toURL: "http://[::1]:8910/some/path", want: []string{"A=a"}},
			{toURL: "https://[::1]:8910", want: []string{"A=a"}},
			{toURL: "https://[::1]:8910/", want: []string{"A=a"}},
			{toURL: "https://[::1]:8910/some/path", want: []string{"A=a"}},
		},
	},
	{
		description: "Localhost only if it's a segment",
		fromURL:     "http://notlocalhost/",
		setCookies:  []string{"A=a; secure"},
		content:     "A=a",
		queries: []query{
			{toURL: "http://notlocalhost", want: nil},
			{toURL: "http://notlocalhost/", want: nil},
			{toURL: "http://notlocalhost/some/path", want: nil},
			{toURL: "https://notlocalhost", want: []string{"A=a"}},
			{toURL: "https://notlocalhost/", want: []string{"A=a"}},
			{toURL: "https://notlocalhost/some/path", want: []string{"A=a"}},
		},
	},
	{
		description: "Explicit path.",
		fromURL:     "http://www.host.test/",
		setCookies:  []string{"A=a; path=/some/path"},
		content:     "A=a",
		queries: []query{
			{toURL: "http://www.host.test", want: nil},
			{toURL: "http://www.host.test/", want: nil},
			{toURL: "http://www.host.test/some", want: nil},
			{toURL: "http://www.host.test/some/", want: nil},
			{toURL: "http://www.host.test/some/path", want: []string{"A=a"}},
			{toURL: "http://www.host.test/some/paths", want: nil},
			{toURL: "http://www.host.test/some/path/foo", want: []string{"A=a"}},
			{toURL: "http://www.host.test/some/path/foo/", want: []string{"A=a"}},
		},
	},
	{
		description: "Implicit path #1: path is a directory.",
		fromURL:     "http://www.host.test/some/path/",
		setCookies:  []string{"A=a"},
		content:     "A=a",
		queries: []query{
			{toURL: "http://www.host.test", want: nil},
			{toURL: "http://www.host.test/", want: nil},
			{toURL: "http://www.host.test/some", want: nil},
			{toURL: "http://www.host.test/some/", want: nil},
			{toURL: "http://www.host.test/some/path", want: []string{"A=a"}},
			{toURL: "http://www.host.test/some/paths", want: nil},
			{toURL: "http://www.host.test/some/path/foo", want: []string{"A=a"}},
			{toURL: "http://www.host.test/some/path/foo/", want: []string{"A=a"}},
		},
	},
	{
		description: "Implicit path #2: path is not a directory.",
		fromURL:     "http://www.host.test/some/path/index.html",
		setCookies:  []string{"A=a"},
		content:     "A=a",
		queries: []query{
			{toURL: "http://www.host.test", want: nil},
			{toURL: "http://www.host.test/", want: nil},
			{toURL: "http://www.host.test/some", want: nil},
			{toURL: "http://www.host.test/some/", want: nil},
			{toURL: "http://www.host.test/some/path", want: []string{"A=a"}},
			{toURL: "http://www.host.test/some/paths", want: nil},
			{toURL: "http://www.host.test/some/path/foo", want: []string{"A=a"}},
			{toURL: "http://www.host.test/some/path/foo/", want: []string{"A=a"}},
		},
	},
	{
		description: "Implicit path #3: no path in URL at all.",
		fromURL:     "http://www.host.test",
		setCookies:  []string{"A=a"},
		content:     "A=a",
		queries: []query{
			{toURL: "http://www.host.test", want: []string{"A=a"}},
			{toURL: "http://www.host.test/", want: []string{"A=a"}},
			{toURL: "http://www.host.test/some/path", want: []string{"A=a"}},
		},
	},
	{
		description: "Cookies are sorted by path length.",
		fromURL:     "http://www.host.test/",
		setCookies: []string{
			"A=a; path=/foo/bar",
			"B=b; path=/foo/bar/baz/qux",
			"C=c; path=/foo/bar/baz",
			"D=d; path=/foo"},
		content: "A=a B=b C=c D=d",
		queries: []query{
			{toURL: "http://www.host.test/foo/bar/baz/qux", want: []string{"B=b", "C=c", "A=a", "D=d"}},
			{toURL: "http://www.host.test/foo/bar/baz/", want: []string{"C=c", "A=a", "D=d"}},
			{toURL: "http://www.host.test/foo/bar", want: []string{"A=a", "D=d"}},
		},
	},
	{
		description: "Creation time determines sorting on same length paths.",
		fromURL:     "http://www.host.test/",
		setCookies: []string{
			"A=a; path=/foo/bar",
			"X=x; path=/foo/bar",
			"Y=y; path=/foo/bar/baz/qux",
			"B=b; path=/foo/bar/baz/qux",
			"C=c; path=/foo/bar/baz",
			"W=w; path=/foo/bar/baz",
			"Z=z; path=/foo",
			"D=d; path=/foo"},
		content: "A=a B=b C=c D=d W=w X=x Y=y Z=z",
		queries: []query{
			{toURL: "http://www.host.test/foo/bar/baz/qux", want: []string{"Y=y", "B=b", "C=c", "W=w", "A=a", "X=x", "Z=z", "D=d"}},
			{toURL: "http://www.host.test/foo/bar/baz/", want: []string{"C=c", "W=w", "A=a", "X=x", "Z=z", "D=d"}},
			{toURL: "http://www.host.test/foo/bar", want: []string{"A=a", "X=x", "Z=z", "D=d"}},
		},
	},
	{
		description: "Sorting of same-name cookies.",
		fromURL:     "http://www.host.test/",
		setCookies: []string{
			"A=1; path=/",
			"A=2; path=/path",
			"A=3; path=/quux",
			"A=4; path=/path/foo",
			"A=5; domain=.host.test; path=/path",
			"A=6; domain=.host.test; path=/quux",
			"A=7; domain=.host.test; path=/path/foo",
		},
		content: "A=1 A=2 A=3 A=4 A=5 A=6 A=7",
		queries: []query{
			{toURL: "http://www.host.test/path", want: []string{"A=2", "A=5", "A=1"}},
			{toURL: "http://www.host.test/path/foo", want: []string{"A=4", "A=7", "A=2", "A=5", "A=1"}},
		},
	},
	{
		description: "Disallow domain cookie on public suffix.",
		fromURL:     "http://www.bbc.co.uk",
		setCookies: []string{
			"a=1",
			"b=2; domain=co.uk",
		},
		content: "a=1",
		queries: []query{{toURL: "http://www.bbc.co.uk", want: []string{"a=1"}}},
	},
	{
		description: "Host cookie on IP.",
		fromURL:     "http://192.168.0.10",
		setCookies:  []string{"a=1"},
		content:     "a=1",
		queries:     []query{{toURL: "http://192.168.0.10", want: []string{"a=1"}}},
	},
	{
		description: "Domain cookies on IP.",
		fromURL:     "http://192.168.0.10",
		setCookies: []string{
			"a=1; domain=192.168.0.10",  // allowed
			"b=2; domain=172.31.9.9",    // rejected, can't set cookie for other IP
			"c=3; domain=.192.168.0.10", // rejected like in most browsers
		},
		content: "a=1",
		queries: []query{
			{toURL: "http://192.168.0.10", want: []string{"a=1"}},
			{toURL: "http://172.31.9.9", want: nil},
			{toURL: "http://www.fancy.192.168.0.10", want: nil},
		},
	},
	{
		description: "Port is ignored #1.",
		fromURL:     "http://www.host.test/",
		setCookies:  []string{"a=1"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://www.host.test", want: []string{"a=1"}},
			{toURL: "http://www.host.test:8080/", want: []string{"a=1"}},
		},
	},
	{
		description: "Port is ignored #2.",
		fromURL:     "http://www.host.test:8080/",
		setCookies:  []string{"a=1"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://www.host.test", want: []string{"a=1"}},
			{toURL: "http://www.host.test:8080/", want: []string{"a=1"}},
			{toURL: "http://www.host.test:1234/", want: []string{"a=1"}},
		},
	},
	{
		description: "IPv6 zone is not treated as a host.",
		fromURL:     "https://example.com/",
		setCookies:  []string{"a=1"},
		content:     "a=1",
		queries: []query{
			{toURL: "https://[::1%25.example.com]:80/", want: nil},
		},
	},
	{
		description: "Retrieval of cookies with quoted values", // issue #46443
		fromURL:     "http://www.host.test/",
		setCookies: []string{
			`cookie-1="quoted"`,
			`cookie-2="quoted with spaces"`,
			`cookie-3="quoted,with,commas"`,
			`cookie-4= ,`,
		},
		content: `cookie-1="quoted" cookie-2="quoted with spaces" cookie-3="quoted,with,commas" cookie-4=" ,"`,
		queries: []query{
			{
				toURL: "http://www.host.test",
				want: []string{
					`cookie-1="quoted"`,
					`cookie-2="quoted with spaces"`,
					`cookie-3="quoted,with,commas"`,
					`cookie-4=" ,"`,
				},
			},
		},
	},
}

func TestBasics(t *testing.T) {
	for idx, test := range basicsTests {
		t.Run(fmt.Sprintf("case %d", idx), func(t *testing.T) {
			jar := newTestJar()
			test.run(t, jar)
		})
	}
}

// updateAndDeleteTests contains jarTests which must be performed on the same
// Jar.
var updateAndDeleteTests = [...]jarTest{
	{
		description: "Set initial cookies.",
		fromURL:     "http://www.host.test",
		setCookies: []string{
			"a=1",
			"b=2; secure",
			"c=3; httponly",
			"d=4; secure; httponly"},
		content: "a=1 b=2 c=3 d=4",
		queries: []query{
			{toURL: "http://www.host.test", want: []string{"a=1", "c=3"}},
			{toURL: "https://www.host.test", want: []string{"a=1", "b=2", "c=3", "d=4"}},
		},
	},
	{
		description: "Update value via http.",
		fromURL:     "http://www.host.test",
		setCookies: []string{
			"a=w",
			"b=x; secure",
			"c=y; httponly",
			"d=z; secure; httponly"},
		content: "a=w b=x c=y d=z",
		queries: []query{
			{toURL: "http://www.host.test", want: []string{"a=w", "c=y"}},
			{toURL: "https://www.host.test", want: []string{"a=w", "b=x", "c=y", "d=z"}},
		},
	},
	{
		description: "Clear Secure flag from an http.",
		fromURL:     "http://www.host.test/",
		setCookies: []string{
			"b=xx",
			"d=zz; httponly",
		},
		content: "a=w b=xx c=y d=zz",
		queries: []query{{toURL: "http://www.host.test", want: []string{"a=w", "b=xx", "c=y", "d=zz"}}},
	},
	{
		description: "Delete all.",
		fromURL:     "http://www.host.test/",
		setCookies: []string{
			"a=1; max-Age=-1",                    // delete via MaxAge
			"b=2; " + expiresIn(-10),             // delete via Expires
			"c=2; max-age=-1; " + expiresIn(-10), // delete via both
			"d=4; max-age=-1; " + expiresIn(10)}, // MaxAge takes precedence
		content: "",
		queries: []query{{toURL: "http://www.host.test", want: nil}},
	},
	{
		description: "Refill #1.",
		fromURL:     "http://www.host.test",
		setCookies: []string{
			"A=1",
			"A=2; path=/foo",
			"A=3; domain=.host.test",
			"A=4; path=/foo; domain=.host.test"},
		content: "A=1 A=2 A=3 A=4",
		queries: []query{{toURL: "http://www.host.test/foo", want: []string{"A=2", "A=4", "A=1", "A=3"}}},
	},
	{
		description: "Refill #2.",
		fromURL:     "http://www.google.com",
		setCookies: []string{
			"A=6",
			"A=7; path=/foo",
			"A=8; domain=.google.com",
			"A=9; path=/foo; domain=.google.com"},
		content: "A=1 A=2 A=3 A=4 A=6 A=7 A=8 A=9",
		queries: []query{
			{toURL: "http://www.host.test/foo", want: []string{"A=2", "A=4", "A=1", "A=3"}},
			{toURL: "http://www.google.com/foo", want: []string{"A=7", "A=9", "A=6", "A=8"}},
		},
	},
	{
		description: "Delete A7.",
		fromURL:     "http://www.google.com",
		setCookies:  []string{"A=; path=/foo; max-age=-1"},
		content:     "A=1 A=2 A=3 A=4 A=6 A=8 A=9",
		queries: []query{
			{toURL: "http://www.host.test/foo", want: []string{"A=2", "A=4", "A=1", "A=3"}},
			{toURL: "http://www.google.com/foo", want: []string{"A=9", "A=6", "A=8"}},
		},
	},
	{
		description: "Delete A4.",
		fromURL:     "http://www.host.test",
		setCookies:  []string{"A=; path=/foo; domain=host.test; max-age=-1"},
		content:     "A=1 A=2 A=3 A=6 A=8 A=9",
		queries: []query{
			{toURL: "http://www.host.test/foo", want: []string{"A=2", "A=1", "A=3"}},
			{toURL: "http://www.google.com/foo", want: []string{"A=9", "A=6", "A=8"}},
		},
	},
	{
		description: "Delete A6.",
		fromURL:     "http://www.google.com",
		setCookies:  []string{"A=; max-age=-1"},
		content:     "A=1 A=2 A=3 A=8 A=9",
		queries: []query{
			{toURL: "http://www.host.test/foo", want: []string{"A=2", "A=1", "A=3"}},
			{toURL: "http://www.google.com/foo", want: []string{"A=9", "A=8"}},
		},
	},
	{
		description: "Delete A3.",
		fromURL:     "http://www.host.test",
		setCookies:  []string{"A=; domain=host.test; max-age=-1"},
		content:     "A=1 A=2 A=8 A=9",
		queries: []query{
			{toURL: "http://www.host.test/foo", want: []string{"A=2", "A=1"}},
			{toURL: "http://www.google.com/foo", want: []string{"A=9", "A=8"}},
		},
	},
	{
		description: "No cross-domain delete.",
		fromURL:     "http://www.host.test",
		setCookies: []string{
			"A=; domain=google.com; max-age=-1",
			"A=; path=/foo; domain=google.com; max-age=-1",
		},
		content: "A=1 A=2 A=8 A=9",
		queries: []query{
			{toURL: "http://www.host.test/foo", want: []string{"A=2", "A=1"}},
			{toURL: "http://www.google.com/foo", want: []string{"A=9", "A=8"}},
		},
	},
	{
		description: "Delete A8 and A9.",
		fromURL:     "http://www.google.com",
		setCookies: []string{
			"A=; domain=google.com; max-age=-1",
			"A=; path=/foo; domain=google.com; max-age=-1",
		},
		content: "A=1 A=2",
		queries: []query{
			{toURL: "http://www.host.test/foo", want: []string{"A=2", "A=1"}},
			{toURL: "http://www.google.com/foo", want: nil},
		},
	},
}

func TestUpdateAndDelete(t *testing.T) {
	jar := newTestJar()
	for idx, test := range updateAndDeleteTests {
		t.Run(fmt.Sprintf("case %d", idx), func(t *testing.T) {
			test.run(t, jar)
		})
	}
}

func TestExpiration(t *testing.T) {
	jar := newTestJar()
	jarTest{
		description: "Expiration.",
		fromURL:     "http://www.host.test",
		setCookies: []string{
			"a=1",
			"b=2; max-age=3",
			"c=3; " + expiresIn(3),
			"d=4; max-age=5",
			"e=5; " + expiresIn(5),
			"f=6; max-age=100",
		},
		content: "a=1 b=2 c=3 d=4 e=5 f=6", // executed at t0 + 1001 ms
		queries: []query{
			{toURL: "http://www.host.test", want: []string{"a=1", "b=2", "c=3", "d=4", "e=5", "f=6"}}, // t0 + 2002 ms
			{toURL: "http://www.host.test", want: []string{"a=1", "d=4", "e=5", "f=6"}},               // t0 + 3003 ms
			{toURL: "http://www.host.test", want: []string{"a=1", "d=4", "e=5", "f=6"}},               // t0 + 4004 ms
			{toURL: "http://www.host.test", want: []string{"a=1", "f=6"}},                             // t0 + 5005 ms
			{toURL: "http://www.host.test", want: []string{"a=1", "f=6"}},                             // t0 + 6006 ms
		},
	}.run(t, jar)
}

//
// Tests derived from Chromium's cookie_store_unittest.h.
//

// See http://src.chromium.org/viewvc/chrome/trunk/src/net/cookies/cookie_store_unittest.h?revision=159685&content-type=text/plain
// Some of the original tests are in a bad condition (e.g.
// DomainWithTrailingDotTest) or are not RFC 6265 conforming (e.g.
// TestNonDottedAndTLD #1 and #6) and have not been ported.

// chromiumBasicsTests contains fundamental tests. Each jarTest has to be
// performed on a fresh, empty Jar.
var chromiumBasicsTests = [...]jarTest{
	{
		description: "DomainWithTrailingDotTest.",
		fromURL:     "http://www.google.com/",
		setCookies: []string{
			"a=1; domain=.www.google.com.",
			"b=2; domain=.www.google.com..",
		},
		content: "",
		queries: []query{
			{toURL: "http://www.google.com", want: nil},
		},
	},
	{
		description: "ValidSubdomainTest #1.",
		fromURL:     "http://a.b.c.d.com",
		setCookies: []string{
			"a=1; domain=.a.b.c.d.com",
			"b=2; domain=.b.c.d.com",
			"c=3; domain=.c.d.com",
			"d=4; domain=.d.com"},
		content: "a=1 b=2 c=3 d=4",
		queries: []query{
			{toURL: "http://a.b.c.d.com", want: []string{"a=1", "b=2", "c=3", "d=4"}},
			{toURL: "http://b.c.d.com", want: []string{"b=2", "c=3", "d=4"}},
			{toURL: "http://c.d.com", want: []string{"c=3", "d=4"}},
			{toURL: "http://d.com", want: []string{"d=4"}},
		},
	},
	{
		description: "ValidSubdomainTest #2.",
		fromURL:     "http://a.b.c.d.com",
		setCookies: []string{
			"a=1; domain=.a.b.c.d.com",
			"b=2; domain=.b.c.d.com",
			"c=3; domain=.c.d.com",
			"d=4; domain=.d.com",
			"X=bcd; domain=.b.c.d.com",
			"X=cd; domain=.c.d.com"},
		content: "X=bcd X=cd a=1 b=2 c=3 d=4",
		queries: []query{
			{toURL: "http://b.c.d.com", want: []string{"b=2", "c=3", "d=4", "X=bcd", "X=cd"}},
			{toURL: "http://c.d.com", want: []string{"c=3", "d=4", "X=cd"}},
		},
	},
	{
		description: "InvalidDomainTest #1.",
		fromURL:     "http://foo.bar.com",
		setCookies: []string{
			"a=1; domain=.yo.foo.bar.com",
			"b=2; domain=.foo.com",
			"c=3; domain=.bar.foo.com",
			"d=4; domain=.foo.bar.com.net",
			"e=5; domain=ar.com",
			"f=6; domain=.",
			"g=7; domain=/",
			"h=8; domain=http://foo.bar.com",
			"i=9; domain=..foo.bar.com",
			"j=10; domain=..bar.com",
			"k=11; domain=.foo.bar.com?blah",
			"l=12; domain=.foo.bar.com/blah",
			"m=12; domain=.foo.bar.com:80",
			"n=14; domain=.foo.bar.com:",
			"o=15; domain=.foo.bar.com#sup",
		},
		content: "", // Jar is empty.
		queries: []query{{toURL: "http://foo.bar.com", want: nil}},
	},
	{
		description: "InvalidDomainTest #2.",
		fromURL:     "http://foo.com.com",
		setCookies:  []string{"a=1; domain=.foo.com.com.com"},
		content:     "",
		queries:     []query{{toURL: "http://foo.bar.com", want: nil}},
	},
	{
		description: "DomainWithoutLeadingDotTest #1.",
		fromURL:     "http://manage.hosted.filefront.com",
		setCookies:  []string{"a=1; domain=filefront.com"},
		content:     "a=1",
		queries:     []query{{toURL: "http://www.filefront.com", want: []string{"a=1"}}},
	},
	{
		description: "DomainWithoutLeadingDotTest #2.",
		fromURL:     "http://www.google.com",
		setCookies:  []string{"a=1; domain=www.google.com"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://www.google.com", want: []string{"a=1"}},
			{toURL: "http://sub.www.google.com", want: []string{"a=1"}},
			{toURL: "http://something-else.com", want: nil},
		},
	},
	{
		description: "CaseInsensitiveDomainTest.",
		fromURL:     "http://www.google.com",
		setCookies: []string{
			"a=1; domain=.GOOGLE.COM",
			"b=2; domain=.www.gOOgLE.coM",
		},
		content: "a=1 b=2",
		queries: []query{{toURL: "http://www.google.com", want: []string{"a=1", "b=2"}}},
	},
	{
		description: "TestIpAddress #1.",
		fromURL:     "http://1.2.3.4/foo",
		setCookies:  []string{"a=1; path=/"},
		content:     "a=1",
		queries:     []query{{toURL: "http://1.2.3.4/foo", want: []string{"a=1"}}},
	},
	{
		description: "TestIpAddress #2.",
		fromURL:     "http://1.2.3.4/foo",
		setCookies: []string{
			"a=1; domain=.1.2.3.4",
			"b=2; domain=.3.4",
		},
		content: "",
		queries: []query{{toURL: "http://1.2.3.4/foo", want: nil}},
	},
	{
		description: "TestIpAddress #3.",
		fromURL:     "http://1.2.3.4/foo",
		setCookies:  []string{"a=1; domain=1.2.3.3"},
		content:     "",
		queries:     []query{{toURL: "http://1.2.3.4/foo", want: nil}},
	},
	{
		description: "TestIpAddress #4.",
		fromURL:     "http://1.2.3.4/foo",
		setCookies:  []string{"a=1; domain=1.2.3.4"},
		content:     "a=1",
		queries:     []query{{toURL: "http://1.2.3.4/foo", want: []string{"a=1"}}},
	},
	{
		description: "TestNonDottedAndTLD #2.",
		fromURL:     "http://com./index.html",
		setCookies:  []string{"a=1"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://com./index.html", want: []string{"a=1"}},
			{toURL: "http://no-cookies.com./index.html", want: nil},
		},
	},
	{
		description: "TestNonDottedAndTLD #3.",
		fromURL:     "http://a.b",
		setCookies: []string{
			"a=1; domain=.b",
			"b=2; domain=b",
		},
		content: "",
		queries: []query{{toURL: "http://bar.foo", want: nil}},
	},
	{
		description: "TestNonDottedAndTLD #4.",
		fromURL:     "http://google.com",
		setCookies: []string{
			"a=1; domain=.com",
			"b=2; domain=com",
		},
		content: "",
		queries: []query{{toURL: "http://google.com", want: nil}},
	},
	{
		description: "TestNonDottedAndTLD #5.",
		fromURL:     "http://google.co.uk",
		setCookies: []string{
			"a=1; domain=.co.uk",
			"b=2; domain=.uk",
		},
		content: "",
		queries: []query{
			{toURL: "http://google.co.uk", want: nil},
			{toURL: "http://else.co.com", want: nil},
			{toURL: "http://else.uk", want: nil},
		},
	},
	{
		description: "TestHostEndsWithDot.",
		fromURL:     "http://www.google.com",
		setCookies: []string{
			"a=1",
			"b=2; domain=.www.google.com.",
		},
		content: "a=1",
		queries: []query{{toURL: "http://www.google.com", want: []string{"a=1"}}},
	},
	{
		description: "PathTest",
		fromURL:     "http://www.google.izzle",
		setCookies:  []string{"a=1; path=/wee"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://www.google.izzle/wee", want: []string{"a=1"}},
			{toURL: "http://www.google.izzle/wee/", want: []string{"a=1"}},
			{toURL: "http://www.google.izzle/wee/war", want: []string{"a=1"}},
			{toURL: "http://www.google.izzle/wee/war/more/more", want: []string{"a=1"}},
			{toURL: "http://www.google.izzle/weehee", want: nil},
			{toURL: "http://www.google.izzle/", want: nil},
		},
	},
}

func TestChromiumBasics(t *testing.T) {
	for idx, test := range chromiumBasicsTests {
		t.Run(fmt.Sprintf("case %d", idx), func(t *testing.T) {
			jar := newTestJar()
			test.run(t, jar)
		})
	}
}

// chromiumDomainTests contains jarTests which must be executed all on the
// same Jar.
var chromiumDomainTests = [...]jarTest{
	{
		description: "Fill #1.",
		fromURL:     "http://www.google.izzle",
		setCookies:  []string{"A=B"},
		content:     "A=B",
		queries:     []query{{toURL: "http://www.google.izzle", want: []string{"A=B"}}},
	},
	{
		description: "Fill #2.",
		fromURL:     "http://www.google.izzle",
		setCookies:  []string{"C=D; domain=.google.izzle"},
		content:     "A=B C=D",
		queries:     []query{{toURL: "http://www.google.izzle", want: []string{"A=B", "C=D"}}},
	},
	{
		description: "Verify A is a host cookie and not accessible from subdomain.",
		fromURL:     "http://unused.nil",
		setCookies:  []string{},
		content:     "A=B C=D",
		queries:     []query{{toURL: "http://foo.www.google.izzle", want: []string{"C=D"}}},
	},
	{
		description: "Verify domain cookies are found on proper domain.",
		fromURL:     "http://www.google.izzle",
		setCookies:  []string{"E=F; domain=.www.google.izzle"},
		content:     "A=B C=D E=F",
		queries:     []query{{toURL: "http://www.google.izzle", want: []string{"A=B", "C=D", "E=F"}}},
	},
	{
		description: "Leading dots in domain attributes are optional.",
		fromURL:     "http://www.google.izzle",
		setCookies:  []string{"G=H; domain=www.google.izzle"},
		content:     "A=B C=D E=F G=H",
		queries:     []query{{toURL: "http://www.google.izzle", want: []string{"A=B", "C=D", "E=F", "G=H"}}},
	},
	{
		description: "Verify domain enforcement works #1.",
		fromURL:     "http://www.google.izzle",
		setCookies:  []string{"K=L; domain=.bar.www.google.izzle"},
		content:     "A=B C=D E=F G=H",
		queries:     []query{{toURL: "http://bar.www.google.izzle", want: []string{"C=D", "E=F", "G=H"}}},
	},
	{
		description: "Verify domain enforcement works #2.",
		fromURL:     "http://unused.nil",
		setCookies:  []string{},
		content:     "A=B C=D E=F G=H",
		queries:     []query{{toURL: "http://www.google.izzle", want: []string{"A=B", "C=D", "E=F", "G=H"}}},
	},
}

func TestChromiumDomain(t *testing.T) {
	jar := newTestJar()
	for idx, test := range chromiumDomainTests {
		t.Run(fmt.Sprintf("case %d", idx), func(t *testing.T) {
			test.run(t, jar)
		})
	}
}

// chromiumDeletionTests must be performed all on the same Jar.
var chromiumDeletionTests = [...]jarTest{
	{
		description: "Create session cookie a1.",
		fromURL:     "http://www.google.com",
		setCookies:  []string{"a=1"},
		content:     "a=1",
		queries:     []query{{toURL: "http://www.google.com", want: []string{"a=1"}}},
	},
	{
		description: "Delete sc a1 via MaxAge.",
		fromURL:     "http://www.google.com",
		setCookies:  []string{"a=1; max-age=-1"},
		content:     "",
		queries:     []query{{toURL: "http://www.google.com", want: nil}},
	},
	{
		description: "Create session cookie b2.",
		fromURL:     "http://www.google.com",
		setCookies:  []string{"b=2"},
		content:     "b=2",
		queries:     []query{{toURL: "http://www.google.com", want: []string{"b=2"}}},
	},
	{
		description: "Delete sc b2 via Expires.",
		fromURL:     "http://www.google.com",
		setCookies:  []string{"b=2; " + expiresIn(-10)},
		content:     "",
		queries:     []query{{toURL: "http://www.google.com", want: nil}},
	},
	{
		description: "Create persistent cookie c3.",
		fromURL:     "http://www.google.com",
		setCookies:  []string{"c=3; max-age=3600"},
		content:     "c=3",
		queries:     []query{{toURL: "http://www.google.com", want: []string{"c=3"}}},
	},
	{
		description: "Delete pc c3 via MaxAge.",
		fromURL:     "http://www.google.com",
		setCookies:  []string{"c=3; max-age=-1"},
		content:     "",
		queries:     []query{{toURL: "http://www.google.com", want: nil}},
	},
	{
		description: "Create persistent cookie d4.",
		fromURL:     "http://www.google.com",
		setCookies:  []string{"d=4; max-age=3600"},
		content:     "d=4",
		queries:     []query{{toURL: "http://www.google.com", want: []string{"d=4"}}},
	},
	{
		description: "Delete pc d4 via Expires.",
		fromURL:     "http://www.google.com",
		setCookies:  []string{"d=4; " + expiresIn(-10)},
		content:     "",
		queries:     []query{{toURL: "http://www.google.com", want: nil}},
	},
}

func TestChromiumDeletion(t *testing.T) {
	jar := newTestJar()
	for idx, test := range chromiumDeletionTests {
		t.Run(fmt.Sprintf("case %d", idx), func(t *testing.T) {
			test.run(t, jar)
		})
	}
}

// domainHandlingTests tests and documents the rules for domain handling.
// Each test must be performed on an empty new Jar.
var domainHandlingTests = [...]jarTest{
	{
		description: "Host cookie",
		fromURL:     "http://www.host.test",
		setCookies:  []string{"a=1"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://www.host.test", want: []string{"a=1"}},
			{toURL: "http://host.test", want: nil},
			{toURL: "http://bar.host.test", want: nil},
			{toURL: "http://foo.www.host.test", want: nil},
			{toURL: "http://other.test", want: nil},
			{toURL: "http://test", want: nil},
		},
	},
	{
		description: "Domain cookie #1",
		fromURL:     "http://www.host.test",
		setCookies:  []string{"a=1; domain=host.test"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://www.host.test", want: []string{"a=1"}},
			{toURL: "http://host.test", want: []string{"a=1"}},
			{toURL: "http://bar.host.test", want: []string{"a=1"}},
			{toURL: "http://foo.www.host.test", want: []string{"a=1"}},
			{toURL: "http://other.test", want: nil},
			{toURL: "http://test", want: nil},
		},
	},
	{
		description: "Domain cookie #2",
		fromURL:     "http://www.host.test",
		setCookies:  []string{"a=1; domain=.host.test"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://www.host.test", want: []string{"a=1"}},
			{toURL: "http://host.test", want: []string{"a=1"}},
			{toURL: "http://bar.host.test", want: []string{"a=1"}},
			{toURL: "http://foo.www.host.test", want: []string{"a=1"}},
			{toURL: "http://other.test", want: nil},
			{toURL: "http://test", want: nil},
		},
	},
	{
		description: "Host cookie on IDNA domain #1",
		fromURL:     "http://www.bücher.test",
		setCookies:  []string{"a=1"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://www.bücher.test", want: []string{"a=1"}},
			{toURL: "http://www.xn--bcher-kva.test", want: []string{"a=1"}},
			{toURL: "http://bücher.test", want: nil},
			{toURL: "http://xn--bcher-kva.test", want: nil},
			{toURL: "http://bar.bücher.test", want: nil},
			{toURL: "http://bar.xn--bcher-kva.test", want: nil},
			{toURL: "http://foo.www.bücher.test", want: nil},
			{toURL: "http://foo.www.xn--bcher-kva.test", want: nil},
			{toURL: "http://other.test", want: nil},
			{toURL: "http://test", want: nil},
		},
	},
	{
		description: "Host cookie on IDNA domain #2",
		fromURL:     "http://www.xn--bcher-kva.test",
		setCookies:  []string{"a=1"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://www.bücher.test", want: []string{"a=1"}},
			{toURL: "http://www.xn--bcher-kva.test", want: []string{"a=1"}},
			{toURL: "http://bücher.test", want: nil},
			{toURL: "http://xn--bcher-kva.test", want: nil},
			{toURL: "http://bar.bücher.test", want: nil},
			{toURL: "http://bar.xn--bcher-kva.test", want: nil},
			{toURL: "http://foo.www.bücher.test", want: nil},
			{toURL: "http://foo.www.xn--bcher-kva.test", want: nil},
			{toURL: "http://other.test", want: nil},
			{toURL: "http://test", want: nil},
		},
	},
	{
		description: "Domain cookie on IDNA domain #1",
		fromURL:     "http://www.bücher.test",
		setCookies:  []string{"a=1; domain=xn--bcher-kva.test"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://www.bücher.test", want: []string{"a=1"}},
			{toURL: "http://www.xn--bcher-kva.test", want: []string{"a=1"}},
			{toURL: "http://bücher.test", want: []string{"a=1"}},
			{toURL: "http://xn--bcher-kva.test", want: []string{"a=1"}},
			{toURL: "http://bar.bücher.test", want: []string{"a=1"}},
			{toURL: "http://bar.xn--bcher-kva.test", want: []string{"a=1"}},
			{toURL: "http://foo.www.bücher.test", want: []string{"a=1"}},
			{toURL: "http://foo.www.xn--bcher-kva.test", want: []string{"a=1"}},
			{toURL: "http://other.test", want: nil},
			{toURL: "http://test", want: nil},
		},
	},
	{
		description: "Domain cookie on IDNA domain #2",
		fromURL:     "http://www.xn--bcher-kva.test",
		setCookies:  []string{"a=1; domain=xn--bcher-kva.test"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://www.bücher.test", want: []string{"a=1"}},
			{toURL: "http://www.xn--bcher-kva.test", want: []string{"a=1"}},
			{toURL: "http://bücher.test", want: []string{"a=1"}},
			{toURL: "http://xn--bcher-kva.test", want: []string{"a=1"}},
			{toURL: "http://bar.bücher.test", want: []string{"a=1"}},
			{toURL: "http://bar.xn--bcher-kva.test", want: []string{"a=1"}},
			{toURL: "http://foo.www.bücher.test", want: []string{"a=1"}},
			{toURL: "http://foo.www.xn--bcher-kva.test", want: []string{"a=1"}},
			{toURL: "http://other.test", want: nil},
			{toURL: "http://test", want: nil},
		},
	},
	{
		description: "Host cookie on TLD.",
		fromURL:     "http://com",
		setCookies:  []string{"a=1"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://com", want: []string{"a=1"}},
			{toURL: "http://any.com", want: nil},
			{toURL: "http://any.test", want: nil},
		},
	},
	{
		description: "Domain cookie on TLD becomes a host cookie.",
		fromURL:     "http://com",
		setCookies:  []string{"a=1; domain=com"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://com", want: []string{"a=1"}},
			{toURL: "http://any.com", want: nil},
			{toURL: "http://any.test", want: nil},
		},
	},
	{
		description: "Host cookie on public suffix.",
		fromURL:     "http://co.uk",
		setCookies:  []string{"a=1"},
		content:     "a=1",
		queries: []query{
			{toURL: "http://co.uk", want: []string{"a=1"}},
			{toURL: "http://uk", want: nil},
			{toURL: "http://some.co.uk", want: nil},
			{toURL: "http://foo.some.co.uk", want: nil},
			{toURL: "http://any.uk", want: nil},
		},
	},
	{
		description: "Domain cookie on public suffix is ignored.",
		fromURL:     "http://some.co.uk",
		setCookies:  []string{"a=1; domain=co.uk"},
		content:     "",
		queries: []query{
			{toURL: "http://co.uk", want: nil},
			{toURL: "http://uk", want: nil},
			{toURL: "http://some.co.uk", want: nil},
			{toURL: "http://foo.some.co.uk", want: nil},
			{toURL: "http://any.uk", want: nil},
		},
	},
}

func TestDomainHandling(t *testing.T) {
	for idx, test := range domainHandlingTests {
		t.Run(fmt.Sprintf("case %d", idx), func(t *testing.T) {
			jar := newTestJar()
			test.run(t, jar)
		})
	}
}

func TestIssue19384(t *testing.T) {
	cookies := []*http.Cookie{{Name: "name", Value: "value"}}
	for _, host := range []string{"", ".", "..", "..."} {
		jar, _ := New(nil)
		u := &url.URL{Scheme: "http", Host: host, Path: "/"}
		if got := jar.Cookies(u); len(got) != 0 {
			t.Errorf("host %q, got %v", host, got)
		}
		jar.SetCookies(u, cookies)
		if got := jar.Cookies(u); len(got) != 1 || got[0].Value != "value" {
			t.Errorf("host %q, got %v", host, got)
		}
	}
}
