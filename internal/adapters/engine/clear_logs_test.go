package engine

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestClearLogFileTruncatesRegularFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "container.log")
	if err := os.WriteFile(path, []byte("line one\nline two\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := clearLogFile(path); err != nil {
		t.Fatalf("clearLogFile() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Fatalf("log size = %d, want 0", info.Size())
	}
}

func TestDockerLogClearHelperCommandIsConstrained(t *testing.T) {
	logPath := "/var/lib/docker/containers/abc/abc-json.log"
	cmd := dockerLogClearHelperCmd(logPath)
	want := []string{
		"docker", "run", "--rm",
		"--network", "none",
		"--read-only",
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges",
		"-v", "/var/lib/docker:/var/lib/docker",
		"alpine:3.22",
		"truncate", "-s", "0", logPath,
	}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("helper args = %#v, want %#v", cmd.Args, want)
	}
}

func TestIsDockerJSONLogPath(t *testing.T) {
	if !isDockerJSONLogPath("/var/lib/docker/containers/abc/abc-json.log") {
		t.Fatal("valid Docker json log path rejected")
	}
	if isDockerJSONLogPath("/tmp/not-a-log") {
		t.Fatal("invalid Docker log path accepted")
	}
}

func TestClearLogFileRejectsRelativePath(t *testing.T) {
	if err := clearLogFile("container.log"); err == nil {
		t.Fatal("clearLogFile() accepted relative path")
	}
}
