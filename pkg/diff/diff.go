package diff

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lmquang/code-review/pkg/git"
)

// Formatter represents a diff formatter
type Formatter struct {
	ignoredPatterns []string
	gitClient       git.IGit
}

// NewFormatter creates a new diff formatter
func NewFormatter(ignoredPatterns []string) IDiff {
	return &Formatter{
		ignoredPatterns: ignoredPatterns,
		gitClient:       git.NewClient(),
	}
}

// Format prepares the git diff output for AI model review, separating original content and diff content
func (f *Formatter) Format(diff string, changedFiles []string) (string, string, []error) {
	fileChanges := strings.Split(diff, "diff --git")

	var originalContent strings.Builder
	var diffContent strings.Builder
	var errors []error

	originalContent.WriteString("<original-content>\n")
	diffContent.WriteString("<git-diff>\n")

	for _, change := range fileChanges {
		if change == "" {
			continue
		}

		change = "diff --git" + change

		fileNameStart := strings.Index(change, "a/")
		fileNameEnd := strings.Index(change, "\n")
		fileName := change[fileNameStart+2 : fileNameEnd]
		fileName = f.cleanFilePath(fileName)

		if f.shouldIgnoreFile(fileName) {
			continue
		}

		originalContent.WriteString(fmt.Sprintf("  <file path=\"%s\">\n", f.escapeXML(fileName)))
		diffContent.WriteString("  <file>\n")
		diffContent.WriteString(fmt.Sprintf("    <name>%s</name>\n", f.escapeXML(fileName)))

		branchPoint, err := f.gitClient.ExecCommand("git", "merge-base", "HEAD", "@{-1}")
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to find branch point for %s: %v", fileName, err))
			continue
		}

		fileContent, err := f.gitClient.GetFileContentAtBranchPoint(fileName, branchPoint)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get original content for %s: %v", fileName, err))
			originalContent.WriteString("    Unable to retrieve original content\n")
		} else {
			originalContent.WriteString(fmt.Sprintf("    <![CDATA[%s]]>\n", fileContent))
		}

		diffContent.WriteString("    <changes>\n")
		diffContent.WriteString(fmt.Sprintf("      <![CDATA[%s]]>\n", change))
		diffContent.WriteString("    </changes>\n")

		originalContent.WriteString("  </file>\n")
		diffContent.WriteString("  </file>\n")
	}

	originalContent.WriteString("</original-content>")
	diffContent.WriteString("</git-diff>")

	return originalContent.String(), diffContent.String(), errors
}

// cleanFilePath removes the 'a/' and 'b/' prefixes from the file path
func (f *Formatter) cleanFilePath(path string) string {
	path = strings.TrimPrefix(path, "a/")
	path = strings.TrimPrefix(path, "b/")
	return path
}

// shouldIgnoreFile checks if a file should be ignored based on the ignore patterns
func (f *Formatter) shouldIgnoreFile(fileName string) bool {
	for _, pattern := range f.ignoredPatterns {
		matched, err := filepath.Match(pattern, fileName)
		if err == nil && matched {
			return true
		}
		// If the pattern doesn't contain a separator, also check against the base name
		if !strings.Contains(pattern, string(filepath.Separator)) {
			matched, err = filepath.Match(pattern, filepath.Base(fileName))
			if err == nil && matched {
				return true
			}
		}
	}
	return false
}

// escapeXML escapes special characters for XML
func (f *Formatter) escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// SplitAndTrimPatterns splits a comma-separated string and trims each resulting string
func SplitAndTrimPatterns(s string) []string {
	patterns := strings.Split(s, ",")
	for i, pattern := range patterns {
		patterns[i] = strings.TrimSpace(pattern)
	}
	return patterns
}
