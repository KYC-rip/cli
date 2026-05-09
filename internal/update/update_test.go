package update

import "testing"

func TestSemverNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"v0.1.6", "v0.1.5", true},
		{"v0.1.5", "v0.1.5", false},
		{"v0.1.4", "v0.1.5", false},
		{"v0.2.0", "v0.1.99", true},
		{"v1.0.0", "v0.99.99", true},
		{"0.1.6", "v0.1.5", true},
		{"v0.1.6", "dev", true},      // any vN > "dev"
		{"v0.1.6", "0.1.5", true},
	}
	for _, tc := range cases {
		if got := semverNewer(tc.a, tc.b); got != tc.want {
			t.Errorf("semverNewer(%q,%q)=%v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestAssetName(t *testing.T) {
	a, ext := assetName("0.1.6")
	if a == "" || ext == "" {
		t.Fatalf("empty asset for current platform: %q %q", a, ext)
	}
}
