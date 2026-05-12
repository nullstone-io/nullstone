package modules

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
)

// HashArchiveContents returns a hex-encoded sha256 over a canonical digest of the
// archive's regular file entries: for each file, sorted by name, the line
// "<sha256(content)>  <name>\n" is fed into a running sha256.
//
// This intentionally ignores tar/gzip framing (mtime, ownership, entry order, gzip
// metadata) so an identical source produces identical digests across runs. The format
// matches Go's golang.org/x/mod/sumdb/dirhash.Hash1 line shape but emits the final
// digest as hex rather than "h1:base64" so it round-trips cleanly through plain text
// fields.
//
// This function MUST stay byte-for-byte in sync with the equivalent helper in
// oracle (oracle/manifest/checksum.go); a fixture test in each repo pins the
// expected output for the same input.
func HashArchiveContents(data []byte, ext string) (string, error) {
	ext = strings.TrimPrefix(ext, ".")
	switch ext {
	case "tgz", "tar.gz":
	default:
		return "", fmt.Errorf("unsupported archive extension %q", ext)
	}

	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("gunzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	type entry struct {
		name string
		sum  [sha256.Size]byte
	}
	var entries []entry
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}
		fh := sha256.New()
		if _, err := io.Copy(fh, tr); err != nil {
			return "", fmt.Errorf("hash %q: %w", hdr.Name, err)
		}
		var sum [sha256.Size]byte
		copy(sum[:], fh.Sum(nil))
		entries = append(entries, entry{name: hdr.Name, sum: sum})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })

	h := sha256.New()
	for _, e := range entries {
		fmt.Fprintf(h, "%s  %s\n", hex.EncodeToString(e.sum[:]), e.name)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
