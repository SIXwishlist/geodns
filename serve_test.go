package main

import (
	"github.com/miekg/dns"
	. "launchpad.net/gocheck"
	"strings"
	"time"
)

func (s *ConfigSuite) TestServing(c *C) {

	Zones := make(Zones)
	setupPgeodnsZone(Zones)
	go configReader("dns", Zones)
	go listenAndServe(":8853", &Zones)

	time.Sleep(100 * time.Millisecond)

	r := exchange(c, "_status.pgeodns.", dns.TypeTXT)
	txt := r.Answer[0].(*dns.RR_TXT).Txt[0]
	if !strings.HasPrefix(txt, "{") {
		c.Log("Unexpected result for _status.pgeodns", txt)
		c.Fail()
	}

	r = exchange(c, "bar.example.com.", dns.TypeA)
	ip := r.Answer[0].(*dns.RR_A).A
	if ip.String() != "192.168.1.2" {
		c.Log("Unexpected result for bar.example.com", ip, "!= 192.168.1.2")
		c.Fail()
	}

	r = exchange(c, "example.com.", dns.TypeSOA)
	soa := r.Answer[0].(*dns.RR_SOA)
	serial := soa.Serial
	if serial != 3 {
		c.Log("Didn't get SOA record with serial=3 for bar.example.com/AAAA")
		c.Fail()
	}
}

func exchange(c *C, name string, dnstype uint16) *dns.Msg {
	msg := new(dns.Msg)
	cli := new(dns.Client)

	msg.SetQuestion(name, dnstype)
	r, err := cli.Exchange(msg, "127.0.0.1:8853")
	if err != nil {
		c.Log("err", err)
		c.Fail()
	}
	return r
}