package artifacts

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
)

func NewTargzer(w io.Writer, name string) *Targzer {
	gzip := gzip.NewWriter(w)
	gzip.Name = name
	return &Targzer{
		gzip:    gzip,
		tarball: tar.NewWriter(gzip),
	}
}

type Targzer struct {
	gzip    *gzip.Writer
	tarball *tar.Writer
}

func (t *Targzer) Close() {
	if t.tarball != nil {
		t.tarball.Close()
	}
	if t.gzip != nil {
		t.gzip.Close()
	}
}

func (t Targzer) AddFile(header *tar.Header, r io.Reader) error {
	if err := t.tarball.WriteHeader(header); err != nil {
		return fmt.Errorf("error writing file header %s: %w", header.Name, err)
	}
	if r == nil {
		// If no reader was specified, there is nothing to write
		// This happens when we are creating a directory
		return nil
	}
	_, err := io.Copy(t.tarball, r)
	if err != nil {
		return fmt.Errorf("error writing file into archive %s: %w", header.Name, err)
	}
	return nil
}
