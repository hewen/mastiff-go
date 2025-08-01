// Package dnsresolver provides DNS resolution with ECS support.
package dnsresolver

import (
	"net"
)

// HostLookupFunc is a function that resolves a domain name to a list of IP addresses.
type HostLookupFunc func(domain string) ([]string, error)

// StdResolver is a standard DNS resolver that uses the net.LookupHost function.
type StdResolver struct {
	lookupHost HostLookupFunc
}

// NewStdResolver creates a new StdResolver.
func NewStdResolver(lookup HostLookupFunc) *StdResolver {
	if lookup == nil {
		lookup = net.LookupHost
	}
	return &StdResolver{
		lookupHost: lookup,
	}
}

// Lookup resolves a domain name to a list of IP addresses.
func (r *StdResolver) Lookup(domain string) ([]string, error) {
	return r.lookupHost(domain)
}
