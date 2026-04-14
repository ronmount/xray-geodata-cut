package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/yichya/xray-geodata-cut/geoip"
	"github.com/yichya/xray-geodata-cut/redko"
)

func main() {
	in := flag.String("in", "", "Path to input geoip.dat")
	out := flag.String("out", "", "Path to output geoip.dat (default: overwrite -in)")
	url := flag.String("url", "https://redko.us/net/ip", "Source URL with key and IP lists")
	only := flag.String("only", "", "Optional comma-separated list of keys to import")
	timeout := flag.Duration("timeout", 15*time.Second, "HTTP timeout")
	flag.Parse()

	if strings.TrimSpace(*in) == "" {
		fmt.Fprintln(os.Stderr, "-in is required")
		os.Exit(2)
	}

	outputPath := strings.TrimSpace(*out)
	if outputPath == "" {
		outputPath = *in
	}

	resp, err := fetch(*url, *timeout)
	if err != nil {
		fail(err)
	}
	defer resp.Body.Close()

	blocks, err := redko.ParseIPBlocks(resp.Body)
	if err != nil {
		fail(err)
	}

	selected := parseOnly(*only)
	geoIPList, err := geoip.LoadGeoIP(*in)
	if err != nil {
		fail(err)
	}

	imported := 0
	for _, block := range blocks {
		if len(selected) > 0 {
			if _, ok := selected[block.Key]; !ok {
				continue
			}
		}
		geoIPList, err = geoip.AddGeoIPEntry(geoIPList, block.Key, block.CIDRs, false)
		if err != nil {
			fail(fmt.Errorf("import key %q: %w", block.Key, err))
		}
		imported++
		fmt.Printf("imported %s: %d IPs\n", block.Key, len(block.CIDRs))
	}

	if imported == 0 {
		fail(fmt.Errorf("no matching keys were imported"))
	}

	if err := geoip.SaveGeoIP(geoIPList, outputPath); err != nil {
		fail(err)
	}
	fmt.Printf("saved %d block(s) to %s\n", imported, outputPath)
}

func fetch(url string, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "xray-geodata-cut/redko-import")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return resp, nil
}

func parseOnly(value string) map[string]struct{} {
	items := strings.Split(value, ",")
	out := make(map[string]struct{}, len(items))
	for _, item := range items {
		if key := strings.ToUpper(strings.TrimSpace(item)); key != "" {
			out[key] = struct{}{}
		}
	}
	return out
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
