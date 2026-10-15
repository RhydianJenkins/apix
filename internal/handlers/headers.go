package handlers

import (
	"fmt"
	"strings"
)

func ParseHeaders(headers []string) map[string]string {
	headerMap := make(map[string]string)
	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			headerMap[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		} else {
			fmt.Printf("Warning: Malformed header '%s' - expected format 'Key: Value'\n", header)
		}
	}
	return headerMap
}
