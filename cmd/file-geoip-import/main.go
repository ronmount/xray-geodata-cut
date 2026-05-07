package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/yichya/xray-geodata-cut/geoip"
)

func main() {
	in := flag.String("in", "", "Path to input geoip.dat")
	out := flag.String("out", "", "Path to output geoip.dat (default: overwrite -in)")
	file := flag.String("file", "./ddos-guard.txt", "Path to text file with IPs or CIDRs, one per line")
	code := flag.String("code", "", "GeoIP code to create or update (default: derived from -file name)")
	reverseMatch := flag.Bool("reverse-match", false, "Set reverse_match for a newly created GeoIP entry")
	flag.Parse()

	inputPath := strings.TrimSpace(*in)
	if inputPath == "" {
		fail(fmt.Errorf("-in is required"))
	}

	outputPath := strings.TrimSpace(*out)
	if outputPath == "" {
		outputPath = inputPath
	}

	filePath := strings.TrimSpace(*file)
	if filePath == "" {
		fail(fmt.Errorf("-file is required"))
	}

	entryCode := strings.TrimSpace(*code)
	if entryCode == "" {
		entryCode = defaultCodeFromPath(filePath)
	}

	cidrs, err := loadCIDRs(filePath)
	if err != nil {
		fail(err)
	}

	geoIPList, err := geoip.LoadGeoIP(inputPath)
	if err != nil {
		fail(err)
	}

	geoIPList, err = geoip.AddGeoIPEntry(geoIPList, entryCode, cidrs, *reverseMatch)
	if err != nil {
		fail(fmt.Errorf("import %q: %w", entryCode, err))
	}

	if err := geoip.SaveGeoIP(geoIPList, outputPath); err != nil {
		fail(err)
	}

	fmt.Printf("imported %d record(s) into %s and saved to %s\n", len(cidrs), strings.ToUpper(entryCode), outputPath)
}

func loadCIDRs(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	cidrs := make([]string, 0)
	seen := make(map[string]struct{})
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		value, err := normalizeCIDR(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, lineNo, err)
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		cidrs = append(cidrs, value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(cidrs) == 0 {
		return nil, fmt.Errorf("%s: no IPs or CIDRs found", path)
	}

	return cidrs, nil
}

func normalizeCIDR(value string) (string, error) {
	if ip := net.ParseIP(value); ip != nil {
		if ipv4 := ip.To4(); ipv4 != nil {
			return fmt.Sprintf("%s/32", ipv4.String()), nil
		}
		return fmt.Sprintf("%s/128", ip.To16().String()), nil
	}

	ip, ipNet, err := net.ParseCIDR(value)
	if err != nil {
		return "", fmt.Errorf("invalid IP or CIDR %q", value)
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	} else {
		ip = ip.To16()
	}
	maskSize, _ := ipNet.Mask.Size()
	return fmt.Sprintf("%s/%d", ip.String(), maskSize), nil
}

func defaultCodeFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSpace(strings.TrimSuffix(base, ext))
	if name == "" {
		return "CUSTOM"
	}
	replacer := strings.NewReplacer(" ", "-", "_", "-", ".", "-")
	return strings.ToUpper(replacer.Replace(name))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
