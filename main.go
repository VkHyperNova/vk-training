package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	Bold   = "\033[1m"
	Italic = "\033[3m"

	driveLabel      = "VK DATA"
	driveMountPoint = "/media/veikko/VK DATA"
	file            = "training.json"
	baseDB          = "TRAINING"
	baseLocal       = "DATABASES"
	// baseLocal = "DEVELOPMENT"

	usage = Bold + "Usage:" + Reset + `
  training <command> [arguments]

` + Bold + "Commands:" + Reset + `
  log, add, a <name> <value>   Log an item (value can be a decimal, e.g. 12.5)
  undo                         Remove the last logged item
  help, -h, --help             Show this help

` + Bold + "Examples:" + Reset + `
  training log pushups 30
  training add run 5.25
  training undo
`
)

var (
	LocalFile  = filepath.Join(baseLocal, baseDB, file)
	BackupFile = filepath.Join(driveMountPoint, baseLocal, baseDB, file)
)

var notices []string

func notice(msg string) { notices = append(notices, msg) }

func main() {

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			fmt.Print(usage)
			return
		}
	}

	t := &Training{}

	if err := t.Load(LocalFile); err != nil {
		log.Fatalf("load from file: %v", err)
	}

	var err error
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "log", "add", "a":
			err = t.Log()
		case "undo":
			err = t.Undo()
		default:
			fmt.Printf("%serror: unknown command: %s%s\n\n", Red, os.Args[1], Reset)
			fmt.Print(usage)
			os.Exit(1)
		}
	}

	t.Report()

	for _, n := range notices {
		fmt.Println(Yellow + n + Reset)
	}

	if err != nil {
		fmt.Printf("%serror: %v%s\n", Red, err, Reset)
		os.Exit(1)
	}
}
