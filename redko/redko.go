package redko

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

type Block struct {
	Key   string
	CIDRs []string
}

func ParseIPBlocks(r io.Reader) ([]Block, error) {
	scanner := bufio.NewScanner(r)
	blocks := make([]Block, 0)
	indexByKey := make(map[string]int)
	current := -1
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if ip := net.ParseIP(line); ip != nil {
			if current < 0 {
				return nil, fmt.Errorf("line %d: IP %q without a key", lineNo, line)
			}
			blocks[current].CIDRs = append(blocks[current].CIDRs, ipToCIDR(ip))
			continue
		}

		key := strings.ToUpper(line)
		if idx, ok := indexByKey[key]; ok {
			current = idx
			continue
		}

		blocks = append(blocks, Block{Key: key})
		current = len(blocks) - 1
		indexByKey[key] = current
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for i, block := range blocks {
		if len(block.CIDRs) == 0 {
			return nil, fmt.Errorf("key %q has no IPs", blocks[i].Key)
		}
	}

	return blocks, nil
}

func ipToCIDR(ip net.IP) string {
	if ipv4 := ip.To4(); ipv4 != nil {
		return fmt.Sprintf("%s/32", ipv4.String())
	}
	return fmt.Sprintf("%s/128", ip.To16().String())
}
