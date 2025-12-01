package resolver

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Mreza2020/DNS-Switcher/internal/config"
	"github.com/miekg/dns"
)

type Result struct {
	Server string
	RTT    time.Duration
	Error  error
}

var (
	DomainTesting string
)

func init() {
	DomainTesting = os.Getenv("DomainTesting")

	if DomainTesting == "" {
		log.Fatal("Environment variable DomainTesting is not set. Please set it before running.")
		return
	}
}

// MeasureRTT performs a DNS query against a given DNS server and measures round-trip time (RTT).
// It sends an A-record query for the provided qname, using the specified timeout.
// Returns a Result struct containing:
//   - Server: DNS server tested
//   - RTT: measured round-trip duration
//   - Error: non-nil if exchange failed, timeout occurred, or no valid answer was returned
func MeasureRTT(server string, qname string, timeout time.Duration) Result {
	c := new(dns.Client)
	c.Timeout = timeout

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(qname), dns.TypeA)

	start := time.Now()
	r, _, err := c.Exchange(m, server+":53")
	rtt := time.Since(start)

	if err != nil {
		return Result{Server: server, RTT: rtt, Error: err}
	}
	if r.Rcode != dns.RcodeSuccess || len(r.Answer) == 0 {
		return Result{Server: server, RTT: rtt, Error: fmt.Errorf("no answer or rcode %d", r.Rcode)}
	}
	return Result{Server: server, RTT: rtt, Error: nil}
}

// FindFastestProfile evaluates multiple DNS profiles and determines which profile
// has the lowest average DNS RTT across its configured DNS servers.
// For each profile:
//   - MeasureRTT is executed for every server
//   - Only successful RTT samples are averaged
//   - Profiles with zero successful responses are ignored
//
// Returns a pointer to the fastest profile, or nil if none have valid responding servers.
func FindFastestProfile(profiles []config.Profile) *config.Profile {
	var fastest *config.Profile
	var bestRTT time.Duration

	for i := range profiles {
		p := profiles[i]
		var sum time.Duration
		var count int
		for _, s := range p.Servers {
			r := MeasureRTT(s, DomainTesting, 2*time.Second)
			if r.Error == nil {
				sum += r.RTT
				count++
			}
		}

		if count == 0 {
			continue
		}

		avg := sum / time.Duration(count)

		if fastest == nil || avg < bestRTT {
			bestRTT = avg
			fastest = &p
		}
	}

	return fastest
}
