package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Item struct {
	Name  string    `json:"name"`
	Value float64   `json:"value"`
	Date  time.Time `json:"date"`
}

type Training struct {
	Training []Item `json:"training"`
}

func (t *Training) Log() error {

	if len(os.Args) < 4 {
		return fmt.Errorf("usage: log <name> <value>")
	}

	name := strings.ToLower(os.Args[2])

	value, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		return fmt.Errorf("invalid value %q: %w", os.Args[3], err)
	}

	newItem := Item{
		Name:  name,
		Value: value,
		Date:  time.Now(),
	}

	t.Training = append(t.Training, newItem)

	return t.Save()
}

func (t *Training) Undo() error {
	if len(t.Training) == 0 {
		return fmt.Errorf("nothing to undo")
	}
	t.Training = t.Training[:len(t.Training)-1]
	return t.Save()
}
