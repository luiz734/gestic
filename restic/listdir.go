package restic

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/dustin/go-humanize"
)

type DirData struct {
	Children     []DirData
	Path         string
	PathReadable string
	Size         uint64
	SizeRadable  string
}

func (d DirData) Less() {

}

func sortEntriesBySize(entries []DirData) []DirData {
	cmpFunc := func(a, b DirData) int {
		return int(b.Size) - int(a.Size)
	}
	slices.SortFunc(entries, cmpFunc)
	return entries
}

func DirSize(path string) (DirData, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return DirData{
		Path: path,
		Size: uint64(size),
	}, err
}

func GetDirEntries(dirPath string) ([]DirData, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return []DirData{}, fmt.Errorf("can't read dir: %w", err)
	}

	var entriesData []DirData
	for _, e := range entries {
		subDirPath := path.Join(dirPath, e.Name())
		dirData, err := DirSize(subDirPath)
		if err != nil {
			return []DirData{}, fmt.Errorf("can't get dir size: %w", err)
		}
		dirData.SizeRadable = humanize.Bytes(dirData.Size)
		dirData.PathReadable, err = filepath.Rel(dirPath, subDirPath)
		if err != nil {
			return []DirData{}, fmt.Errorf("can't get relative path: %w", err)
		}
		entriesData = append(entriesData, dirData)

	}

	return sortEntriesBySize(entriesData), nil
}
