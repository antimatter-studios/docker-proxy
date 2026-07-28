package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func paths(t *testing.T) (httpPath, streamPath string) {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "default.conf"), filepath.Join(dir, "stream.conf")
}

func rejects() ([]byte, error) {
	return []byte("nginx: [emerg] unknown directive"), errors.New("exit status 1")
}

func accepts() ([]byte, error) {
	return []byte("syntax is ok"), nil
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

// The bug this whole function exists for: refusing to reload protected the running
// nginx, but the rejected config stayed on disk, so the next container restart loaded
// it and nginx would not start at all.
func TestRejectedConfigDoesNotStayOnDisk(t *testing.T) {
	httpPath, streamPath := paths(t)
	if err := os.WriteFile(httpPath, []byte("server { listen 80; }"), 0644); err != nil {
		t.Fatal(err)
	}

	status, err := applyConfig(httpPath, streamPath, "this is not valid nginx", "", rejects)
	if err == nil {
		t.Fatal("a rejected config was accepted")
	}
	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", status, http.StatusBadRequest)
	}
	if got := read(t, httpPath); got != "server { listen 80; }" {
		t.Errorf("the working config was not restored, on disk is: %q", got)
	}
}

// The gap in the first version of the fix: with nothing to restore it wrote nothing
// back, so a rejected config survived on a proxy that had not been configured yet —
// exactly the first-boot case, where nothing else would put a valid file there.
func TestRejectedConfigIsRemovedWhenThereWasNoPrevious(t *testing.T) {
	httpPath, streamPath := paths(t)

	if _, err := applyConfig(httpPath, streamPath, "this is not valid nginx", "", rejects); err == nil {
		t.Fatal("a rejected config was accepted")
	}
	if _, err := os.Stat(httpPath); !os.IsNotExist(err) {
		t.Errorf("rejected config left behind: %q", read(t, httpPath))
	}
}

// A stream config that did not exist before must not survive rejection either; nginx
// loads stream.d on startup the same way.
func TestRejectedStreamConfigIsRemoved(t *testing.T) {
	httpPath, streamPath := paths(t)
	if err := os.WriteFile(httpPath, []byte("server { listen 80; }"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := applyConfig(httpPath, streamPath, "server { listen 80; }", "bad stream config", rejects); err == nil {
		t.Fatal("a rejected config was accepted")
	}
	if _, err := os.Stat(streamPath); !os.IsNotExist(err) {
		t.Errorf("rejected stream config left behind: %q", read(t, streamPath))
	}
}

func TestAcceptedConfigIsWritten(t *testing.T) {
	httpPath, streamPath := paths(t)

	status, err := applyConfig(httpPath, streamPath, "server { listen 80; }", "server { listen 25; }", accepts)
	if err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	if status != 0 {
		t.Errorf("status = %d, want 0 on success", status)
	}
	if got := read(t, httpPath); got != "server { listen 80; }" {
		t.Errorf("http config = %q", got)
	}
	if got := read(t, streamPath); got != "server { listen 25; }" {
		t.Errorf("stream config = %q", got)
	}
}

// An empty stream section means "no stream config", not "an empty one": nginx fails to
// start on a stream block with no servers in it.
func TestEmptyStreamConfigRemovesTheFile(t *testing.T) {
	httpPath, streamPath := paths(t)
	if err := os.WriteFile(streamPath, []byte("server { listen 25; }"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := applyConfig(httpPath, streamPath, "server { listen 80; }", "", accepts); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	if _, err := os.Stat(streamPath); !os.IsNotExist(err) {
		t.Errorf("stream config should have been removed, contains: %q", read(t, streamPath))
	}
}
