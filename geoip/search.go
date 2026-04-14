package geoip

import (
	"github.com/xtls/xray-core/app/router"
	"github.com/xtls/xray-core/common/net"
)

func Search(gin *router.GeoIPList, addr string) []string {
	var result []string
	addrParsed := net.ParseAddress(addr)

	for _, x := range gin.Entry {
		m, err := router.BuildOptimizedGeoIPMatcher(x)
		if err != nil {
			return result
		}
		if m.Match(addrParsed.IP()) {
			result = append(result, x.CountryCode)
		}
	}

	return result
}
