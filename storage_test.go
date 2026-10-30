package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// setupStorage points the package globals at a temp dir and stubs the
// drive mount. Everything is restored when the test finishes.
func setupStorage(t *testing.T, mount func() error) (local, backup string) {
	t.Helper()

	dir := t.TempDir()
	local = filepath.Join(dir, "local", "training.json")
	backup = filepath.Join(dir, "backup", "training.json")

	oldLocal, oldBackup, oldMount, oldNotices := LocalFile, BackupFile, mountBackup, notices
	LocalFile, BackupFile, mountBackup = local, backup, mount
	notices = nil

	t.Cleanup(func() {
		LocalFile, BackupFile, mountBackup, notices = oldLocal, oldBackup, oldMount, oldNotices
	})

	return local, backup
}

func sampleTraining() *Training {
	return &Training{Training: []Item{
		{Name: "pushups", Value: 30, Date: time.Date(2026, 10, 30, 8, 0, 0, 0, time.UTC)},
		{Name: "run", Value: 5.25, Date: time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)},
	}}
}

func assertSameTraining(t *testing.T, got, want *Training) {
	t.Helper()

	if len(got.Training) != len(want.Training) {
		t.Fatalf("got %d items, want %d", len(got.Training), len(want.Training))
	}
	for i := range want.Training {
		g, w := got.Training[i], want.Training[i]
		if g.Name != w.Name || g.Value != w.Value || !g.Date.Equal(w.Date) {
			t.Errorf("item %d: got %+v, want %+v", i, g, w)
		}
	}
}

/* Load */

func TestLoad_MissingFileStartsEmpty(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })

	tr := &Training{}
	if err := tr.Load(local); err != nil {
		t.Fatalf("Load returned error for missing file: %v", err)
	}
	if len(tr.Training) != 0 {
		t.Errorf("expected empty training, got %d items", len(tr.Training))
	}
	if len(notices) != 1 || !strings.Contains(notices[0], "No data file found") {
		t.Errorf("expected a 'no data file' notice, got %v", notices)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })

	if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(local, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := (&Training{}).Load(local)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parsing JSON") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_ReadError(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })

	// A directory can't be read as a file, and isn't ErrNotExist.
	dir := filepath.Dir(local)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	err := (&Training{}).Load(dir)
	if err == nil {
		t.Fatal("expected error when reading a directory")
	}
	if !strings.Contains(err.Error(), "error reading file") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_DecodesDecimalAndIntegerValues(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })

	if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `{"training":[
		{"name":"a","value":5,"date":"2026-10-30T08:00:00Z"},
		{"name":"b","value":5.25,"date":"2026-10-30T09:00:00Z"}
	]}`
	if err := os.WriteFile(local, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}

	tr := &Training{}
	if err := tr.Load(local); err != nil {
		t.Fatal(err)
	}
	if got := []float64{tr.Training[0].Value, tr.Training[1].Value}; !reflect.DeepEqual(got, []float64{5, 5.25}) {
		t.Errorf("values = %v, want [5 5.25]", got)
	}
}

/* Save */

func TestSave_RoundTrip(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })

	want := sampleTraining()
	if err := want.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got := &Training{}
	if err := got.Load(local); err != nil {
		t.Fatalf("Load: %v", err)
	}
	assertSameTraining(t, got, want)
}

func TestSave_WritesBackupWhenMounted(t *testing.T) {
	local, backup := setupStorage(t, func() error { return nil })

	if err := sampleTraining().Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	localData, err := os.ReadFile(local)
	if err != nil {
		t.Fatal(err)
	}
	backupData, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("backup not written: %v", err)
	}
	if string(localData) != string(backupData) {
		t.Error("backup contents differ from local file")
	}
	if len(notices) != 0 {
		t.Errorf("unexpected notices: %v", notices)
	}
}

func TestSave_MountFailureIsNotFatal(t *testing.T) {
	local, backup := setupStorage(t, func() error { return errors.New("drive not found") })

	if err := sampleTraining().Save(); err != nil {
		t.Fatalf("Save should succeed when backup fails, got: %v", err)
	}

	if _, err := os.Stat(local); err != nil {
		t.Errorf("local file should exist: %v", err)
	}
	if _, err := os.Stat(backup); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("backup should not exist, stat err = %v", err)
	}
	if len(notices) != 1 || !strings.Contains(notices[0], "Backup failed") {
		t.Errorf("expected a 'Backup failed' notice, got %v", notices)
	}
	if !strings.Contains(notices[0], "drive not found") {
		t.Errorf("notice should include the mount error, got %q", notices[0])
	}
}

func TestSave_LocalFailureIsFatal(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })

	// Make the parent of the local file a regular file so MkdirAll fails.
	parent := filepath.Dir(local)
	if err := os.MkdirAll(filepath.Dir(parent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(parent, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := sampleTraining().Save()
	if err == nil {
		t.Fatal("expected error when local save fails")
	}
	if !strings.Contains(err.Error(), "local save") {
		t.Errorf("unexpected error: %v", err)
	}
}

/* backup */

func TestBackup_WritesFileWhenMounted(t *testing.T) {
	_, backupPath := setupStorage(t, func() error { return nil })

	if err := backup([]byte("payload")); err != nil {
		t.Fatalf("backup: %v", err)
	}

	got, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "payload" {
		t.Errorf("content = %q, want %q", got, "payload")
	}
}

func TestBackup_WrapsMountError(t *testing.T) {
	mountErr := errors.New("drive not found")
	_, backupPath := setupStorage(t, func() error { return mountErr })

	err := backup([]byte("payload"))
	if !errors.Is(err, mountErr) {
		t.Errorf("expected error wrapping mount error, got %v", err)
	}
	if _, statErr := os.Stat(backupPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("backup file should not exist, stat err = %v", statErr)
	}
}

/* writeFileAtomic */

func TestWriteFileAtomic_CreatesDirsAndSetsPerm(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "c.json")

	if err := writeFileAtomic(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("content = %q, want %q", got, "hello")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("perm = %v, want 0600", info.Mode().Perm())
	}
}

func TestWriteFileAtomic_OverwritesAndLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")

	if err := writeFileAtomic(path, []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(path, []byte("second"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "second" {
		t.Errorf("content = %q, want %q", got, "second")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "data.json" {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("expected only data.json in dir, found %v", names)
	}
}

func TestWriteFileAtomic_FailureCleansUpTempFile(t *testing.T) {
	dir := t.TempDir()

	// Renaming a file onto a non-empty directory fails.
	target := filepath.Join(dir, "target")
	if err := os.MkdirAll(filepath.Join(target, "child"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := writeFileAtomic(target, []byte("data"), 0o600); err == nil {
		t.Fatal("expected rename onto a directory to fail")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("temp file %q was left behind", e.Name())
		}
	}
}
