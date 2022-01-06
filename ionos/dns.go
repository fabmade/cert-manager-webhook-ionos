package ionos

import (
	"fmt"
	"github.com/miekg/dns"
)

func (e *ionosSolver) handleDNSRequest(w dns.ResponseWriter, req *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(req)
	switch req.Opcode {
	case dns.OpcodeQuery:
		for _, q := range msg.Question {
			if err := e.addDNSAnswer(q, msg, req); err != nil {
				msg.SetRcode(req, dns.RcodeServerFailure)
				break
			}
		}
	}
	w.WriteMsg(msg)
}

func (e *ionosSolver) addDNSAnswer(q dns.Question, msg *dns.Msg, req *dns.Msg) error {
	switch q.Qtype {
	// Always return loopback for any A query
	case dns.TypeA:
		rr, err := dns.NewRR(fmt.Sprintf("%s 5 IN A 127.0.0.1", q.Name))
		if err != nil {
			return err
		}
		msg.Answer = append(msg.Answer, rr)
		return nil

	// TXT records are the only important record for ACME dns-01 challenges
	case dns.TypeTXT:
		e.RLock()
		record, err := e.ionosClient.GetRecordById(e.context, "", q.Name)
		e.RUnlock()
		if err != nil {
			msg.SetRcode(req, dns.RcodeNameError)
			return nil
		}
		rr, err := dns.NewRR(fmt.Sprintf("%s 5 IN TXT %s", q.Name, record))
		if err != nil {
			return err
		}
		msg.Answer = append(msg.Answer, rr)
		return nil

	// NS and SOA are for authoritative lookups, return obviously invalid data
	case dns.TypeNS:
		rr, err := dns.NewRR(fmt.Sprintf("%s 5 IN NS ns.ionos-acme-webook.invalid.", q.Name))
		if err != nil {
			return err
		}
		msg.Answer = append(msg.Answer, rr)
		return nil
	case dns.TypeSOA:
		rr, err := dns.NewRR(fmt.Sprintf("%s 5 IN SOA %s 20 5 5 5 5", "ns.ionos-acme-webook.invalid.", "ns.ionos-acme-webook.invalid."))
		if err != nil {
			return err
		}
		msg.Answer = append(msg.Answer, rr)
		return nil
	default:
		return fmt.Errorf("unimplemented record type %v", q.Qtype)
	}
}
