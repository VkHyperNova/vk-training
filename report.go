package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
)

func (t *Training) PrintDashboard() {

	clearScreen()

	fmt.Print(Blue + Bold + "\nVK TRAINING v1\n\n" + Reset)

	t.PrintSummary()

	fmt.Println(Blue + "\n< log, undo >" + Reset)
}

func (t *Training) PrintSummary() {
	// Guard against empty slice panic
	if len(t.Training) == 0 {
		return
	}

	// Group and Sum up Items
	totals := make(map[string]float64)
	for _, item := range t.Training {
		totals[item.Name] += item.Value
	}

	// Convert to sortable slice, skipping zero-sum entries
	type itemSum struct {
		name string
		sum  float64
	}
	pairs := make([]itemSum, 0, len(totals))
	for name, sum := range totals {
		if sum != 0 {
			pairs = append(pairs, itemSum{name, sum})
		}
	}

	// Sort highest to lowest
	sort.Slice(pairs, func(a, b int) bool {
		return pairs[a].sum > pairs[b].sum
	})

	// Last added item is used for highlight
	lastItem := t.Training[len(t.Training)-1]

	for _, p := range pairs {
		line := fmt.Sprintf("\t%s: %g", p.name, p.sum)

		if p.name == lastItem.Name {
			line += fmt.Sprintf(" %s%s| +%g%s", Bold, Yellow, lastItem.Value, Reset)
		}

		fmt.Println(line)
	}
}

func clearScreen() {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}
