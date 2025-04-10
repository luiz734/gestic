package restic

import (
	"fmt"
	"os"
	"path"
	"slices"

	"github.com/dustin/go-humanize"
)

// DirData represents a file or directory with its properties and children.
type DirData struct {
	Children     []DirData // List of child entries (empty for files)
	Path         string    // Full absolute path
	PathReadable string    // Name relative to parent directory
	Size         uint64    // Size (file size or sum of children's sizes)
	SizeReadable string    // Human-readable size
}

// GetDirEntries returns the immediate entries of dirPath, with directories' Children fields recursively populated.
func GetDirEntries(dirPath string) ([]DirData, error) {
	// Read all entries in the directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("can't read dir: %w", err)
	}

	var entriesData []DirData
	for _, e := range entries {
		// Construct the full path for this entry
		entryPath := path.Join(dirPath, e.Name())

		// Get file or directory info
		info, err := e.Info()
		if err != nil {
			return nil, fmt.Errorf("can't get entry info: %w", err)
		}

		if info.IsDir() {
			// Directory: recursively get its children
			children, err := GetDirEntries(entryPath)
			if err != nil {
				return nil, err
			}

			// Calculate total size as the sum of children's sizes
			var size uint64
			for _, child := range children {
				size += child.Size
			}

			// Create DirData for the directory
			dirData := DirData{
				Children:     children,
				Path:         entryPath,
				Size:         size,
				SizeReadable: humanize.Bytes(size),
				PathReadable: e.Name(), // Relative to parent, matches original behavior
			}
			entriesData = append(entriesData, dirData)
		} else {
			// File: no children, size is the file size
			size := uint64(info.Size())
			dirData := DirData{
				Children:     []DirData{}, // Empty for files
				Path:         entryPath,
				Size:         size,
				SizeReadable: humanize.Bytes(size),
				PathReadable: e.Name(), // Relative to parent
			}
			entriesData = append(entriesData, dirData)
		}
	}

	// Sort entries by size in descending order
	slices.SortFunc(entriesData, func(a, b DirData) int {
		return int(b.Size) - int(a.Size)
	})

	return entriesData, nil
}
