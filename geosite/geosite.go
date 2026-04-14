package geosite

import (
	"fmt"
	"os"
	"strings"

	"github.com/xtls/xray-core/app/router"
	"google.golang.org/protobuf/proto"
)

func LoadGeoSite(fn string) (*router.GeoSiteList, error) {
	geoSiteBytes, err1 := os.ReadFile(fn)
	if err1 != nil {
		return nil, err1
	}
	var geoSiteList router.GeoSiteList
	if err2 := proto.Unmarshal(geoSiteBytes, &geoSiteList); err2 != nil {
		return nil, err2
	}
	return &geoSiteList, nil
}

func GetGeoSiteCodes(in *router.GeoSiteList) []string {
	result := make([]string, len(in.GetEntry()))
	for index, x := range in.GetEntry() {
		result[index] = x.CountryCode
	}
	return result
}

func CutGeoSiteCodes(in *router.GeoSiteList, codesToKeep []string) *router.GeoSiteList {
	out := &router.GeoSiteList{
		Entry: make([]*router.GeoSite, 0, len(codesToKeep)),
	}
	kept := make(map[string]bool, len(codesToKeep))
	for _, x := range in.GetEntry() {
		for _, y := range codesToKeep {
			u := strings.ToUpper(y)
			if x.CountryCode == u {
				if kept[u] {
					continue
				}
				out.Entry = append(out.Entry, x)
				kept[u] = true
			}
		}
	}

	return out
}

func SaveGeoSite(in *router.GeoSiteList, fn string) error {
	b, err := proto.Marshal(in)
	if err != nil {
		return err
	}
	return os.WriteFile(fn, b, 0644)
}

func AddGeoSiteEntry(in *router.GeoSiteList, code string, domains []string, domainType router.Domain_Type) (*router.GeoSiteList, error) {
	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	if normalizedCode == "" {
		return nil, fmt.Errorf("geosite code is required")
	}

	parsedDomains := make([]*router.Domain, 0, len(domains))
	seenDomains := make(map[string]struct{}, len(domains))
	for _, raw := range domains {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		key := domainKey(domainType, value)
		if _, ok := seenDomains[key]; ok {
			continue
		}
		seenDomains[key] = struct{}{}
		parsedDomains = append(parsedDomains, &router.Domain{
			Type:  domainType,
			Value: value,
		})
	}
	if len(parsedDomains) == 0 {
		return nil, fmt.Errorf("at least one domain rule is required")
	}

	out := &router.GeoSiteList{
		Entry: make([]*router.GeoSite, len(in.GetEntry())),
	}
	copy(out.Entry, in.GetEntry())

	for _, entry := range out.Entry {
		if strings.EqualFold(entry.GetCountryCode(), normalizedCode) {
			existing := make(map[string]struct{}, len(entry.GetDomain()))
			for _, domain := range entry.GetDomain() {
				existing[domainKey(domain.GetType(), domain.GetValue())] = struct{}{}
			}
			for _, domain := range parsedDomains {
				key := domainKey(domain.GetType(), domain.GetValue())
				if _, ok := existing[key]; ok {
					continue
				}
				entry.Domain = append(entry.Domain, domain)
				existing[key] = struct{}{}
			}
			return out, nil
		}
	}

	out.Entry = append(out.Entry, &router.GeoSite{
		CountryCode: normalizedCode,
		Domain:      parsedDomains,
	})
	return out, nil
}

func domainKey(domainType router.Domain_Type, value string) string {
	return fmt.Sprintf("%d:%s", domainType, strings.TrimSpace(value))
}
