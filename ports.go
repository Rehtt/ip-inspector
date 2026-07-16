package main

import (
	"fmt"
	"strconv"
	"strings"
)

func parsePorts(value string) ([]int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("-ports is required")
	}

	parts := strings.Split(value, ",")
	ports := make([]int, 0, len(parts))
	seen := make(map[int]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("invalid empty port in %q", value)
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return nil, fmt.Errorf("invalid port %q: must be an integer between 1 and 65535", part)
			}
		}

		port, err := strconv.Atoi(part)
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("invalid port %q: must be an integer between 1 and 65535", part)
		}
		if _, ok := seen[port]; ok {
			continue
		}
		seen[port] = struct{}{}
		ports = append(ports, port)
	}

	return ports, nil
}
