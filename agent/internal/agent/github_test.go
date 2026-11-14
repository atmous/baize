package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"testing"
)

func TestVersionNumber(t *testing.T) {
	cases := []struct {
		value string
		want  int64
	}{
		{"20261112", 20261112},
		{"v20261112", 20261112},
		{"V20261112", 20261112},
		{" 20261112 ", 20261112},
		{"dev", 0},
		{"", 0},
		{"1.2.3", 0},
	}
	for _, tc := range cases {
		if got := versionNumber(tc.value); got != tc.want {
			t.Errorf("versionNumber(%q) = %d, want %d", tc.value, got, tc.want)
		}
	}
}

func TestIsUpdateCandidate(t *testing.T) {
	const selfHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const otherHash = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	client := &githubClient{currentHash: selfHash}
	cases := []struct {
		name    string
		version string
		digest  string
		current string
		want    bool
	}{
		{"newer version always qualifies", "20261113", otherHash, "20261112", true},
		{"newer version qualifies without a digest", "20261113", "", "20261112", true},
		{"older version never qualifies", "20261111", otherHash, "20261112", false},
		{"same version with different bytes qualifies", "20261112", otherHash, "20261112", true},
		{"same version with identical bytes does not qualify", "20261112", selfHash, "20261112", false},
		{"same version without a digest does not qualify", "20261112", "", "20261112", false},
		{"dotted versions fall back to a hash comparison", "1.2.3", otherHash, "1.2.3", true},
		{"dotted versions with identical bytes do not qualify", "1.2.3", selfHash, "1.2.3", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := client.isUpdateCandidate(tc.version, tc.digest, tc.current); got != tc.want {
				t.Errorf("isUpdateCandidate(%q, %q, %q) = %v, want %v", tc.version, tc.digest, tc.current, got, tc.want)
			}
		})
	}
}

func TestCurrentBinaryHash(t *testing.T) {
	// The hash is read from the running test binary itself, so verify it
	// against an independent hash of os.Executable and that it is cached.
	client := &githubClient{}
	hash, err := client.currentBinaryHash()
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) != 64 || strings.Trim(hash, "0123456789abcdef") != "" {
		t.Fatalf("currentBinaryHash returned a malformed digest: %q", hash)
	}
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	want := sha256.New()
	if _, err := io.Copy(want, file); err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(want.Sum(nil)); got != hash {
		t.Fatalf("currentBinaryHash = %s, want %s", hash, got)
	}
	if again, err := client.currentBinaryHash(); err != nil || again != hash {
		t.Fatalf("currentBinaryHash was not cached: %q, %v", again, err)
	}
}
