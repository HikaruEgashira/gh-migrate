package scripts

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Exec executes a command or script file.
// If the argument is an existing file path, it executes as a shell script.
// Otherwise, it executes as a shell command.
func Exec(cmdOption string, titleTemplate *string, bodyTemplate *string, currentPath string) error {
	// Check if the argument is a file path
	var scriptPath string
	if filepath.IsAbs(cmdOption) {
		scriptPath = cmdOption
	} else {
		scriptPath = filepath.Join(currentPath, cmdOption)
	}

	if _, err := os.Stat(scriptPath); err == nil {
		// File exists, execute as script
		return execScript(scriptPath, cmdOption, titleTemplate, bodyTemplate)
	}

	// Not a file, execute as command
	return execCommand(cmdOption, titleTemplate, bodyTemplate)
}

func execCommand(cmdOption string, titleTemplate *string, bodyTemplate *string) error {
	*titleTemplate = *titleTemplate + " run " + cmdOption
	*bodyTemplate = *bodyTemplate + "\n" + cmdOption

	runOutput, err := exec.Command("sh", "-c", cmdOption).CombinedOutput()
	if err != nil {
		return err
	}
	log.Printf("INFO: %s", string(runOutput))
	return nil
}

func execScript(scriptPath string, scriptName string, titleTemplate *string, bodyTemplate *string) error {
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	*titleTemplate = *titleTemplate + " run " + scriptName
	*bodyTemplate = *bodyTemplate + "\n" + "```sh\n" + string(scriptContent) + "\n```"

	runOutput, err := exec.Command("sh", scriptPath).CombinedOutput()
	if err != nil {
		return err
	}
	log.Printf("INFO: %s", string(runOutput))
	return nil
}
