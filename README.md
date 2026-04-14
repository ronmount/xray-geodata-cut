# xray-geodata-cut

Cut unneeded data from geoip.dat or geosite.dat, build geoip.dat from ASNs, or add new records into geoip/geodat files

```
Usage of xray-geodata-cut:
  -in string
        Path to GeoData file / ASNs split by comma
  -add
        Add entries to GeoIP or GeoSite file
  -code string
        GeoIP or GeoSite code to add or update
  -domain-type string
        GeoSite domain rule type: plain, regex, domain/rootdomain, full (default "plain")
  -keep string
        GeoIP or GeoSite codes to keep (private is always kept for GeoIP) (default "cn,private,geolocation-!cn")
  -out string
        Path to processed file
  -reverse-match
        Set reverse_match for newly created GeoIP entries
  -search string
        Search GeoIP or GeoSite Item
  -show
        Print codes in GeoIP or GeoSite file
  -trimipv6
        Trim all IPv6 ranges in GeoIP file
  -type string
        ASN (asn), GeoIP (geoip) or GeoSite/GeoDat (geosite/geodat)
  -value string
        CIDRs for GeoIP or domain rules for GeoSite, split by comma

```

ASN information comes from [https://github.com/ipverse/asn-ip/](https://github.com/ipverse/asn-ip/)

Examples for search: 

```
sh-5.1$ go run . -type asn -in 24429,4134 -search 106.124.1.2
AS4134
sh-5.1$ go run . -in /usr/local/share/xray/geoip.dat -type geoip -search 114.114.114.114
CN
sh-5.1$ go run . -in /usr/local/share/xray/geoip.dat -type geoip -search 192.0.2.1
PRIVATE
sh-5.1$ go run . -in /usr/local/share/xray/geoip.dat -type geoip -search 127.0.0.1
PRIVATE
TEST
sh-5.1$ go run . -in /usr/local/share/xray/geosite.dat -type geosite -search bilibili.com
BILIBILI
CN
GEOLOCATION-CN
sh-5.1$ go run . -in /usr/local/share/xray/geosite.dat -type geosite -search baidu.com
BAIDU
CN
GEOLOCATION-CN
sh-5.1$ go run . -in /usr/local/share/xray/geosite.dat -type geosite -search youtube.com
CATEGORY-COMPANIES
GEOLOCATION-!CN
GOOGLE
YOUTUBE
sh-5.1$ go run . -in /usr/local/share/xray/geosite.dat -type geosite -search www.netflix.com
CATEGORY-ENTERTAINMENT
GEOLOCATION-!CN
NETFLIX
```

Examples for add:

```bash
go run . -type geoip -in ./geoip.dat -out ./geoip-new.dat -add -code RU-CUSTOM -value 203.0.113.0/24,2001:db8::/32
go run . -type geodat -in ./geosite.dat -out ./geosite-new.dat -add -code MY-SITES -domain-type full -value example.com,api.example.com
go run . -type geosite -in ./geosite.dat -out ./geosite-new.dat -add -code MY-ROOTS -domain-type domain -value example.org,example.net
```

When `-add` is used and the code already exists, new CIDRs or domain rules are appended to the existing entry, while duplicates are ignored.

Import blocks from `https://redko.us/net/ip`:

```bash
go run ./cmd/redko-geoip-import -in ./geoip.dat
go run ./cmd/redko-geoip-import -in ./geoip.dat -out ./geoip-new.dat
go run ./cmd/redko-geoip-import -in ./geoip.dat -only 2gis,yandex
```

The importer expects a text format where a key line is followed by one or more plain IP addresses. Each IP is converted to `/32` or `/128` and merged into the matching GeoIP entry. If the key does not exist yet, a new GeoIP entry is created automatically.
