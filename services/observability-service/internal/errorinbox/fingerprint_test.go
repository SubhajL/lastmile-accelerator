package errorinbox

import "testing"

func TestNormalizeStack_StripsVolatileDetails(t *testing.T) {
	in := "panic: something\nmain.main /tmp/build/abc/main.go:123+0x1a\n at 0x7ffee123\n"
	out := NormalizeStack(in)
	if out == in { t.Fatalf("expected normalization change") }
	if want := "<line>"; !contains(out, want) { t.Fatalf("expected to contain %q, got %q", want, out) }
}

func TestFingerprint_StableAcrossNoise(t *testing.T) {
	msg := "typeerror: undefined is not a function"
	stack1 := NormalizeStack("/app/src/a.js:10\n/app/src/b.js:20")
	stack2 := NormalizeStack("/app/src/a.js:10+0x1\n/app/src/b.js:20+0x2")
	fp1 := Fingerprint(msg, stack1)
	fp2 := Fingerprint(msg, stack2)
	if fp1 != fp2 { t.Fatalf("expected same fingerprint, got %s vs %s", fp1, fp2) }
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (contains(s[1:], sub) || (s[:len(sub)] == sub))) ) }
