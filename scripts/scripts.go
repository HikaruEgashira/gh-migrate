package scripts

import (
	"log"
	"os"
	"os/exec"
)

func ExecCommand(cmdOption string, titleTemplate *string, bodyTemplate *string) error {
	*titleTemplate = *titleTemplate + " run " + cmdOption
	*bodyTemplate = *bodyTemplate + "\n" + cmdOption

	runOutput, err := exec.Command("sh", "-c", cmdOption).CombinedOutput()
	if err != nil {
		return err
	}
	log.Printf("INFO: %s", string(runOutput))
	return nil
}

func ExecScript(scriptOption string, titleTemplate *string, bodyTemplate *string, currentPath string, scriptType string) error {
	scriptContent, err := os.ReadFile(currentPath + "/" + scriptOption)
	if err != nil {
		return err
	}

	*titleTemplate = *titleTemplate + " run " + scriptType + " " + scriptOption
	*bodyTemplate = *bodyTemplate + "\n" + "```" + scriptType + "\n" + string(scriptContent) + "\n```"

	var runOutput []byte
	switch scriptType {
	case "sh":
		runOutput, err = exec.Command("sh", currentPath+"/"+scriptOption).CombinedOutput()
	case "astgrep":
		runOutput, err = exec.Command("sg", "scan", "-r", currentPath+"/"+scriptOption, "--no-ignore", "hidden", "-U").CombinedOutput()
	case "semgrep":
		runOutput, err = exec.Command("semgrep", "--config", currentPath+"/"+scriptOption).CombinedOutput()
	}

	if err != nil {
		return err
	}
	log.Printf("INFO: %s", string(runOutput))
	return nil
}
