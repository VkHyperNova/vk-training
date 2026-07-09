package cmd

import (
	"fmt"
	"vk-training/pkg/db"
	"vk-training/pkg/util"
)

func Start(t *db.Training) {

	for {

		t.PrintDashboard()

		var name string
		var value int

		fmt.Scanln(&name, &value)

		switch name {
		case "import", "i":
			if err := t.Import(); err != nil {
				fmt.Println(err)
			}
			util.PressAnyKey()
		case "export", "e":
			if err := t.Export(); err != nil {
				fmt.Println(err)
			}
			util.PressAnyKey()
		case "undo":
			t.Undo()
		case "unmount":
			if err := util.UnmountDrive(); err != nil {
				fmt.Println(err)
			}
			util.PressAnyKey()
		case "quit", "q":
			return
		default:
			if name == "" || value == 0.0 {
				fmt.Println("Type: Name Value")
				util.PressAnyKey()
				continue
			}
			t.Add(name, value)
		}
	}
}
