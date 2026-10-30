package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/VkHyperNova/vkmount"
)

/* Storage */

var mountBackup = func() error {
	return vkmount.Drive{Label: driveLabel}.Mount()
}

func (t *Training) Load(source string) error {

	byteValue, err := os.ReadFile(source)

	// No file yet: start with empty data
	if errors.Is(err, os.ErrNotExist) {
		notice("No data file found at " + source + ", starting empty")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", source, err)
	}

	if err := json.Unmarshal(byteValue, t); err != nil {
		return fmt.Errorf("error parsing JSON from file %s: %w", source, err)
	}

	return nil
}

func (t *Training) Save() error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	// Local save is the one that must succeed.
	if err := writeFileAtomic(LocalFile, data, 0o600); err != nil {
		return fmt.Errorf("local save: %w", err)
	}

	// Backup is best-effort and never fails the save.
	if err := backup(data); err != nil {
		notice("Backup failed: " + err.Error())
	}
	return nil
}

/* Helper Functions */

func backup(data []byte) error {
	if err := mountBackup(); err != nil {
		return fmt.Errorf("mount drive: %w", err)
	}
	return writeFileAtomic(BackupFile, data, 0o600)
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {

	// writeFileAtomic writes to a temp file in the same dir, then renames it,
	// so readers never see a partially written file.

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // harmless after a successful rename

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), perm); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
