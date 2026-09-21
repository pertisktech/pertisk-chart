package chart

import (
	"strconv"
	"strings"
)

// CompareVersions compares two Helm chart version strings numerically.
// Returns >0 if a > b, <0 if a < b, and 0 if equal.
// Handles optional "v" prefix and prerelease suffixes (e.g. 1.2.3-alpha).
func CompareVersions(a, b string) int {
	a = strings.TrimSpace(strings.TrimPrefix(a, "v"))
	b = strings.TrimSpace(strings.TrimPrefix(b, "v"))
	if a == b {
		return 0
	}

	aMain, aPre, _ := strings.Cut(a, "-")
	bMain, bPre, _ := strings.Cut(b, "-")

	aParts := splitVersionParts(aMain)
	bParts := splitVersionParts(bMain)
	n := len(aParts)
	if len(bParts) > n {
		n = len(bParts)
	}

	for i := 0; i < n; i++ {
		ai, bi := 0, 0
		if i < len(aParts) {
			ai = aParts[i]
		}
		if i < len(bParts) {
			bi = bParts[i]
		}
		if ai != bi {
			return ai - bi
		}
	}

	// Release without prerelease is newer than one with prerelease.
	if aPre == "" && bPre != "" {
		return 1
	}
	if aPre != "" && bPre == "" {
		return -1
	}
	return strings.Compare(aPre, bPre)
}

func splitVersionParts(v string) []int {
	if v == "" {
		return []int{0}
	}
	raw := strings.Split(v, ".")
	parts := make([]int, len(raw))
	for i, p := range raw {
		// Take leading digits only (e.g. "10+meta" -> 10)
		num := p
		for j, r := range p {
			if r < '0' || r > '9' {
				num = p[:j]
				break
			}
		}
		if num == "" {
			parts[i] = 0
			continue
		}
		n, err := strconv.Atoi(num)
		if err != nil {
			parts[i] = 0
			continue
		}
		parts[i] = n
	}
	return parts
}
