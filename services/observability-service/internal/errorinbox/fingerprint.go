package errorinbox

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var (
	addrRe   = regexp.MustCompile(`0x[0-9a-fA-F]+`)
	lineRe   = regexp.MustCompile(`:(\d+)(\+0x[0-9a-fA-F]+|\+\d+)?`) // strip :123 or :123+0x4 or :123+4
	tempPath = regexp.MustCompile(`/tmp/[\w\-_/]+`)
)

// NormalizeStack removes volatile details like memory addresses and line numbers.
func NormalizeStack(raw string) string {
	s := addrRe.ReplaceAllString(raw, "0x0")
	s = lineRe.ReplaceAllString(s, ":<line>")
	s = tempPath.ReplaceAllString(s, "/tmp/<path>")
	// Collapse multiple spaces, trim
	s = strings.ReplaceAll(s, "\r", "")
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" { continue }
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}

// Fingerprint hashes message + normalized stack into a stable hex string.
func Fingerprint(message, normalizedStack string) string {
	h := sha256.New()
	h.Write([]byte(strings.TrimSpace(message)))
	h.Write([]byte{'\n'})
	h.Write([]byte(normalizedStack))
	return hex.EncodeToString(h.Sum(nil))
}
