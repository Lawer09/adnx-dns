package util

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"net"
	"regexp"
	"strings"
)

var subRe = regexp.MustCompile(`^[a-z]+$`)

func IsValidIPv4(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	return parsed != nil && parsed.To4() != nil
}

func NormalizeSubdomain(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func ValidateSubdomain(s string) error {
	s = NormalizeSubdomain(s)
	if s == "" { return errors.New("subdomain is empty") }
	if !subRe.MatchString(s) { return errors.New("subdomain must contain lowercase letters only") }
	return nil
}

func RandomLowercase(n int) string {
	letters := "abcdefghijklmnopqrstuvwxyz"
	var b strings.Builder
	for i:=0;i<n;i++ {
		v, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b.WriteByte(letters[v.Int64()])
	}
	return b.String()
}

func JSON(v any) []byte {
	b,_ := json.Marshal(v)
	return b
}
