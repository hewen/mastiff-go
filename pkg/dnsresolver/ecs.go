// Package dnsresolver provides DNS resolution with ECS support.
package dnsresolver

import (
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DNSClient is an interface that defines methods for DNS resolution.
type DNSClient interface {
	// Exchange sends a DNS request and returns the response.
	Exchange(m *dns.Msg, addr string) (*dns.Msg, time.Duration, error)
}

// NewECSResolver creates a new ECSResolver.
func NewECSResolver(dnsServer, clientIP string, client DNSClient) *ECSResolver {
	if client == nil {
		client = &dns.Client{
			Timeout: 3 * time.Second,
			Net:     "udp",
		}
	}

	return &ECSResolver{
		DNSServer: dnsServer,
		ClientIP:  clientIP,
		Client:    client,
	}
}

// ECSResolver is a DNS resolver that supports ECS (EDNS Client Subnet) requests.
type ECSResolver struct {
	Client    DNSClient
	DNSServer string
	ClientIP  string
}

// Lookup resolves a domain name to a list of IP addresses using ECS requests.
func (r *ECSResolver) Lookup(domain string) ([]string, error) {
	domain = strings.TrimSpace(domain)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT

	ecsIP := net.ParseIP(r.ClientIP).To4()
	ecs := &dns.EDNS0_SUBNET{
		Code:          dns.EDNS0SUBNET,
		Family:        1, // IPv4
		SourceNetmask: 24,
		SourceScope:   0,
		Address:       ecsIP,
	}
	o.Option = append(o.Option, ecs)
	m.Extra = append(m.Extra, o)

	in, _, err := r.Client.Exchange(m, r.DNSServer)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, a := range in.Answer {
		if arec, ok := a.(*dns.A); ok {
			results = append(results, arec.A.String())
		}
	}
	return results, nil
}
