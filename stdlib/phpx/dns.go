//go:build phpx

package phpx

import (
	"context"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerDNS)
}

// The DNS_* record types, with php's values.
const (
	dnsA     = 1
	dnsNS    = 2
	dnsCNAME = 16
	dnsPTR   = 2048
	dnsMX    = 16384
	dnsTXT   = 32768
	dnsSRV   = 33554432
	dnsAAAA  = 134217728
	dnsANY   = 268435456
	dnsAll   = 251721779
)

// dnsTTL is what a record reports: the resolver in the standard library does
// not expose the time to live.
const dnsTTL = 300

func dnsContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func dnsRecord(host, kind string, extra ...any) *model.Array {
	rec := model.NewArray()
	rec.Set("host", host)
	rec.Set("class", "IN")
	rec.Set("ttl", int64(dnsTTL))
	rec.Set("type", kind)
	for i := 0; i+1 < len(extra); i += 2 {
		rec.Set(extra[i], extra[i+1])
	}
	return rec
}

// dnsLookup answers the records of one type for host, in php's record shape.
func dnsLookup(host string, kind int64) []*model.Array {
	ctx, cancel := dnsContext()
	defer cancel()
	host = strings.TrimSuffix(host, ".")
	r := net.DefaultResolver
	var out []*model.Array
	switch kind {
	case dnsA, dnsAAAA:
		addrs, err := r.LookupIPAddr(ctx, host)
		if err != nil {
			return nil
		}
		for _, a := range addrs {
			is4 := a.IP.To4() != nil
			if kind == dnsA && is4 {
				out = append(out, dnsRecord(host, "A", "ip", a.IP.String()))
			} else if kind == dnsAAAA && !is4 {
				out = append(out, dnsRecord(host, "AAAA", "ipv6", a.IP.String()))
			}
		}
	case dnsMX:
		mx, err := r.LookupMX(ctx, host)
		if err != nil {
			return nil
		}
		for _, m := range mx {
			out = append(out, dnsRecord(host, "MX", "pri", int64(m.Pref), "target", strings.TrimSuffix(m.Host, ".")))
		}
	case dnsNS:
		ns, err := r.LookupNS(ctx, host)
		if err != nil {
			return nil
		}
		for _, n := range ns {
			out = append(out, dnsRecord(host, "NS", "target", strings.TrimSuffix(n.Host, ".")))
		}
	case dnsTXT:
		txt, err := r.LookupTXT(ctx, host)
		if err != nil {
			return nil
		}
		for _, t := range txt {
			entries := model.NewArray()
			entries.Append(t)
			out = append(out, dnsRecord(host, "TXT", "txt", t, "entries", entries))
		}
	case dnsCNAME:
		c, err := r.LookupCNAME(ctx, host)
		if err != nil || strings.TrimSuffix(c, ".") == host {
			return nil
		}
		out = append(out, dnsRecord(host, "CNAME", "target", strings.TrimSuffix(c, ".")))
	case dnsPTR:
		names, err := r.LookupAddr(ctx, host)
		if err != nil {
			return nil
		}
		for _, n := range names {
			out = append(out, dnsRecord(host, "PTR", "target", strings.TrimSuffix(n, ".")))
		}
	case dnsSRV:
		_, srv, err := r.LookupSRV(ctx, "", "", host)
		if err != nil {
			return nil
		}
		for _, s := range srv {
			out = append(out, dnsRecord(host, "SRV", "pri", int64(s.Priority), "weight", int64(s.Weight),
				"port", int64(s.Port), "target", strings.TrimSuffix(s.Target, ".")))
		}
	}
	return out
}

// dnsKinds lists the record types a request names, in php's order.
func dnsKinds(mask int64) []int64 {
	if mask == dnsANY || mask == dnsAll {
		return []int64{dnsA, dnsNS, dnsCNAME, dnsMX, dnsTXT, dnsAAAA}
	}
	var out []int64
	for _, k := range []int64{dnsA, dnsNS, dnsCNAME, dnsPTR, dnsMX, dnsTXT, dnsSRV, dnsAAAA} {
		if mask&k != 0 {
			out = append(out, k)
		}
	}
	return out
}

var dnsNames = map[string]int64{
	"A": dnsA, "NS": dnsNS, "CNAME": dnsCNAME, "PTR": dnsPTR, "MX": dnsMX, "TXT": dnsTXT, "SRV": dnsSRV,
	"AAAA": dnsAAAA, "ANY": dnsANY,
}

func registerDNS(rt *runner.Runtime) {
	getRecord := func(host string, rest ...any) any {
		mask := int64(dnsANY)
		if len(rest) > 0 && rest[0] != nil {
			mask = phpval.Int(rest[0])
		}
		out := model.NewArray()
		for _, kind := range dnsKinds(mask) {
			for _, rec := range dnsLookup(host, kind) {
				out.Append(rec)
			}
		}
		if out.Len() == 0 && len(dnsKinds(mask)) == 0 {
			return false
		}
		return out
	}
	rt.RegisterFunc("dns_get_record", getRecord)
	check := func(host string, rest ...any) bool {
		kind := int64(dnsMX)
		if len(rest) > 0 && rest[0] != nil {
			name := strings.ToUpper(phpval.String(rest[0]))
			k, ok := dnsNames[name]
			if !ok {
				return false
			}
			kind = k
		}
		if kind == dnsANY {
			kind = dnsA
		}
		return len(dnsLookup(host, kind)) > 0
	}
	rt.RegisterFunc("checkdnsrr", check)
	rt.RegisterFunc("dns_check_record", check)
	getmx := func(host string, hosts func(any), rest ...any) bool {
		recs := dnsLookup(host, dnsMX)
		sort.SliceStable(recs, func(i, j int) bool {
			a, _ := recs[i].Get("pri")
			b, _ := recs[j].Get("pri")
			return phpval.Int(a) < phpval.Int(b)
		})
		names, weights := model.NewArray(), model.NewArray()
		for _, r := range recs {
			t, _ := r.Get("target")
			p, _ := r.Get("pri")
			names.Append(t)
			weights.Append(p)
		}
		if hosts != nil {
			hosts(names)
		}
		if len(rest) > 0 {
			if set, ok := rest[0].(func(any)); ok {
				set(weights)
			}
		}
		return len(recs) > 0
	}
	rt.RegisterFunc("getmxrr", getmx)
	rt.RegisterFunc("dns_get_mx", getmx)
	rt.RegisterFunc("gethostbyname", func(host string) string {
		ctx, cancel := dnsContext()
		defer cancel()
		addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		for _, a := range addrs {
			if a.IP.To4() != nil {
				return a.IP.String()
			}
		}
		if err != nil || len(addrs) == 0 {
			return host
		}
		return host
	})
	rt.RegisterFunc("gethostbynamel", func(host string) any {
		ctx, cancel := dnsContext()
		defer cancel()
		addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return false
		}
		out := model.NewArray()
		for _, a := range addrs {
			if a.IP.To4() != nil {
				out.Append(a.IP.String())
			}
		}
		if out.Len() == 0 {
			return false
		}
		return out
	})
	rt.RegisterFunc("gethostbyaddr", func(ip string) any {
		ctx, cancel := dnsContext()
		defer cancel()
		names, err := net.DefaultResolver.LookupAddr(ctx, ip)
		if err != nil || len(names) == 0 {
			return ip
		}
		return strings.TrimSuffix(names[0], ".")
	})
}
