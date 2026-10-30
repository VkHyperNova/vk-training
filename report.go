package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (t *Training) Report() {

	clearScreen()

	fmt.Print(Blue + Bold + "\nVK TRAINING v1\n\n" + Reset)

	t.printSummary()

	fmt.Println(Blue + "\n< log, undo >" + Reset)
}

const barWidth = 20

type itemSum struct {
	name string
	sum  float64
}

func summarize(items []Item) []itemSum {
	totals := make(map[string]float64)
	for _, item := range items {
		totals[item.Name] += item.Value
	}

	pairs := make([]itemSum, 0, len(totals))
	for name, sum := range totals {
		if sum != 0 {
			pairs = append(pairs, itemSum{name, sum})
		}
	}

	sort.Slice(pairs, func(a, b int) bool {
		if pairs[a].sum != pairs[b].sum {
			return pairs[a].sum > pairs[b].sum
		}
		return pairs[a].name < pairs[b].name // stable order for ties
	})
	return pairs
}

func (t *Training) printSummary() {
	pairs := summarize(t.Training)
	if len(pairs) == 0 {
		fmt.Println(Gray + "\tNo training logged yet" + Reset)
		return
	}

	last := t.Training[len(t.Training)-1]
	maxSum := pairs[0].sum // sorted descending, so first is the largest

	// Column widths
	nameW, valW := 0, 0
	for _, p := range pairs {
		nameW = max(nameW, utf8.RuneCountInString(p.name))
		valW = max(valW, len(formatValue(p.sum)))
	}

	for _, p := range pairs {
		line := fmt.Sprintf("\t%-*s  %*s  %s%-*s%s",
			nameW, p.name,
			valW, formatValue(p.sum),
			Cyan, barWidth, bar(p.sum, maxSum), Reset,
		)

		if p.name == last.Name {
			line += fmt.Sprintf("  %s%s+%s%s", Bold, Yellow, formatValue(last.Value), Reset)
		}

		fmt.Println(line)
	}
}

func bar(sum, maxSum float64) string {
	if sum <= 0 || maxSum <= 0 {
		return ""
	}
	n := int(math.Round(sum / maxSum * barWidth))
	return strings.Repeat("█", max(n, 1))
}

func formatValue(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
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
