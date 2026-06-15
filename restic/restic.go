package restic

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"
)

type Snapshot struct {
	Id      string
	Date    time.Time
	Size    uint64
	SizeStr string
	Path    string
}

type SnapshotsMetadata struct {
	NewerFullPath string
	NewerId       string
	OlderFullPath string
	OlderId       string
}

type SnapshotSummaryJson struct {
	BytesProcessed uint64 `json:"total_bytes_processed"`
	DataAdded      uint64 `json:"data_added"`
	DataPacked     uint64 `json:"data_added_packed"`
}

type SnapshotOutputJson struct {
	ShortId             string              `json:"short_id"`
	Time                time.Time           `json:"time"`
	SnapshotSummaryJson SnapshotSummaryJson `json:"summary"`
}

func (s Snapshot) String() string {
	layout := "2006-01-02 15:04:05"
	return fmt.Sprintf("%s\t%s\t%s", s.Id, s.Date.Format(layout), s.SizeStr)
}

func GetSnapshots(repoPath, mountPath string) ([]Snapshot, error) {
	var err error
	if _, err := os.Stat(repoPath); err != nil {
		return []Snapshot{}, fmt.Errorf("mount directory not found: %w", err)
	}

	args := []string{"-r", repoPath, "snapshots", "--json"}
	var cmd *exec.Cmd
	if cmd = exec.Command("restic", args...); cmd == nil {
		return []Snapshot{}, fmt.Errorf("can't execute restic command: %w", err)
	}
	//key := "123"
	//var stdin bytes.Buffer
	//stdin.Write([]byte(key))
	//cmd.Stdin = &stdin
	cmd.Stdin = os.Stdin
	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return []Snapshot{}, fmt.Errorf("error return from restic command: %w", err)
	}
	snapshots, err := parseCmdSnapshots(output)
	if err != nil {
		return []Snapshot{}, fmt.Errorf("parsing command snapshot: %w", err)
	}
	snapshots, err = checkDirectoriesConsistency(snapshots, mountPath)
	if err != nil {
		return []Snapshot{}, fmt.Errorf("directory consistency error: %w", err)
	}

	return snapshots, nil

}

func parseCmdSnapshots(jsonOutput []byte) ([]Snapshot, error) {
	var snapshotsJson []SnapshotOutputJson
	err := json.Unmarshal(jsonOutput, &snapshotsJson)
	if err != nil {
		return []Snapshot{}, fmt.Errorf("failed to parse snapshot details: %w", err)
	}
	var snapshots []Snapshot
	for _, snapshotJson := range snapshotsJson {
		sizeStr := formatBytes(snapshotJson.SnapshotSummaryJson.BytesProcessed)

		s := Snapshot{
			Id:      snapshotJson.ShortId,
			Date:    snapshotJson.Time,
			Size:    uint64(123),
			SizeStr: sizeStr,
		}
		snapshots = append(snapshots, s)

	}

	return snapshots, err
}

func snapshotContainsTime(s []Snapshot, t time.Time) int {
	// Fix nanosecond precision issue
	targetTime := t.Truncate(time.Second)

	for index, x := range s {
		if x.Date.Truncate(time.Second).Equal(targetTime) {
			return index
		}
	}
	return -1
}

// Checks if the output of `restic snapshots` has a directory
// associated with each entry. It compares the time for the
// command output with the filename in the snapshots directory
func checkDirectoriesConsistency(s []Snapshot, mountPath string) ([]Snapshot, error) {
	mountPath = path.Join(mountPath, "snapshots")
	if _, err := os.Stat(mountPath); err != nil {
		return []Snapshot{}, fmt.Errorf("mount directory not found: %w", err)
	}

	dateTimeLayout := "2006-01-02T15:04:05-07:00"

	dirEntries, err := os.ReadDir(mountPath)
	if err != nil {
		errMsg := fmt.Errorf("directory missing or not mounted: %w", err)
		return []Snapshot{}, errMsg
	}
	for _, entry := range dirEntries {
		// The directory has a symlink to the most recent snapshot. We ignore it
		if entry.Name() == "latest" {
			continue
		}

		t, err := time.Parse(dateTimeLayout, entry.Name())
		// Bad naming and it is not the previous case. It should never happen.
		if err != nil {
			panic(err)
		}

		index := snapshotContainsTime(s, t)
		if index == -1 {
			errMsg := fmt.Errorf("mismatch entries for snapshot %s", entry.Name())
			return []Snapshot{}, errMsg
		}
		s[index].Path = path.Join(mountPath, entry.Name())

	}

	return s, nil
}

// Don't use humanize package in this case
// More decimals is better for size comparison
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.3f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
