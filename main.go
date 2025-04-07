package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type Snapshot struct {
	id   string
	date time.Time
	size uint64
}

func (s Snapshot) String() string {
	return fmt.Sprintf("%s\t%s\t%s\n", s.id, (s.date.Format("2006-02-01 15:04:05")), humanize.Bytes(s.size))
}

func parseCmdSnapshots(rawOutput []byte) []Snapshot {
	var snapshots []Snapshot

	// Split and remove header/footer
	tokens := strings.Split(string(rawOutput), "\n")
	for _, t := range tokens[2 : len(tokens)-3] {
		lineData := strings.Fields(t)
		layout := "2006-01-02 15:04:05"
		timeStr := fmt.Sprintf("%s %s",
			strings.TrimSpace(lineData[1]),
			strings.TrimSpace(lineData[2]),
		)
		t, err := time.Parse(layout, timeStr)
		if err != nil {
			panic(err)
		}
		s := Snapshot{
			id:   strings.TrimSpace(lineData[0]),
			date: t,
			size: uint64(123),
		}
		snapshots = append(snapshots, s)
		fmt.Printf("%s", s)
	}

	return snapshots
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
	parseCmdSnapshots(output)
	// fmt.Printf("%s", output)
}
