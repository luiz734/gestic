package restic

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

type Snapshot struct {
	Id      string
	Date    time.Time
	Size    uint64
	SizeStr string
	Path    string
}

func (s Snapshot) String() string {
	layout := "2006-01-02 15:04:05"
	return fmt.Sprintf("%s\t%s\t%s", s.Id, s.Date.Format(layout), s.SizeStr)
}

func GetSnapshopts() ([]Snapshot, error) {
	var err error
	args := []string{"-r", "/mnt/storage/__restic", "snapshots"}
	var cmd *exec.Cmd
	if cmd = exec.Command("restic", args...); cmd == nil {
		return []Snapshot{}, fmt.Errorf("Can't execute restic command: %w", err)
	}
	// Temporary key id: cb7483d865dc145c31176c6bfecc299d1381b82a40adb0ed4cb9ff5cd53f5269
	key := "123"
	var stdin bytes.Buffer
	stdin.Write([]byte(key))
	cmd.Stdin = &stdin
	// cmd.Stdin = os.Stdin
	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return []Snapshot{}, fmt.Errorf("Error return from restic command: %w", err)
	}
	snapshots, err := parseCmdSnapshots(output)
	if err != nil {
		return []Snapshot{}, fmt.Errorf("Parsing command snapshot: %w", err)
	}
	snapshots, err = checkDirectoriesConsistency(snapshots)
	if err != nil {
		return []Snapshot{}, fmt.Errorf("Directoy consistency error: %w", err)
	}

	return snapshots, nil

}

func parseCmdSnapshots(rawOutput []byte) ([]Snapshot, error) {
	var snapshots []Snapshot

	// Split and remove header/footer
	tokens := strings.Split(string(rawOutput), "\n")
	start := 2
	end := len(tokens) - 3

	if end <= start {
		errMsg := fmt.Errorf("expected at least 1 snapshot")
		return []Snapshot{}, errMsg
	}

	for _, t := range tokens[start:end] {
		fields := strings.Fields(t)
		// TODO: remove hardcoded timezone
		layout := "2006-01-02 15:04:05-07:00"
		timeStr := fmt.Sprintf("%s %s%s",
			fields[1],
			fields[2],
			"-03:00",
		)
		t, err := time.Parse(layout, timeStr)
		if err != nil {
			panic(err)
		}
		// Use the format: X.YYY Gib
		sizeStr := fmt.Sprintf("%s%s",
			fields[len(fields)-2],
			fields[len(fields)-1],
		)

		s := Snapshot{
			Id:      fields[0],
			Date:    t,
			Size:    uint64(123),
			SizeStr: sizeStr,
		}
		snapshots = append(snapshots, s)
	}

	return snapshots, nil
}

func snapshotContainsTime(s []Snapshot, t time.Time) int {
	for index, x := range s {
		if x.Date.Compare(t) == 0 {
			return index
		}
	}
	return -1
}

// Checks if the output of `restic snapshots` has a directory
// associated with each entry. It compares the time for the
// command output with the filename in the snapshots directory
func checkDirectoriesConsistency(s []Snapshot) ([]Snapshot, error) {
	rootDir := "/home/tohru/tmp/restic/snapshots"
	dateTimeLayout := "2006-01-02T15:04:05-07:00"

	dirEntries, err := os.ReadDir(rootDir)
	if err != nil {
		errMsg := fmt.Errorf("directory missing or not mounted: %w", err)
		return []Snapshot{}, errMsg
	}
	for _, entry := range dirEntries {
		t, err := time.Parse(dateTimeLayout, entry.Name())
		// The directory has a symlink to the most recent snapshot. We ignore it
		if entry.Name() == "latest" {
			continue
		}
		// Bad naming and it is not the previous case. It should never happen.
		if err != nil {
			panic(err)
		}

		index := snapshotContainsTime(s, t)
		if index == -1 {
			errMsg := fmt.Errorf("mismatch entries for snapshot %s", entry.Name())
			return []Snapshot{}, errMsg
		}
		s[index].Path = path.Join(rootDir, entry.Name())

	}

	return s, nil
}
