package dnstest

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

type StepResult struct {
	Step   string `json:"step"`
	Query  string `json:"query"`
	Status string `json:"status"`
	MS     int64  `json:"ms"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

type Result struct {
	OK    bool         `json:"ok"`
	Steps []StepResult `json:"steps"`
}

func Run(resolver, domain string) Result {
	host := resolver
	port := "53"
	if strings.Contains(resolver, ":") {
		parts := strings.SplitN(resolver, ":", 2)
		host, port = parts[0], parts[1]
	}
	addr := net.JoinHostPort(host, port)

	steps := []struct {
		name   string
		qname  string
		qtype  uint16
		qtypes string
	}{
		{"A record", domain, 1, "A"},
		{"NS record", domain, 2, "NS"},
		{"TXT probe", "test." + domain, 16, "TXT"},
	}

	var results []StepResult
	allOK := true

	for _, s := range steps {
		r := queryDNS(addr, s.qname, s.qtype, s.name)
		results = append(results, r)
		if !r.OK {
			allOK = false
		}
	}

	return Result{OK: allOK, Steps: results}
}

func queryDNS(addr, domain string, qtype uint16, stepName string) StepResult {
	start := time.Now()

	conn, err := net.DialTimeout("udp", addr, 5*time.Second)
	if err != nil {
		return StepResult{
			Step: stepName, Query: domain,
			Status: "error", MS: 0, OK: false,
			Detail: fmt.Sprintf("dial failed: %v", err),
		}
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	pkt := buildQuery(domain, qtype)
	if _, err := conn.Write(pkt); err != nil {
		return StepResult{
			Step: stepName, Query: domain,
			Status: "error", OK: false,
			Detail: fmt.Sprintf("send failed: %v", err),
		}
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	ms := time.Since(start).Milliseconds()

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return StepResult{
				Step: stepName, Query: domain,
				Status: "timeout", MS: 5000, OK: false,
				Detail: "no response after 5s",
			}
		}
		return StepResult{
			Step: stepName, Query: domain,
			Status: "error", MS: ms, OK: false,
			Detail: fmt.Sprintf("recv failed: %v", err),
		}
	}

	if n < 12 {
		return StepResult{
			Step: stepName, Query: domain,
			Status: "error", MS: ms, OK: false,
			Detail: "response too short",
		}
	}

	rcode   := int(buf[3] & 0x0F)
	ancount := int(binary.BigEndian.Uint16(buf[6:8]))

	status := "ok"
	detail := fmt.Sprintf("responded in %dms, answers=%d", ms, ancount)
	ok := true

	switch rcode {
	case 0:
		status = "ok"
	case 3:
		status = "nxdomain"
		detail = fmt.Sprintf("NXDOMAIN in %dms (server reachable)", ms)
	default:
		status = fmt.Sprintf("rcode=%d", rcode)
		ok = false
		detail = fmt.Sprintf("DNS error rcode=%d in %dms", rcode, ms)
	}

	return StepResult{
		Step:   stepName,
		Query:  domain,
		Status: status,
		MS:     ms,
		OK:     ok,
		Detail: detail,
	}
}

func buildQuery(domain string, qtype uint16) []byte {
	buf := make([]byte, 0, 64)

	// Header
	buf = append(buf, 0x13, 0x37) // ID
	buf = append(buf, 0x01, 0x00) // flags: RD=1
	buf = append(buf, 0x00, 0x01) // QDCOUNT=1
	buf = append(buf, 0x00, 0x00) // ANCOUNT=0
	buf = append(buf, 0x00, 0x00) // NSCOUNT=0
	buf = append(buf, 0x00, 0x00) // ARCOUNT=0

	// QNAME
	for _, label := range strings.Split(strings.TrimSuffix(domain, "."), ".") {
		buf = append(buf, byte(len(label)))
		buf = append(buf, []byte(label)...)
	}
	buf = append(buf, 0x00) // root

	// QTYPE + QCLASS
	buf = append(buf, byte(qtype>>8), byte(qtype))
	buf = append(buf, 0x00, 0x01) // IN

	return buf
}
