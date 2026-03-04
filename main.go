package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thatisuday/commando"

	components "github.com/palianskas/toire/components"
)

const APP_NAME = "toire"

func main() {
	commando.
		SetExecutableName("toire").
		SetDescription("Toire helps you manage your local media repertoire")

	commando.
		Register(nil).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			prog := tea.NewProgram(components.InitState(), tea.WithAltScreen())

			_, err := prog.Run()

			if err != nil {
				log.Fatal(err)
			}
		})

	commando.
		Register("add").
		SetShortDescription("add a new series|season|episode").
		SetDescription("Add a series|season|episode to collection from the specified path").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			fmt.Println("'add' called with:")
			fmt.Println("    args: ")
			for k, v := range args {
				fmt.Printf("    key: %v, value: %v\n", k, v)
			}

			fmt.Println("    flags: ")
			for k, v := range flags {
				fmt.Printf("    key: %v, value: %v\n", k, v)
			}
		})

	commando.Parse(nil)
}

func HandleFatalErr(err error) {
}
