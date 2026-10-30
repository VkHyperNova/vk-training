package main

import (
	"errors"
	"os"
	"strconv"
	"testing"
	"time"
)

// setArgs replaces os.Args for the duration of the test.
func setArgs(t *testing.T, args ...string) {
	t.Helper()

	old := os.Args
	os.Args = args
	t.Cleanup(func() { os.Args = old })
}

// loadSaved reads the local file back, to check what was actually persisted.
func loadSaved(t *testing.T, path string) *Training {
	t.Helper()

	tr := &Training{}
	if err := tr.Load(path); err != nil {
		t.Fatalf("load saved file: %v", err)
	}
	return tr
}

/* Log */

func TestLog_AddsItemAndSaves(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })
	setArgs(t, "prog", "log", "PushUps", "30")

	tr := &Training{}
	before := time.Now()
	if err := tr.Log(); err != nil {
		t.Fatalf("Log: %v", err)
	}
	after := time.Now()

	if len(tr.Training) != 1 {
		t.Fatalf("got %d items, want 1", len(tr.Training))
	}
	item := tr.Training[0]
	if item.Name != "pushups" {
		t.Errorf("name = %q, want lowercase %q", item.Name, "pushups")
	}
	if item.Value != 30 {
		t.Errorf("value = %v, want 30", item.Value)
	}
	if item.Date.Before(before) || item.Date.After(after) {
		t.Errorf("date %v not between %v and %v", item.Date, before, after)
	}

	saved := loadSaved(t, local)
	if len(saved.Training) != 1 || saved.Training[0].Name != "pushups" || saved.Training[0].Value != 30 {
		t.Errorf("saved data = %+v", saved.Training)
	}
}

func TestLog_AcceptsDecimalValue(t *testing.T) {
	setupStorage(t, func() error { return nil })
	setArgs(t, "prog", "log", "run", "5.25")

	tr := &Training{}
	if err := tr.Log(); err != nil {
		t.Fatalf("Log: %v", err)
	}
	if got := tr.Training[0].Value; got != 5.25 {
		t.Errorf("value = %v, want 5.25", got)
	}
}

func TestLog_AppendsToExistingItems(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })
	setArgs(t, "prog", "add", "run", "2")

	tr := sampleTraining()
	if err := tr.Log(); err != nil {
		t.Fatalf("Log: %v", err)
	}

	if len(tr.Training) != 3 {
		t.Fatalf("got %d items, want 3", len(tr.Training))
	}
	if last := tr.Training[2]; last.Name != "run" || last.Value != 2 {
		t.Errorf("last item = %+v", last)
	}
	if saved := loadSaved(t, local); len(saved.Training) != 3 {
		t.Errorf("saved %d items, want 3", len(saved.Training))
	}
}

func TestLog_TooFewArgs(t *testing.T) {
	for _, args := range [][]string{
		{"prog", "log"},
		{"prog", "log", "pushups"},
	} {
		local, _ := setupStorage(t, func() error { return nil })
		setArgs(t, args...)

		tr := &Training{}
		if err := tr.Log(); err == nil {
			t.Errorf("args %v: expected error", args)
		}
		if len(tr.Training) != 0 {
			t.Errorf("args %v: item was added despite error", args)
		}
		if _, err := os.Stat(local); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("args %v: nothing should be saved, stat err = %v", args, err)
		}
	}
}

func TestLog_InvalidValue(t *testing.T) {
	for _, value := range []string{"abc", "", "12,5", "1.2.3"} {
		local, _ := setupStorage(t, func() error { return nil })
		setArgs(t, "prog", "log", "pushups", value)

		tr := &Training{}
		err := tr.Log()
		if err == nil {
			t.Errorf("value %q: expected error", value)
			continue
		}
		if !errors.Is(err, strconv.ErrSyntax) {
			t.Errorf("value %q: error should wrap strconv.ErrSyntax, got %v", value, err)
		}
		if len(tr.Training) != 0 {
			t.Errorf("value %q: item was added despite error", value)
		}
		if _, statErr := os.Stat(local); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("value %q: nothing should be saved, stat err = %v", value, statErr)
		}
	}
}

func TestLog_SaveFailureReturnsError(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })
	setArgs(t, "prog", "log", "pushups", "30")

	// Make the local file's parent a regular file so the save fails.
	parent := local[:len(local)-len("/training.json")]
	if err := os.WriteFile(parent, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := (&Training{}).Log(); err == nil {
		t.Fatal("expected error when save fails")
	}
}

/* Undo */

func TestUndo_RemovesLastItemAndSaves(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })

	tr := sampleTraining()
	if err := tr.Undo(); err != nil {
		t.Fatalf("Undo: %v", err)
	}

	if len(tr.Training) != 1 || tr.Training[0].Name != "pushups" {
		t.Errorf("after undo: %+v", tr.Training)
	}

	saved := loadSaved(t, local)
	if len(saved.Training) != 1 || saved.Training[0].Name != "pushups" {
		t.Errorf("saved after undo: %+v", saved.Training)
	}
}

func TestUndo_EmptyReturnsError(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })

	err := (&Training{}).Undo()
	if err == nil {
		t.Fatal("expected error when there is nothing to undo")
	}
	if _, statErr := os.Stat(local); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("nothing should be saved, stat err = %v", statErr)
	}
}

func TestUndo_LastRemainingItemLeavesEmptyList(t *testing.T) {
	local, _ := setupStorage(t, func() error { return nil })

	tr := &Training{Training: []Item{{Name: "run", Value: 1, Date: time.Now()}}}
	if err := tr.Undo(); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if len(tr.Training) != 0 {
		t.Errorf("expected empty list, got %+v", tr.Training)
	}
	if saved := loadSaved(t, local); len(saved.Training) != 0 {
		t.Errorf("saved after undo: %+v", saved.Training)
	}
}

func TestUndo_RepeatedUntilEmpty(t *testing.T) {
	setupStorage(t, func() error { return nil })

	tr := sampleTraining()
	for i := 0; i < 2; i++ {
		if err := tr.Undo(); err != nil {
			t.Fatalf("undo %d: %v", i+1, err)
		}
	}
	if err := tr.Undo(); err == nil {
		t.Error("third undo should fail on an empty list")
	}
}
