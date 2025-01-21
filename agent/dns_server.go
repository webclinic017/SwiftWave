package main

import (
	"log"
	"strings"

	"github.com/miekg/dns"
)

func removeDotSuffix(domain string) string {
	if strings.HasSuffix(domain, ".") {
		return domain[:len(domain)-1]
	}
	return domain
}

func resolveDomain(domain string) []dns.RR {
	answers := []dns.RR{}
	client := dns.Client{}
	msg := dns.Msg{}
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	ips := FetchARecordIps(removeDotSuffix(domain))

	if len(ips) > 0 {
		for _, ip := range ips {
			rr, err := dns.NewRR(domain + " A " + ip)
			if err != nil {
				log.Printf("failed to create RR: %v", err)
				continue
			}
			rr.Header().Ttl = 10
			answers = append(answers, rr)
		}
		return answers
	}

	r, _, err := client.Exchange(&msg, "1.1.1.1:53")
	if err != nil || len(r.Answer) == 0 {
		return answers
	}
	return r.Answer

}

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		if q.Qtype == dns.TypeA {
			m.Answer = resolveDomain(q.Name)

		}
	}
}

func startDnsServer() {
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Compress = false
		switch r.Opcode {
		case dns.OpcodeQuery:
			parseQuery(m)
		}
		_ = w.WriteMsg(m)
	})
	server := &dns.Server{Addr: "10.0.1.1:53", Net: "udp"}
	log.Printf("Starting DNS server at %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start DNS server: %s\n ", err)
	}
}
