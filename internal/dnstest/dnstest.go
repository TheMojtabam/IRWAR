package dnstest

import (
	"fmt"
	"net"
	"time"
)

type Step struct {
	OK     bool   `json:"ok"`
	Step   string `json:"step"`
	Query  string `json:"query"`
	Detail string `json:"detail"`
	MS     int64  `json:"ms"`
}

type Result struct {
	OK    bool   `json:"ok"`
	Steps []Step `json:"steps"`
}

func Run(resolver, domain string) Result {
	var steps []Step

	// Step 1: Resolver reachability
	step1 := runStep("Resolver Reachability", resolver+":53", func() (string, error) {
		conn, err := net.DialTimeout("udp", resolver+":53", 3*time.Second)
		if err != nil {
			return "", fmt.Errorf("cannot reach %s:53 — %v", resolver, err)
		}
		conn.Close()
		return fmt.Sprintf("reached %s:53", resolver), nil
	})
	steps = append(steps, step1)

	// Step 2: A record query for domain
	step2 := runStep("DNS A Record Lookup", domain, func() (string, error) {
		addrs, err := net.LookupHost(domain)
		if err != nil {
			return "", fmt.Errorf("lookup failed: %v", err)
		}
		if len(addrs) == 0 {
			return "", fmt.Errorf("no addresses returned")
		}
		return fmt.Sprintf("resolved: %s", addrs[0]), nil
	})
	steps = append(steps, step2)

	// Step 3: TXT record query (DNS tunnel check)
	step3 := runStep("TXT Record Query", "test."+domain, func() (string, error) {
		txts, err := net.LookupTXT("test." + domain)
		if err != nil {
			return "no TXT record (may be ok)", nil
		}
		if len(txts) > 0 {
			t := txts[0]
			if len(t) > 60 {
				t = t[:60] + "..."
			}
			return "TXT: " + t, nil
		}
		return "TXT query ok", nil
	})
	steps = append(steps, step3)

	allOK := true
	for _, s := range steps {
		if !s.OK {
			allOK = false
			break
		}
	}

	return Result{OK: allOK, Steps: steps}
}

func runStep(name, query string, fn func() (string, error)) Step {
	start := time.Now()
	detail, err := fn()
	ms := time.Since(start).Milliseconds()
	if err != nil {
		return Step{OK: false, Step: name, Query: query, Detail: err.Error(), MS: ms}
	}
	return Step{OK: true, Step: name, Query: query, Detail: detail, MS: ms}
}
