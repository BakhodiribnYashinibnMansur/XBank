// Package httpclient provides an instrumented HTTP client for external API calls.
// ssrf_guard.go — validates outbound URLs to prevent Server-Side Request Forgery (OWASP A10).
package httpclient

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// SSRFGuard validates URLs before making outbound HTTP requests.
// It blocks requests to private/internal networks, restricted protocols,
// and unapproved ports to prevent SSRF attacks.
type SSRFGuard struct {
	allowedHosts map[string]struct{}
	allowedPorts map[string]struct{}
	blockedCIDRs []*net.IPNet
}

// NewSSRFGuard creates a new SSRF guard with the given allowed hosts.
// Default blocked CIDRs: loopback, private networks, link-local, metadata endpoints.
// Default allowed ports: 80, 443.
func NewSSRFGuard(allowedHosts []string, allowedPorts []string) *SSRFGuard {
	hosts := make(map[string]struct{}, len(allowedHosts))
	for _, h := range allowedHosts {
		hosts[strings.ToLower(h)] = struct{}{}
	}

	ports := make(map[string]struct{}, len(allowedPorts))
	if len(allowedPorts) == 0 {
		ports["80"] = struct{}{}
		ports["443"] = struct{}{}
		ports[""] = struct{}{} // default port (no port specified)
	} else {
		for _, p := range allowedPorts {
			ports[p] = struct{}{}
		}
		ports[""] = struct{}{}
	}

	cidrs := []string{
		"127.0.0.0/8",    // loopback
		"10.0.0.0/8",     // private class A
		"172.16.0.0/12",  // private class B
		"192.168.0.0/16", // private class C
		"169.254.0.0/16", // link-local (AWS metadata: 169.254.169.254)
		"0.0.0.0/8",      // current network
		"100.64.0.0/10",  // shared address space (CGN)
		"198.18.0.0/15",  // benchmark testing
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 unique local
		"fe80::/10",      // IPv6 link-local
	}

	blocked := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			blocked = append(blocked, network)
		}
	}

	return &SSRFGuard{
		allowedHosts: hosts,
		allowedPorts: ports,
		blockedCIDRs: blocked,
	}
}

// Validate checks whether the given raw URL is safe for outbound requests.
// Returns an error if the URL targets a private network, uses a blocked protocol,
// or connects to a disallowed port.
func (g *SSRFGuard) Validate(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("ssrf: invalid URL: %w", err)
	}

	// 1. Protocol check — only http and https allowed
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("ssrf: blocked protocol %q (only http/https allowed)", parsed.Scheme)
	}

	// 2. Host must not be empty
	hostname := parsed.Hostname()
	if hostname == "" {
		return fmt.Errorf("ssrf: empty hostname")
	}

	// 3. Port check
	port := parsed.Port()
	if _, ok := g.allowedPorts[port]; !ok {
		return fmt.Errorf("ssrf: blocked port %q", port)
	}

	// 4. If allowedHosts is configured, enforce allowlist
	if len(g.allowedHosts) > 0 {
		if _, ok := g.allowedHosts[strings.ToLower(hostname)]; !ok {
			return fmt.Errorf("ssrf: host %q not in allowlist", hostname)
		}
	}

	// 5. Resolve hostname and check against blocked CIDRs
	ips, err := net.LookupHost(hostname)
	if err != nil {
		return fmt.Errorf("ssrf: DNS resolution failed for %q: %w", hostname, err)
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		for _, cidr := range g.blockedCIDRs {
			if cidr.Contains(ip) {
				return fmt.Errorf("ssrf: blocked IP %s (resolves to private network %s)", ipStr, cidr.String())
			}
		}
	}

	return nil
}
