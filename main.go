package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Snapshot struct {
	id      string
	date    time.Time
	size    uint64
	sizeStr string
	path    string
}

func (s Snapshot) String() string {
	return fmt.Sprintf("%s\t%s\t%s\n", s.id, (s.date.Format("2006-01-02 15:04:05")), s.sizeStr)
}

func parseCmdSnapshots(rawOutput []byte) []Snapshot {
	var snapshots []Snapshot

	// Split and remove header/footer
	tokens := strings.Split(string(rawOutput), "\n")
	for _, t := range tokens[2 : len(tokens)-3] {
		lineData := strings.Fields(t)
		// TODO: remove hardcoded timezone
		layout := "2006-01-02 15:04:05-07:00"
		timeStr := fmt.Sprintf("%s %s%s",
			strings.TrimSpace(lineData[1]),
			strings.TrimSpace(lineData[2]),
			"-03:00",
		)
		t, err := time.Parse(layout, timeStr)
		if err != nil {
			panic(err)
		}

		sizeStr := fmt.Sprintf("%s%s",
			strings.TrimSpace(lineData[len(lineData)-2]),
			strings.TrimSpace(lineData[len(lineData)-1]),
		)

		s := Snapshot{
			id:      strings.TrimSpace(lineData[0]),
			date:    t,
			size:    uint64(123),
			sizeStr: sizeStr,
		}
		snapshots = append(snapshots, s)
	}

	return snapshots
}

func snapshotContainsTime(s []Snapshot, t time.Time) (int, error) {
	for index, x := range s {
		fmt.Printf("%s\n", x.date.String())
		if x.date.Compare(t) == 0 {
			return index, nil
		}
	}
	err := errors.New("snapshot does not contain time")
	return -1, err
}
func dirSize(path string) (uint64, error) {
	var size int64
	err := filepath.Walk(path, func(dir string, info os.FileInfo, err error) error {
		fmt.Printf("%s\n", dir)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return uint64(size), err
}

func fillSizeAndPath(s []Snapshot) []Snapshot {
	rootDir := "/home/tohru/tmp/restic/snapshots"
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		errMsg := fmt.Sprintf("Directory not mounted: %s", err)
		panic(errMsg)
	}
	for _, entry := range entries {
		layout := "2006-01-02T15:04:05-07:00"
		t, err := time.Parse(layout, entry.Name())
		if entry.Name() == "latest" {
			continue
		}
		if err != nil {
			panic(err)
		}
		_, err = snapshotContainsTime(s, t)
		if err != nil {
			panic("Missing matching time: " + err.Error())
		}
		//fullPath := path.Join(rootDir, entry.Name())
		//dSize, err := dirSize(fullPath)
		//if err != nil {
		//	s[tIndex].size = dSize
		//}
	}

	return s
}

func main() {
	var err error
	_ = err
	args := []string{"-r", "/mnt/storage/__restic", "snapshots"}
	var cmd *exec.Cmd
	if cmd = exec.Command("restic", args...); cmd == nil {
		panic(fmt.Sprintf("Error running restic command: %v", cmd))
	}
	// cb7483d865dc145c31176c6bfecc299d1381b82a40adb0ed4cb9ff5cd53f5269
	key := "123"
	var stdin bytes.Buffer
	stdin.Write([]byte(key))
	cmd.Stdin = &stdin
	// cmd.Stdin = os.Stdin

	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	s := Snapshot{
		id:   "1234",
		date: time.Now(),
		size: 1231412,
	}
	_ = s

	output, err := cmd.Output()
	if err != nil {
		panic(fmt.Sprintf("Error reading output", cmd))
	}
	snapshots := parseCmdSnapshots(output)
	snapshots = fillSizeAndPath(snapshots)
	// fmt.Printf("%s", output)
	for _, s := range snapshots {
		fmt.Printf("%s", s)
	}
}
