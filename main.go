package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/xtls/xray-core/app/router"
	"google.golang.org/protobuf/proto"

	"github.com/yichya/xray-geodata-cut/asn"
	"github.com/yichya/xray-geodata-cut/geoip"
	"github.com/yichya/xray-geodata-cut/geosite"
)

func main() {
	ft := flag.String("type", "", "ASN (asn), GeoIP (geoip) or GeoSite (geosite)")
	in := flag.String("in", "", "Path to GeoData file / ASNs split by comma")
	show := flag.Bool("show", false, "Print codes in GeoIP or GeoSite file")
	search := flag.String("search", "", "Search GeoIP or GeoSite Item")
	add := flag.Bool("add", false, "Add entries to GeoIP or GeoSite file")
	code := flag.String("code", "", "GeoIP or GeoSite code to add or update")
	value := flag.String("value", "", "CIDRs for GeoIP or domain rules for GeoSite, split by comma")
	domainType := flag.String("domain-type", "plain", "GeoSite domain rule type: plain, regex, domain/rootdomain, full")
	reverseMatch := flag.Bool("reverse-match", false, "Set reverse_match for newly created GeoIP entries")
	keep := flag.String("keep", "cn,private,geolocation-!cn", "GeoIP or GeoSite codes to keep (private is always kept for GeoIP)")
	trimipv6 := flag.Bool("trimipv6", false, "Trim all IPv6 ranges in GeoIP file")
	out := flag.String("out", "", "Path to processed file")

	flag.Parse()
	if ft == nil {
		ft = proto.String("")
	}
	switch *ft {
	case "asn":
		{
			var asnList []int32
			if in != nil {
				for _, x := range strings.Split(*in, ",") {
					if v, err := strconv.ParseInt(x, 10, 64); err != nil {
						panic(err)
					} else {
						asnList = append(asnList, int32(v))
					}
				}
				if *show {
					for _, x := range asnList {
						if resp, err := asn.GetAsnData(x); err != nil {
							panic(err)
						} else {
							for _, y := range resp.Subnets.Ipv4 {
								fmt.Printf("AS%d %s\n", x, y)
							}
							if !*trimipv6 {
								for _, y := range resp.Subnets.Ipv6 {
									fmt.Printf("AS%d %s\n", x, y)
								}
							}
						}
					}
				} else if *search != "" {
					if gin, err := asn.BuildGeoIp(asnList, *trimipv6); err != nil {
						panic(err)
					} else {
						for _, x := range geoip.Search(gin, *search) {
							fmt.Println(x)
						}
					}
				} else {
					if data, err := asn.BuildGeoIp(asnList, *trimipv6); err != nil {
						panic(err)
					} else {
						if err = geoip.SaveGeoIP(data, *out); err != nil {
							panic(err)
						}
					}
				}
			} else {
				flag.Usage()
			}
		}
	case "geoip":
		{
			gin, err := geoip.LoadGeoIP(*in)
			if err != nil {
				panic(err)
			}
			if *show {
				fmt.Println(geoip.GetGeoIPCodes(gin))
			} else if *search != "" {
				for _, x := range geoip.Search(gin, *search) {
					fmt.Println(x)
				}
			} else if *add {
				gout, err := geoip.AddGeoIPEntry(gin, *code, splitCSV(*value), *reverseMatch)
				if err != nil {
					panic(err)
				}
				if err = geoip.SaveGeoIP(gout, *out); err != nil {
					panic(err)
				}
			} else {
				gout := geoip.CutGeoIPCodes(gin, strings.Split(*keep, ","), *trimipv6)
				if err = geoip.SaveGeoIP(gout, *out); err != nil {
					panic(err)
				}
			}
			return
		}
	case "geosite", "geodat":
		{
			gin, err := geosite.LoadGeoSite(*in)
			if err != nil {
				panic(err)
			}
			if *show {
				fmt.Println(geosite.GetGeoSiteCodes(gin))
			} else if *search != "" {
				for _, x := range geosite.Search(gin, *search) {
					fmt.Println(x)
				}
			} else if *add {
				parsedDomainType, err := parseDomainType(*domainType)
				if err != nil {
					panic(err)
				}
				gout, err := geosite.AddGeoSiteEntry(gin, *code, splitCSV(*value), parsedDomainType)
				if err != nil {
					panic(err)
				}
				if err = geosite.SaveGeoSite(gout, *out); err != nil {
					panic(err)
				}
			} else {
				gout := geosite.CutGeoSiteCodes(gin, strings.Split(*keep, ","))
				if err = geosite.SaveGeoSite(gout, *out); err != nil {
					panic(err)
				}
			}
			return
		}
	default:
		{
			flag.Usage()
		}
	}
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	items := strings.Split(value, ",")
	out := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func parseDomainType(value string) (router.Domain_Type, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "plain":
		return router.Domain_Plain, nil
	case "regex", "regexp":
		return router.Domain_Regex, nil
	case "domain", "rootdomain":
		return router.Domain_Domain, nil
	case "full":
		return router.Domain_Full, nil
	default:
		return router.Domain_Plain, fmt.Errorf("unknown domain type %q", value)
	}
}
