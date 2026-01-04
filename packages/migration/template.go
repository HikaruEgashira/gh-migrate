package migration

import (
	"fmt"
	"os"
)

// PRTemplateResult contains the template content and its source path
type PRTemplateResult struct {
	Content string
	Path    string
}

// GetPRTemplate returns the PR template content and its source path.
// Priority: --template flag > .github/PULL_REQUEST_TEMPLATE.md > default
func GetPRTemplate(templateFlag string, workPath string) (*PRTemplateResult, error) {
	// 1. --templateオプションが指定されている場合
	if templateFlag != "" {
		content, err := os.ReadFile(templateFlag)
		if err != nil {
			return nil, fmt.Errorf("failed to read template file: %w", err)
		}
		return &PRTemplateResult{Content: string(content), Path: templateFlag}, nil
	}

	// 2. クローン済みのワークディレクトリからテンプレートを探す
	templatePaths := []string{
		".github/PULL_REQUEST_TEMPLATE.md",
		".github/pull_request_template.md",
		"PULL_REQUEST_TEMPLATE.md",
		"pull_request_template.md",
	}

	for _, relPath := range templatePaths {
		fullPath := workPath + "/" + relPath
		if content, err := os.ReadFile(fullPath); err == nil {
			return &PRTemplateResult{Content: string(content), Path: relPath}, nil
		}
	}

	// 3. テンプレートがない場合はnilを返す
	return nil, nil
}
