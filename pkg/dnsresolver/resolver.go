// Package dnsresolver provides DNS resolution with ECS support.
package dnsresolver

// Resolver is an interface that defines methods for DNS resolution.
type Resolver interface {
	// Lookup resolves a domain name to a list of IP addresses.
	Lookup(domain string) ([]string, error)
}
