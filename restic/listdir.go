package restic

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/dustin/go-humanize"
)

// DirData represents a file or directory with its properties and children.
type DirData struct {
	Children     []*DirData // List of child entries (empty for files)
	Path         string     // Full absolute path
	PathReadable string     // Name relative to parent directory
	Size         int64      // Size (file size or sum of children's sizes)
	SizeReadable string     // Human-readable size
	IsDir        bool       // True if entry is a directory
}

// GetDirEntries returns the immediate entries of dirPath, with directories' Children fields recursively populated.
func GetDirEntries(root string) (*DirData, error) {
	maxIO := 100
	semaphore := make(chan struct{}, maxIO)

	var walk func(string) *DirData
	walk = func(currentPath string) *DirData {
		semaphore <- struct{}{}
		entries, err := os.ReadDir(currentPath)
		<-semaphore

		if err != nil {
			//slog.Error("Can't read directory", "path", currentPath)
			return nil
		}

		node := &DirData{
			Path:         currentPath,
			PathReadable: "/" + filepath.Base(currentPath),
			Children:     make([]*DirData, 0, len(entries)),
		}

		var wg sync.WaitGroup
		var mu sync.Mutex

		for _, entry := range entries {
			if entry.IsDir() {
				nextPath := filepath.Join(currentPath, entry.Name())
				wg.Add(1)

				go func(path string) {
					defer wg.Done()
					childNode := walk(path)

					if childNode != nil {
						mu.Lock()
						node.Children = append(node.Children, childNode)
						node.Size += childNode.Size
						mu.Unlock()
					}
				}(nextPath)

			} else {
				info, err := entry.Info()
				if err != nil {
					//slog.Error("Can't read file", "path", filepath.Join(currentPath, entry.Name()))
					continue
				}

				childNode := &DirData{
					Path:         filepath.Join(currentPath, entry.Name()),
					PathReadable: entry.Name(),
					Size:         info.Size(),
					SizeReadable: humanize.Bytes(uint64(info.Size())),
				}

				mu.Lock()
				node.Children = append(node.Children, childNode)
				node.Size += info.Size()
				mu.Unlock()
			}
		}

		wg.Wait()
		node.SizeReadable = humanize.Bytes(uint64(node.Size))
		return node
	}

	return walk(root), nil
}
