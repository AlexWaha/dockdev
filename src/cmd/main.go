package main

import (
	"fmt"
	"os"
	"generator/internal"
)

func main() {
	args := os.Args

	// Process command-line arguments
	if len(args) > 1 {
		// Check for help command
		if args[1] == "-H" || args[1] == "--help" {
			internal.PrintSectionDivider("HELP")
			internal.ShowHelp()
			return
		}

		// Check for delete command
		if args[1] == "rm" {
			if len(args) == 3 {
				internal.DeleteProject(args[2])
				
				// After deletion, if in interactive terminal mode, ask if user wants to continue
				if internal.IsTerminal() {
					internal.PrintDivider()
					if internal.YesNoPrompt("Would you like to perform additional actions?", true) {
						if err := internal.RunInteractiveMode(); err != nil {
							fmt.Println(internal.Error("Error:"), err)
							os.Exit(1)
						}
						return
					}
				}
			} else {
				// Interactive delete mode
				internal.PrintSectionDivider("INTERACTIVE DELETE MODE")
				if err := internal.InteractiveProjectDeletion(); err != nil {
					fmt.Println(internal.Error("Error:"), err)
					os.Exit(1)
				}
				
				// After deletion, if in interactive terminal mode, ask if user wants to continue
				if internal.IsTerminal() {
					internal.PrintDivider()
					if internal.YesNoPrompt("Would you like to perform additional actions?", true) {
						if err := internal.RunInteractiveMode(); err != nil {
							fmt.Println(internal.Error("Error:"), err)
							os.Exit(1)
						}
						return
					}
				}
			}
			return
		}

		// Assume the argument is a domain name
		internal.PrintSectionDivider("CREATING PROJECT: " + args[1])
		if err := internal.GenerateProject(args[1]); err != nil {
			fmt.Println(internal.Error("Error:"), err)
			os.Exit(1)
		}
		
		// After creation, if in interactive terminal mode, ask if user wants to continue
		if internal.IsTerminal() {
			internal.PrintDivider()
			if internal.YesNoPrompt("Would you like to perform additional actions?", true) {
				if err := internal.RunInteractiveMode(); err != nil {
					fmt.Println(internal.Error("Error:"), err)
					os.Exit(1)
				}
				return
			}
		}
		
		return
	}

	// If no arguments are provided, run in fully interactive mode
	if internal.IsTerminal() {
		fmt.Println(internal.Bold(internal.Info("Starting interactive mode. You can create or delete projects until you choose to exit.")))
		if err := internal.RunInteractiveMode(); err != nil {
			fmt.Println(internal.Error("Error:"), err)
			os.Exit(1)
		}
	} else {
		// If not in a terminal but no arguments provided, show help
		internal.PrintSectionDivider("HELP")
		internal.ShowHelp()
		fmt.Println("\n" + internal.Error("Error: No domain specified. Please provide a domain name or run in an interactive terminal."))
		os.Exit(1)
	}
}