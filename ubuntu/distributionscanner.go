package ubuntu

import (
	"bytes"
	"context"
	"regexp"
	"runtime/trace"

	"github.com/quay/zlog"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/label"

	"github.com/quay/claircore"
	"github.com/quay/claircore/internal/indexer"
)

const (
	scannerName    = "ubuntu"
	scannerVersion = "v0.0.1"
	scannerKind    = "distribution"
)

type ubuntuRegex struct {
	release Release
	regexp  *regexp.Regexp
}

var ubuntuRegexes = []ubuntuRegex{
	{
		release: Artful,
		regexp:  regexp.MustCompile(`(?is)\bubuntu\b.*\bartful\b`),
	},
	{
		release: Bionic,
		regexp:  regexp.MustCompile(`(?is)\bubuntu\b.*\bbionic\b`),
	},
	{
		release: Cosmic,
		regexp:  regexp.MustCompile(`(?is)\bubuntu\b.*\bcosmic\b`),
	},
	{
		release: Disco,
		regexp:  regexp.MustCompile(`(?is)\bubuntu\b.*\bdisco\b`),
	},
	{
		release: Precise,
		regexp:  regexp.MustCompile(`(?is)\bubuntu\b.*\bprecise\b`),
	},
	{
		release: Trusty,
		regexp:  regexp.MustCompile(`(?is)\bubuntu\b.*\btrusty\b`),
	},
	{
		release: Xenial,
		regexp:  regexp.MustCompile(`(?is)\bubuntu\b.*\bxenial\b`),
	},
	{
		release: Eoan,
		regexp:  regexp.MustCompile(`(?is)\bubuntu\b.*\beoan\b`),
	},
	{
		release: Focal,
		regexp:  regexp.MustCompile(`(?is)\bubuntu\b.*\bfocal\b`),
	},
	{
		release: Impish,
		regexp:  regexp.MustCompile(`(?is)\bubuntu\b.*\bimpish\b`),
	},
}

const osReleasePath = `etc/os-release`
const lsbReleasePath = `etc/lsb-release`

var _ indexer.DistributionScanner = (*DistributionScanner)(nil)
var _ indexer.VersionedScanner = (*DistributionScanner)(nil)

// DistributionScanner attempts to discover if a layer
// displays characteristics of a Ubuntu distribution
type DistributionScanner struct{}

// Name implements scanner.VersionedScanner.
func (*DistributionScanner) Name() string { return scannerName }

// Version implements scanner.VersionedScanner.
func (*DistributionScanner) Version() string { return scannerVersion }

// Kind implements scanner.VersionedScanner.
func (*DistributionScanner) Kind() string { return scannerKind }

// Scan will inspect the layer for an os-release or lsb-release file
// and perform a regex match for keywords indicating the associated Ubuntu release
//
// If neither file is found a (nil,nil) is returned.
// If the files are found but all regexp fail to match an empty slice is returned.
func (ds *DistributionScanner) Scan(ctx context.Context, l *claircore.Layer) ([]*claircore.Distribution, error) {
	defer trace.StartRegion(ctx, "Scanner.Scan").End()
	ctx = baggage.ContextWithValues(ctx,
		label.String("component", "ubuntu/DistributionScanner.Scan"),
		label.String("version", ds.Version()),
		label.String("layer", l.Hash.String()))
	zlog.Debug(ctx).Msg("start")
	defer zlog.Debug(ctx).Msg("done")
	files, err := l.Files(osReleasePath, lsbReleasePath)
	if err != nil {
		zlog.Debug(ctx).Msg("didn't find an os-release or lsb release file")
		return nil, nil
	}
	for _, buff := range files {
		dist := ds.parse(buff)
		if dist != nil {
			return []*claircore.Distribution{dist}, nil
		}
	}
	return []*claircore.Distribution{}, nil
}

// parse attempts to match all Ubuntu release regexp and returns the associated
// distribution if it exists.
//
// separated into its own method to aid testing.
func (ds *DistributionScanner) parse(buff *bytes.Buffer) *claircore.Distribution {
	for _, ur := range ubuntuRegexes {
		if ur.regexp.Match(buff.Bytes()) {
			return releaseToDist(ur.release)
		}
	}
	return nil
}
