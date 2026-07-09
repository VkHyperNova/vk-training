package db

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"vk-training/pkg/color"
	"vk-training/pkg/config"
	"vk-training/pkg/util"
)

type Item struct {
	NAME  string  `json:"name"`
	VALUE int `json:"value"`
	DATE  string  `json:"date"`
}

type Training struct {
	Training []Item `json:"training"`
}

func (t *Training) LoadFromFile(source string) error {

	file, err := os.Open(source)
	if err != nil {
		return err
	}

	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, t)
	if err != nil {
		return err
	}

	return nil
}

func (t *Training) PrintDashboard() {

	util.ClearScreen()

	fmt.Print(color.Blue + color.Bold + "\nVK TRAINING v1\n\n" + color.Reset)

	fmt.Print(color.Blue + "\n< history, undo, import, export, unmount, quit >" + color.Reset)
	fmt.Print("\n=> ")
}

func (t *Training) Export() error {

	input, err := util.PromptWithSuggestion("Export db to d drive? (y/n) ", "n")
	if err != nil {
		return err
	}

	if input == "y" || input == "yes" {

		if err := util.InitBackupStorage(); err != nil {
			return err
		}

		if err := t.LoadFromFile(config.LocalFile); err != nil {
			return fmt.Errorf("load from file: %w", err)
		}

		finance, err := json.MarshalIndent(t, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(config.BackupFile, finance, 0644); err != nil {
			return err
		}

		fmt.Printf("Database exported to %s\nPress Enter!", config.BackupFile)
		return nil
	}

	fmt.Println("Export canceled!")
	return nil
}

func (t *Training) Import() error {

	input, err := util.PromptWithSuggestion("Import db from d drive? (y/n) ", "n")
	if err != nil {
		return err
	}

	if input == "y" || input == "yes" {

		if err := util.InitBackupStorage(); err != nil {
			return err
		}

		if err := t.LoadFromFile(config.BackupFile); err != nil {
			return fmt.Errorf("load from file: %w", err)
		}

		finance, err := json.MarshalIndent(t, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(config.LocalFile, finance, 0644); err != nil {
			return err
		}

		fmt.Printf("Database imported from %s\nPress Enter!", config.BackupFile)
		return nil
	}

	fmt.Println("Import canceled!")

	return nil
}

func (t *Training) Add(name string, value int) error {

	name = strings.ToLower(name)

	now := time.Now()

	NewItem := Item{
		NAME:  name,
		VALUE: value,
		DATE:  now.Format("02-01-2006"),
	}

	t.Training = append(t.Training, NewItem)

	return t.save()
}

func (t *Training) Undo() error {
	if len(t.Training) == 0 {
		return fmt.Errorf("nothing to undo")
	}
	t.Training = t.Training[:len(t.Training)-1]
	return t.save()
}

func (t *Training) save() error {

	copySlice := make([]Item, len(t.Training))
	copy(copySlice, t.Training)

	copyTraining := Training{Training: copySlice}
	finance, err := json.MarshalIndent(copyTraining, "", "  ")
	if err != nil {
		return err
	}

	// Save Local
	if err := os.WriteFile(config.LocalFile, finance, 0644); err != nil {
		return err
	}
	fmt.Println(color.Green + "Local save!" + color.Reset)

	// Save Backup
	if err := util.InitBackupStorage(); err != nil {
		fmt.Println(color.Yellow + "Backup init failed: " + err.Error() + color.Reset)
		return nil // or return err, depending on your needs
	}

	if err := os.WriteFile(config.BackupFile, finance, 0644); err != nil {
		fmt.Println(color.Yellow + "Backup write failed: " + err.Error() + color.Reset)
		return nil // same decision here
	}

	fmt.Println(color.Green + "Backup save!" + color.Reset)

	return nil
}
