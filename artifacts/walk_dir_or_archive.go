package artifacts

import (
	"fmt"
	"os"
)

type WalkFunc func(file *os.File) error

type Walker interface {
	Walk(fn WalkFunc) error
}

func WalkDirOrArchive(dirOrArchive string) (Walker, error) {
	info, err := os.Stat(dirOrArchive)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("source does not exist %q", dirOrArchive)
	}
	if info.IsDir() {
		return &DirWalker{Directory: info.Name()}, nil
	}
	return &ArchiveWalker{Archive: info.Name()}, nil
}

var _ Walker = &DirWalker{}
type DirWalker struct {
	Directory string
}

func (w DirWalker) Walk(fn WalkFunc) error {
	return nil
}

var _ Walker = &ArchiveWalker{}
type ArchiveWalker struct {
	Archive string
}

func (w ArchiveWalker) Walk(fn WalkFunc) error {
	return nil
}
