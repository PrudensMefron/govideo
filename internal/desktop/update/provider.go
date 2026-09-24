package update

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/updater"
)

var ErrUnverifiedRelease = errors.New("update release is missing its required SHA-256 checksum")

// checksumProvider prevents a missing checksum asset (or a missing entry in
// that asset) from silently downgrading the GitHub provider to an unverified
// download. The Wails GitHub provider deliberately treats checksums as
// optional; GoVideo requires one for every release artifact.
type checksumProvider struct {
	provider updater.Provider
}

func requireChecksum(provider updater.Provider) updater.Provider {
	return &checksumProvider{provider: provider}
}

func (p *checksumProvider) Name() string { return p.provider.Name() }

func (p *checksumProvider) Check(ctx context.Context, request updater.CheckRequest) (*updater.Release, error) {
	release, err := p.provider.Check(ctx, request)
	if err != nil || release == nil {
		return release, err
	}
	verification := release.Verification
	if verification == nil || !strings.EqualFold(verification.DigestAlgo, "sha256") || len(verification.Digest) != 32 {
		return nil, ErrUnverifiedRelease
	}
	return release, nil
}

func (p *checksumProvider) Download(ctx context.Context, release *updater.Release, destination io.Writer, progress func(int64, int64)) error {
	return p.provider.Download(ctx, release, destination, progress)
}
