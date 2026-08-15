package provider

import (
	"fmt"
	"strings"
)

// NormalizePHMobile converts E.164 or local Philippine mobile numbers to the
// 11-digit format expected by Synermaxx (e.g. 09662060323).
func NormalizePHMobile(recipient string) (string, error) {
	r := strings.TrimSpace(recipient)
	r = strings.TrimPrefix(r, "+")

	switch {
	case strings.HasPrefix(r, "63") && len(r) == 12:
		return "0" + r[2:], nil
	case strings.HasPrefix(r, "09") && len(r) == 11:
		return r, nil
	case strings.HasPrefix(r, "9") && len(r) == 10:
		return "0" + r, nil
	default:
		return "", fmt.Errorf("unsupported Philippine mobile format: %q", recipient)
	}
}
