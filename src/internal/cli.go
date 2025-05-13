package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// StringPrompt asks for a string value using the label
func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, Highlight(label)+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

// YesNoPrompt asks yes/no questions using the label
func YesNoPrompt(label string, def bool) bool {
	choices := "Y/n"
	if !def {
		choices = "y/N"
	}

	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s %s ", Highlight(label), ColoredMessage(ColorCyan, "("+choices+")"))
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			return def
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
}

// IsTerminal checks if the program is running in an interactive terminal
func IsTerminal() bool {
	return term.IsTerminal(int(syscall.Stdin))
}

// ShowHelp displays available commands
func ShowHelp() {
	fmt.Println(Bold(ColoredMessage(ColorBlue, "Docker Development Environment Tool")))
	fmt.Println(ColoredMessage(ColorBlue, "=================================="))
	fmt.Println(Bold("Usage:"))
	fmt.Println("  " + ColoredMessage(ColorGreen, "[domain]") + "      - Create a new project with the given domain")
	fmt.Println("  " + ColoredMessage(ColorRed, "rm [domain]") + "   - Remove an existing project")
	fmt.Println("  " + ColoredMessage(ColorCyan, "-H, --help") + "    - Show this help message")
	fmt.Println("\n" + Bold("Examples:"))
	fmt.Println("  " + ColoredMessage(ColorGreen, "myapp.test") + "    - Create a new project with domain myapp.test")
	fmt.Println("  " + ColoredMessage(ColorRed, "rm myapp.test") + " - Remove the project with domain myapp.test")
}

// ListExistingProjects returns a list of existing project domains
func ListExistingProjects() ([]string, error) {
	var projects []string
	
	domainsDir := ProjectDirPrefix
	entries, err := os.ReadDir(domainsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return projects, nil
		}
		return nil, err
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if it's a valid project by looking for docker-compose.yml
			composePath := filepath.Join(domainsDir, entry.Name(), DockerComposeFile)
			if _, err := os.Stat(composePath); err == nil {
				projects = append(projects, entry.Name())
			}
		}
	}
	
	return projects, nil
}

// InteractiveProjectCreation guides the user through creating a new project
func InteractiveProjectCreation() error {
	if !IsTerminal() {
		return fmt.Errorf("cannot run in interactive mode: not a terminal")
	}
	
	domain := StringPrompt("Enter project domain (e.g. app.test):")
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	
	PrintDivider()
	fmt.Println(Bold("PROJECT CONFIGURATION:"))
	
	useSSL := YesNoPrompt("Do you want to enable SSL for this project?", true)
	
	if useSSL {
		fmt.Println(Info("Creating project with SSL enabled..."))
	} else {
		fmt.Println(Info("Creating project without SSL..."))
		// Note: Currently SSL is required, but we'll pass the user's preference
		// to the GenerateProject function which will handle this case
	}
	
	PrintDivider()
	fmt.Println(Bold("GENERATING PROJECT:"))
	return GenerateProject(domain, useSSL)
}

// InteractiveProjectDeletion guides the user through deleting projects
func InteractiveProjectDeletion() error {
	if !IsTerminal() {
		return fmt.Errorf("cannot run in interactive mode: not a terminal")
	}
	
	projects, err := ListExistingProjects()
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}
	
	if len(projects) == 0 {
		fmt.Println(Info("No projects found to delete."))
		return nil
	}
	
	PrintDivider()
	fmt.Println(Bold("AVAILABLE PROJECTS:"))
	for i, project := range projects {
		fmt.Printf("%s %s\n", ColoredMessage(ColorGreen, fmt.Sprintf("%d.", i+1)), Bold(project))
	}
	
	projectIndex := -1
	
	for projectIndex < 0 || projectIndex >= len(projects) {
		indexStr := StringPrompt(fmt.Sprintf("Enter project number to delete (1-%d):", len(projects)))
		var index int
		_, err := fmt.Sscanf(indexStr, "%d", &index)
		if err != nil || index < 1 || index > len(projects) {
			fmt.Println(Warning("Invalid selection. Please try again."))
			continue
		}
		projectIndex = index - 1
	}
	
	domain := projects[projectIndex]
	PrintDivider()
	fmt.Println(Bold("CONFIRMATION:"))
	confirm := YesNoPrompt(fmt.Sprintf("Are you sure you want to delete '%s'?", Bold(domain)), false)
	
	if !confirm {
		fmt.Println(Info("Operation cancelled."))
		return nil
	}
	
	DeleteProject(domain)
	return nil
}

// WaitForKeyPress waits for the user to press any key
func WaitForKeyPress(message string) {
	fmt.Println(Info(message))
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}

// RunInteractiveMode runs the tool in interactive mode, allowing multiple actions until exit
func RunInteractiveMode() error {
	if !IsTerminal() {
		return fmt.Errorf("cannot run in interactive mode: not a terminal")
	}
	
	for {
		PrintSectionDivider("DOCKER DEVELOPMENT ENVIRONMENT TOOL")
		
		projects, err := ListExistingProjects()
		if err != nil {
			return fmt.Errorf("failed to list projects: %w", err)
		}
		
		// Display existing projects
		if len(projects) == 0 {
			fmt.Println(Info("No existing projects found."))
		} else {
			fmt.Println(Bold("Existing projects:"))
			for _, project := range projects {
				fmt.Println("  -", Bold(project))
			}
		}
		
		// Show menu
		PrintDivider()
		fmt.Println(Bold("AVAILABLE ACTIONS:"))
		fmt.Println(ColoredMessage(ColorGreen, "1. Create a new project"))
		fmt.Println(ColoredMessage(ColorRed, "2. Delete an existing project"))
		fmt.Println(Gray("3. Exit"))
		
		option := StringPrompt("Enter your choice (1-3):")
		
		switch option {
		case "1":
			// Create a new project
			PrintSectionDivider("CREATE NEW PROJECT")
			err := InteractiveProjectCreation()
			if err != nil {
				fmt.Printf(Error("Error creating project: %v\n"), err)
				WaitForKeyPress("Press Enter to continue...")
			} else {
				WaitForKeyPress("Project created successfully. Press Enter to continue...")
			}
			
		case "2":
			// Delete an existing project
			PrintSectionDivider("DELETE EXISTING PROJECT")
			if len(projects) == 0 {
				fmt.Println(Info("No projects available to delete."))
				WaitForKeyPress("Press Enter to continue...")
			} else {
				err := InteractiveProjectDeletion()
				if err != nil {
					fmt.Printf(Error("Error deleting project: %v\n"), err)
					WaitForKeyPress("Press Enter to continue...")
				} else {
					WaitForKeyPress("Press Enter to continue...")
				}
			}
			
		case "3":
			// Exit
			PrintSectionDivider("EXITING APPLICATION")
			fmt.Println(Gray("Exiting. Goodbye!"))
			return nil
			
		default:
			fmt.Println(Warning("Invalid option. Please choose 1, 2, or 3."))
			WaitForKeyPress("Press Enter to continue...")
		}
	}
} 