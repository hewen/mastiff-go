// Package dnsresolver provides DNS resolution with ECS support.
package dnsresolver

import (
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

type mockDNSClient struct{}

func (m *mockDNSClient) Exchange(msg *dns.Msg, _ string) (*dns.Msg, time.Duration, error) {
	resp := new(dns.Msg)
	resp.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{Name: msg.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
			A:   net.ParseIP("1.2.3.4"),
		},
	}
	return resp, 0, nil
}

func TestECSResolver_Lookup(t *testing.T) {
	mockClient := &mockDNSClient{}
	resolver := NewECSResolver("mockdns:53", "1.2.3.0", mockClient)

	ips, err := resolver.Lookup("example.com")
	assert.Nil(t, err)
	assert.Equal(t, len(ips), 1)
	assert.Equal(t, ips[0], "1.2.3.4")
}

func TestNewECSResolver(t *testing.T) {
	resolver := NewECSResolver("mockdns:53", "1.2.3.0", nil)
	assert.IsType(t, &dns.Client{}, resolver.Client)
}
