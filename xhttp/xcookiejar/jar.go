package xcookiejar

import (
	"cmp"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/xanygo/anygo/xhttp/internal/ascii"
	"github.com/xanygo/anygo/xmap"
	"github.com/xanygo/anygo/xsync"
)

var _ http.CookieJar = (*Jar)(nil)

type Storage interface {
	Get(ctx context.Context, key string) ([]Entry, error)
	Set(ctx context.Context, key string, items []Entry) error
	DeleteEntry(ctx context.Context, key string, id ...string) error
}

// Jar implements the [net/http.CookieJar] interface.
type Jar struct {
	ctx     context.Context
	PSList  cookiejar.PublicSuffixList
	Storage Storage

	lastErr xsync.Value[error]
}

func (j *Jar) WithContext(ctx context.Context) *Jar {
	return &Jar{
		PSList:  j.PSList,
		Storage: j.Storage,
		ctx:     ctx,
	}
}

// LastErr 读取最后一次调用 Cookies 或者 SetCookies 方法的错误
func (j *Jar) LastErr() error {
	return j.lastErr.Load()
}

func isLocalhost(host string) bool {
	host = strings.TrimSuffix(host, ".")
	if idx := strings.LastIndex(host, "."); idx >= 0 {
		host = host[idx+1:]
	}
	return strings.EqualFold(host, "localhost")
}

// hasDotSuffix reports whether s ends in "."+suffix.
func hasDotSuffix(s, suffix string) bool {
	return len(s) > len(suffix) && s[len(s)-len(suffix)-1] == '.' && s[len(s)-len(suffix):] == suffix
}

// Cookies implements the Cookies method of the [http.CookieJar] interface.
//
// It returns an empty slice if the URL's scheme is not HTTP or HTTPS.
func (j *Jar) Cookies(u *url.URL) []*http.Cookie {
	j.lastErr.Clear()
	cookies, err := j.cookies(j.getContext(), u, time.Now())
	if err != nil {
		j.lastErr.Store(err)
	}
	return cookies
}

func (j *Jar) CookiesContext(ctx context.Context, u *url.URL) ([]*http.Cookie, error) {
	return j.cookies(ctx, u, time.Now())
}

func (j *Jar) getContext() context.Context {
	if j.ctx != nil {
		return j.ctx
	}
	return context.Background()
}

func (j *Jar) cookies(ctx context.Context, u *url.URL, now time.Time) (cookies []*http.Cookie, err error) {
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, nil
	}
	host, err := canonicalHost(u.Host)
	if err != nil {
		return nil, err
	}
	key := jarKey(host, j.PSList)

	items, err := j.Storage.Get(ctx, key)
	if err != nil || len(items) == 0 {
		return nil, err
	}

	https := u.Scheme == "https"
	path := u.Path
	if path == "" {
		path = "/"
	}

	var needDelete []string
	var selected []Entry
	for _, e := range items {
		if e.Persistent && !e.Expires.After(now) {
			needDelete = append(needDelete, e.ID())
			continue
		}
		if !e.shouldSend(https, host, path) {
			continue
		}
		selected = append(selected, e)
	}

	if len(needDelete) > 0 {
		go j.Storage.DeleteEntry(ctx, key, needDelete...)
	}

	// sort according to RFC 6265 section 5.4 point 2: by longest
	// path and then by earliest creation time.
	slices.SortFunc(selected, entrySort)
	for _, e := range selected {
		cookies = append(cookies, &http.Cookie{Name: e.Name, Value: e.Value, Quoted: e.Quoted})
	}
	return cookies, nil
}

func entrySort(a, b Entry) int {
	if r := cmp.Compare(b.Path, a.Path); r != 0 {
		return r
	}
	if r := a.Creation.Compare(b.Creation); r != 0 {
		return r
	}
	return cmp.Compare(a.SeqNum, b.SeqNum)
}

// SetCookies implements the SetCookies method of the [http.CookieJar] interface.
//
// It does nothing if the URL's scheme is not HTTP or HTTPS.
func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.lastErr.Clear()
	err := j.setCookies(j.getContext(), u, cookies, time.Now())
	if err != nil {
		j.lastErr.Store(err)
	}
}

func (j *Jar) SetCookiesContext(ctx context.Context, u *url.URL, cookies []*http.Cookie) error {
	return j.setCookies(ctx, u, cookies, time.Now())
}

// setCookies is like SetCookies but takes the current time as parameter.
func (j *Jar) setCookies(ctx context.Context, u *url.URL, cookies []*http.Cookie, now time.Time) error {
	if len(cookies) == 0 {
		return nil
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil
	}
	host, err := canonicalHost(u.Host)
	if err != nil {
		return err
	}
	key := jarKey(host, j.PSList)
	defPath := defaultPath(u.Path)

	items, err := j.Storage.Get(ctx, key)
	if err != nil {
		return err
	}
	oldValues := make(map[string]Entry, len(cookies))
	for _, item := range items {
		oldValues[item.ID()] = item
	}

	var needDelete []string

	modified := make(map[string]Entry, len(cookies))
	for _, cookie := range cookies {
		e, remove, err := j.newEntry(cookie, now, defPath, host)
		if err != nil {
			continue
		}
		id := e.ID()
		if remove {
			needDelete = append(needDelete, id)
			continue
		}

		if old, ok := oldValues[id]; ok {
			e.Creation = old.Creation
		} else {
			e.Creation = now
		}
		modified[id] = e
	}

	if len(needDelete) > 0 {
		j.Storage.DeleteEntry(ctx, key, needDelete...)
	}

	if len(modified) > 0 {
		return j.Storage.Set(ctx, key, xmap.Values(modified))
	}
	return nil
}

// canonicalHost strips port from host if present and returns the canonicalized
// host name.
func canonicalHost(host string) (string, error) {
	var err error
	if hasPort(host) {
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			return "", err
		}
	}
	// Strip trailing dot from fully qualified domain names.
	host = strings.TrimSuffix(host, ".")
	encoded, err := toASCII(host)
	if err != nil {
		return "", err
	}

	// We know this is ascii, no need to check.
	lower, _ := ascii.ToLower(encoded)
	return lower, nil
}

// hasPort reports whether host contains a port number. host may be a host
// name, an IPv4 or an IPv6 address.
func hasPort(host string) bool {
	colons := strings.Count(host, ":")
	if colons == 0 {
		return false
	}
	if colons == 1 {
		return true
	}
	return host[0] == '[' && strings.Contains(host, "]:")
}

// jarKey returns the key to use for a jar.
func jarKey(host string, psl cookiejar.PublicSuffixList) string {
	if isIP(host) {
		return host
	}

	var i int
	if psl == nil {
		i = strings.LastIndex(host, ".")
		if i <= 0 {
			return host
		}
	} else {
		suffix := psl.PublicSuffix(host)
		if suffix == host {
			return host
		}
		i = len(host) - len(suffix)
		if i <= 0 || host[i-1] != '.' {
			// The provided public suffix list psl is broken.
			// Storing cookies under host is a safe stopgap.
			return host
		}
		// Only len(suffix) is used to determine the jar key from
		// here on, so it is okay if psl.PublicSuffix("www.buggy.psl")
		// returns "com" as the jar key is generated from host.
	}
	prevDot := strings.LastIndex(host[:i-1], ".")
	return host[prevDot+1:]
}

// isIP reports whether host is an IP address.
func isIP(host string) bool {
	if strings.ContainsAny(host, ":%") {
		// Probable IPv6 address.
		// Hostnames can't contain : or %, so this is definitely not a valid host.
		// Treating it as an IP is the more conservative option, and avoids the risk
		// of interpreting ::1%.www.example.com as a subdomain of www.example.com.
		return true
	}
	return net.ParseIP(host) != nil
}

// defaultPath returns the directory part of a URL's path according to
// RFC 6265 section 5.1.4.
func defaultPath(path string) string {
	if len(path) == 0 || path[0] != '/' {
		return "/" // Path is empty or malformed.
	}

	i := strings.LastIndex(path, "/") // Path starts with "/", so i != -1.
	if i == 0 {
		return "/" // Path has the form "/abc".
	}
	return path[:i] // Path is either of form "/abc/xyz" or "/abc/xyz/".
}

// newEntry creates an entry from an http.Cookie c. now is the current time and
// is compared to c.Expires to determine deletion of c. defPath and host are the
// default-path and the canonical host name of the URL c was received from.
//
// remove records whether the jar should delete this cookie, as it has already
// expired with respect to now. In this case, e may be incomplete, but it will
// be valid to call e.id (which depends on e's Name, Domain and Path).
//
// A malformed c.Domain will result in an error.
func (j *Jar) newEntry(c *http.Cookie, now time.Time, defPath, host string) (e Entry, remove bool, err error) {
	e.Name = c.Name

	if c.Path == "" || c.Path[0] != '/' {
		e.Path = defPath
	} else {
		e.Path = c.Path
	}

	e.Domain, e.HostOnly, err = j.domainAndType(host, c.Domain)
	if err != nil {
		return e, false, err
	}

	// MaxAge takes precedence over Expires.
	if c.MaxAge < 0 {
		return e, true, nil
	} else if c.MaxAge > 0 {
		e.Expires = now.Add(time.Duration(c.MaxAge) * time.Second)
		e.Persistent = true
	} else {
		if c.Expires.IsZero() {
			e.Expires = endOfTime
			e.Persistent = false
		} else {
			if !c.Expires.After(now) {
				return e, true, nil
			}
			e.Expires = c.Expires
			e.Persistent = true
		}
	}

	e.Value = c.Value
	e.Quoted = c.Quoted
	e.Secure = c.Secure
	e.HttpOnly = c.HttpOnly
	e.SameSite = c.SameSite

	return e, false, nil
}

var (
	errIllegalDomain   = errors.New("cookiejar: illegal cookie domain attribute")
	errMalformedDomain = errors.New("cookiejar: malformed cookie domain attribute")
)

// endOfTime is the time when session (non-persistent) cookies expire.
// This instant is representable in most date/time formats (not just
// Go's time.Time) and should be far enough in the future.
var endOfTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

// domainAndType determines the cookie's domain and hostOnly attribute.
func (j *Jar) domainAndType(host, domain string) (string, bool, error) {
	if domain == "" {
		// No domain attribute in the SetCookie header indicates a
		// host cookie.
		return host, true, nil
	}

	if isIP(host) {
		// RFC 6265 is not super clear here, a sensible interpretation
		// is that cookies with an IP address in the domain-attribute
		// are allowed.

		// RFC 6265 section 5.2.3 mandates to strip an optional leading
		// dot in the domain-attribute before processing the cookie.
		//
		// Most browsers don't do that for IP addresses, only curl
		// (version 7.54) and IE (version 11) do not reject a
		//     Set-Cookie: a=1; domain=.127.0.0.1
		// This leading dot is optional and serves only as hint for
		// humans to indicate that a cookie with "domain=.bbc.co.uk"
		// would be sent to every subdomain of bbc.co.uk.
		// It just doesn't make sense on IP addresses.
		// The other processing and validation steps in RFC 6265 just
		// collapse to:
		if host != domain {
			return "", false, errIllegalDomain
		}

		// According to RFC 6265 such cookies should be treated as
		// domain cookies.
		// As there are no subdomains of an IP address the treatment
		// according to RFC 6265 would be exactly the same as that of
		// a host-only cookie. Contemporary browsers (and curl) do
		// allows such cookies but treat them as host-only cookies.
		// So do we as it just doesn't make sense to label them as
		// domain cookies when there is no domain; the whole notion of
		// domain cookies requires a domain name to be well defined.
		return host, true, nil
	}

	// From here on: If the cookie is valid, it is a domain cookie (with
	// the one exception of a public suffix below).
	// See RFC 6265 section 5.2.3.
	domain = strings.TrimPrefix(domain, ".")

	if len(domain) == 0 || domain[0] == '.' {
		// Received either "Domain=." or "Domain=..some.thing",
		// both are illegal.
		return "", false, errMalformedDomain
	}

	domain, isASCII := ascii.ToLower(domain)
	if !isASCII {
		// Received non-ASCII domain, e.g. "perché.com" instead of "xn--perch-fsa.com"
		return "", false, errMalformedDomain
	}

	if domain[len(domain)-1] == '.' {
		// We received stuff like "Domain=www.example.com.".
		// Browsers do handle such stuff (actually differently) but
		// RFC 6265 seems to be clear here (e.g. section 4.1.2.3) in
		// requiring a reject.  4.1.2.3 is not normative, but
		// "Domain Matching" (5.1.3) and "Canonicalized Host Names"
		// (5.1.2) are.
		return "", false, errMalformedDomain
	}

	// See RFC 6265 section 5.3 #5.
	if j.PSList != nil {
		if ps := j.PSList.PublicSuffix(domain); ps != "" && !hasDotSuffix(domain, ps) {
			if host == domain {
				// This is the one exception in which a cookie
				// with a domain attribute is a host cookie.
				return host, true, nil
			}
			return "", false, errIllegalDomain
		}
	}

	// The domain must domain-match host: www.mycompany.com cannot
	// set cookies for .ourcompetitors.com.
	if host != domain && !hasDotSuffix(host, domain) {
		return "", false, errIllegalDomain
	}

	return domain, false, nil
}
