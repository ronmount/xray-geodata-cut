package geoip

import (
	"fmt"
	stdnet "net"
	"os"
	"strings"

	"github.com/xtls/xray-core/app/router"
	"github.com/xtls/xray-core/common/net"
	"google.golang.org/protobuf/proto"
)

func LoadGeoIP(fn string) (*router.GeoIPList, error) {
	geoIPBytes, err1 := os.ReadFile(fn)
	if err1 != nil {
		return nil, err1
	}
	var geoIPList router.GeoIPList
	if err2 := proto.Unmarshal(geoIPBytes, &geoIPList); err2 != nil {
		return nil, err2
	}
	return &geoIPList, nil
}

func GetGeoIPCodes(in *router.GeoIPList) []string {
	result := make([]string, len(in.GetEntry()))
	for index, x := range in.GetEntry() {
		result[index] = x.CountryCode
	}
	return result
}

func CutGeoIPCodes(in *router.GeoIPList, codesToKeep []string, trimIPv6 bool) *router.GeoIPList {
	out := &router.GeoIPList{
		Entry: make([]*router.GeoIP, 0, len(codesToKeep)+1),
	}
	kept := make(map[string]bool, len(codesToKeep)+1)
	for _, x := range in.GetEntry() {
		for _, y := range codesToKeep {
			u := strings.ToUpper(x.CountryCode)
			switch u {
			case strings.ToUpper(strings.TrimSpace(y)), "PRIVATE":
				{
					if kept[u] {
						continue
					}
					if trimIPv6 {
						newEntry := &router.GeoIP{
							ReverseMatch: x.GetReverseMatch(),
							CountryCode:  x.GetCountryCode(),
							Cidr:         make([]*router.CIDR, 0, len(x.GetCidr())),
						}
						for _, c := range x.Cidr {
							if len(c.Ip) == net.IPv4len {
								newEntry.Cidr = append(newEntry.Cidr, c)
							}
						}
						out.Entry = append(out.Entry, newEntry)
					} else {
						out.Entry = append(out.Entry, x)
					}
					kept[u] = true
				}
			}
		}
	}

	return out
}

func SaveGeoIP(in *router.GeoIPList, fn string) error {
	b, err := proto.Marshal(in)
	if err != nil {
		return err
	}
	return os.WriteFile(fn, b, 0644)
}

func AddGeoIPEntry(in *router.GeoIPList, code string, cidrs []string, reverseMatch bool) (*router.GeoIPList, error) {
	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	if normalizedCode == "" {
		return nil, fmt.Errorf("geoip code is required")
	}

	parsedCIDRs := make([]*router.CIDR, 0, len(cidrs))
	seenCIDRs := make(map[string]struct{}, len(cidrs))
	for _, raw := range cidrs {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		_, ipNet, err := stdnet.ParseCIDR(value)
		if err != nil {
			return nil, fmt.Errorf("parse CIDR %q: %w", value, err)
		}
		maskSize, _ := ipNet.Mask.Size()
		ip := ipNet.IP
		if ipv4 := ip.To4(); ipv4 != nil {
			ip = ipv4
		} else {
			ip = ip.To16()
		}
		key := ipNet.String()
		if _, ok := seenCIDRs[key]; ok {
			continue
		}
		seenCIDRs[key] = struct{}{}
		parsedCIDRs = append(parsedCIDRs, &router.CIDR{
			Ip:     append([]byte(nil), ip...),
			Prefix: uint32(maskSize),
		})
	}
	if len(parsedCIDRs) == 0 {
		return nil, fmt.Errorf("at least one CIDR is required")
	}

	out := &router.GeoIPList{
		Entry: make([]*router.GeoIP, len(in.GetEntry())),
	}
	copy(out.Entry, in.GetEntry())

	for _, entry := range out.Entry {
		if strings.EqualFold(entry.GetCountryCode(), normalizedCode) {
			existing := make(map[string]struct{}, len(entry.GetCidr()))
			for _, cidr := range entry.GetCidr() {
				existing[cidrKey(cidr)] = struct{}{}
			}
			for _, cidr := range parsedCIDRs {
				key := cidrKey(cidr)
				if _, ok := existing[key]; ok {
					continue
				}
				entry.Cidr = append(entry.Cidr, cidr)
				existing[key] = struct{}{}
			}
			return out, nil
		}
	}

	out.Entry = append(out.Entry, &router.GeoIP{
		CountryCode:  normalizedCode,
		Cidr:         parsedCIDRs,
		ReverseMatch: reverseMatch,
	})
	return out, nil
}

func cidrKey(cidr *router.CIDR) string {
	if cidr == nil {
		return ""
	}
	return fmt.Sprintf("%s/%d", stdnet.IP(cidr.GetIp()).String(), cidr.GetPrefix())
}
