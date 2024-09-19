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

// Format prepares the git diff output for AI model review in XML format
func (f *Formatter) Format(diff string, changedFiles []string) (string, []error) {
	fileChanges := strings.Split(diff, "diff --git")

	var formattedDiff strings.Builder
	var errors []error

	formattedDiff.WriteString("<git-diff>\n")

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

		formattedDiff.WriteString("  <file>\n")
		formattedDiff.WriteString(fmt.Sprintf("    <n>%s<n>\n", f.escapeXML(fileName)))

		branchPoint, err := f.gitClient.ExecCommand("git", "merge-base", "HEAD", "@{-1}")
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to find branch point for %s: %v", fileName, err))
			continue
		}

		originalContent, err := f.gitClient.GetFileContentAtBranchPoint(fileName, branchPoint)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get original content for %s: %v", fileName, err))
			formattedDiff.WriteString("    <original-content>Unable to retrieve</original-content>\n")
		} else {
			formattedDiff.WriteString("    <original-content>\n")
			formattedDiff.WriteString(fmt.Sprintf("      <![CDATA[%s]]>\n", originalContent))
			formattedDiff.WriteString("    </original-content>\n")
		}

		formattedDiff.WriteString("    <changes>\n")
		formattedDiff.WriteString(fmt.Sprintf("      <![CDATA[%s]]>\n", change))
		formattedDiff.WriteString("    </changes>\n")
		formattedDiff.WriteString("  </file>\n")
	}

	formattedDiff.WriteString("</git-diff>")

	return formattedDiff.String(), errors
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
