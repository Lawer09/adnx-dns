package util

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"regexp"
	"strings"
)

var subRe = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)

func IsValidIPv4(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	return parsed != nil && parsed.To4() != nil
}

func NormalizeSubdomain(s string) string {
	return strings.ToLower(strings.Trim(strings.TrimSpace(s), "."))
}

func ValidateSubdomain(s string) error {
	s = NormalizeSubdomain(s)
	if s == "" {
		return errors.New("subdomain is empty")
	}
	if len(s) > 63 {
		return errors.New("subdomain too long")
	}
	if !subRe.MatchString(s) || strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return errors.New("subdomain must be lowercase letters, numbers or hyphen, and start with a letter")
	}
	return nil
}

func RandomLowercase(n int) string {
	if n <= 0 {
		n = 8
	}
	const letters = "abcdefghijklmnopqrstuvwxyz"
	var b strings.Builder
	for i := 0; i < n; i++ {
		v, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b.WriteByte(letters[v.Int64()])
	}
	return b.String()
}

func JSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func SplitFQDN(fqdn string) (subdomain, domain string, err error) {
	fqdn = strings.ToLower(strings.Trim(strings.TrimSpace(fqdn), "."))
	parts := strings.Split(fqdn, ".")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid fqdn")
	}
	subdomain = parts[0]
	domain = strings.Join(parts[1:], ".")
	return subdomain, domain, nil
}
