package xcookiejar

import (
	"fmt"
	"net/http"
	"net/netip"
	"strings"
	"time"
)

// Entry is the internal representation of a cookie.
//
// This struct type is not used outside of this package per se, but the exported
// fields are those of RFC 6265.
type Entry struct {
	Name       string        `json:"Name,omitzero"`
	Value      string        `json:"Value,omitzero"`
	Quoted     bool          `json:"Quoted,omitzero"`
	Domain     string        `json:"Domain,omitzero"`
	Path       string        `json:"Path,omitzero"`
	SameSite   http.SameSite `json:"SameSite,omitzero"`
	Secure     bool          `json:"Secure,omitzero"`
	HttpOnly   bool          `json:"HttpOnly,omitzero"`
	Persistent bool          `json:"Persistent,omitzero"`
	HostOnly   bool          `json:"HostOnly,omitzero"`
	Expires    time.Time     `json:"Expires,omitzero"`
	Creation   time.Time     `json:"Creation,omitzero"`
}

// id returns the domain;path;name triple of e as an id.
func (e *Entry) ID() string {
	return fmt.Sprintf("%s;%s;%s", e.Domain, e.Path, e.Name)
}

// shouldSend determines whether e's cookie qualifies to be included in a
// request to host/path. It is the caller's responsibility to check if the
// cookie is expired.
func (e *Entry) shouldSend(https bool, host, path string) bool {
	return e.domainMatch(host) && e.pathMatch(path) && e.secureMatch(https)
}

// domainMatch checks whether e's Domain allows sending e back to host.
// It differs from "domain-match" of RFC 6265 section 5.1.3 because we treat
// a cookie with an IP address in the Domain always as a host cookie.
func (e *Entry) domainMatch(host string) bool {
	if e.Domain == host {
		return true
	}
	return !e.HostOnly && hasDotSuffix(host, e.Domain)
}

// pathMatch implements "path-match" according to RFC 6265 section 5.1.4.
func (e *Entry) pathMatch(requestPath string) bool {
	if requestPath == e.Path {
		return true
	}
	if strings.HasPrefix(requestPath, e.Path) {
		if e.Path[len(e.Path)-1] == '/' {
			return true // The "/any/" matches "/any/path" case.
		} else if requestPath[len(e.Path)] == '/' {
			return true // The "/any" matches "/any/path" case.
		}
	}
	return false
}

// secureMatch checks whether a cookie should be sent based on the protocol
// and the Secure flag. Localhost is considered a secure origin regardless
// of protocol, matching browser behavior.
func (e *Entry) secureMatch(https bool) bool {
	if !e.Secure {
		// Cookies not marked secure are always sent.
		return true
	}
	// Everything below is about cookies marked secure.
	if https {
		// HTTPS request matches secure cookies.
		return true
	}
	// Consider localhost to be secure like browsers.
	if isLocalhost(e.Domain) {
		return true
	}
	ip, err := netip.ParseAddr(e.Domain)
	if err == nil && ip.IsLoopback() {
		return true
	}
	return false
}
