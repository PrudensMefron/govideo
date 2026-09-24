package update

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/updater"
)

type fakeProvider struct {
	release *updater.Release
	err     error
}

func (f *fakeProvider) Name() string { return "fake" }
func (f *fakeProvider) Check(context.Context, updater.CheckRequest) (*updater.Release, error) {
	return f.release, f.err
}
func (f *fakeProvider) Download(_ context.Context, _ *updater.Release, dst io.Writer, progress func(int64, int64)) error {
	_, err := io.Copy(dst, bytes.NewBufferString("artifact"))
	return err
}

func TestChecksumProviderRejectsMissingVerification(t *testing.T) {
	provider := requireChecksum(&fakeProvider{release: &updater.Release{Version: "1.2.3"}})
	_, err := provider.Check(context.Background(), updater.CheckRequest{})
	if !errors.Is(err, ErrUnverifiedRelease) {
		t.Fatalf("expected ErrUnverifiedRelease, got %v", err)
	}
}

func TestChecksumProviderRequiresSHA256(t *testing.T) {
	provider := requireChecksum(&fakeProvider{release: &updater.Release{
		Version:      "1.2.3",
		Verification: &updater.Verification{DigestAlgo: "sha512", Digest: make([]byte, 64)},
	}})
	_, err := provider.Check(context.Background(), updater.CheckRequest{})
	if !errors.Is(err, ErrUnverifiedRelease) {
		t.Fatalf("expected ErrUnverifiedRelease, got %v", err)
	}
}

func TestChecksumProviderAcceptsSHA256(t *testing.T) {
	release := &updater.Release{
		Version:      "1.2.3",
		Verification: &updater.Verification{DigestAlgo: "sha256", Digest: make([]byte, 32)},
	}
	provider := requireChecksum(&fakeProvider{release: release})
	got, err := provider.Check(context.Background(), updater.CheckRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != release {
		t.Fatal("provider did not return the verified release")
	}
}

func TestReleaseVersion(t *testing.T) {
	for _, test := range []struct {
		version string
		want    bool
	}{
		{"dev", false},
		{"0.0.0-dev", false},
		{"", false},
		{"nightly", false},
		{"0.1.0", true},
		{"v1.2.3", true},
	} {
		if got := isReleaseVersion(test.version); got != test.want {
			t.Errorf("isReleaseVersion(%q) = %v, want %v", test.version, got, test.want)
		}
	}
}
