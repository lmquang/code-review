package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Client represents a Git client
type Client struct{}

// NewClient creates a new Git client
func NewClient() *Client {
	return &Client{}
}

// GetDiff executes 'git diff' against the branch point and returns the output and changed files
func (c *Client) GetDiff() (string, []string, error) {
	currentBranch, err := c.ExecCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", nil, fmt.Errorf("failed to get current branch: %v", err)
	}

	parentBranch, err := c.ExecCommand("git", "rev-parse", "--abbrev-ref", "@{-1}")
	if err != nil {
		return "", nil, fmt.Errorf("failed to determine parent branch: %v", err)
	}

	branchPoint, err := c.ExecCommand("git", "merge-base", currentBranch, parentBranch)
	if err != nil {
		return "", nil, fmt.Errorf("failed to find branch point: %v", err)
	}

	fmt.Printf("Comparing %s against %s from branch point %s\n", currentBranch, parentBranch, branchPoint[:7])

	changedFiles, err := c.ExecCommand("git", "diff", "--name-only", branchPoint, currentBranch)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get list of changed files: %v", err)
	}

	diff, err := c.ExecCommand("git", "diff", branchPoint, currentBranch)
	if err != nil {
		return "", nil, fmt.Errorf("failed to execute git diff: %v", err)
	}

	return diff, strings.Split(changedFiles, "\n"), nil
}

// GetFileContentAtBranchPoint retrieves the content of a file at the branch point
func (c *Client) GetFileContentAtBranchPoint(file, branchPoint string) (string, error) {
	files := strings.Split(file, " ")
	if len(files) >= 1 {
		file = files[0]
	} else if len(files) == 0 {
		return "", fmt.Errorf("invalid file path")
	}
	// Check if the file exists at the branch point
	_, err := c.ExecCommand("git", "cat-file", "-e", fmt.Sprintf("%s:%s", branchPoint, file))
	if err != nil {
		if strings.Contains(err.Error(), "Not a valid object name") {
			return "[NEW FILE]", nil
		}
		return "", fmt.Errorf("error checking file existence: %v", err)
	}

	// File exists, get its content
	content, err := c.ExecCommand("git", "show", fmt.Sprintf("%s:%s", branchPoint, file))
	if err != nil {
		return "", fmt.Errorf("error getting file content: %v", err)
	}
	return content, nil
}

// ExecCommand is a helper function to execute git commands
func (c *Client) ExecCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, stderr.String())
	}
	return strings.TrimSpace(out.String()), nil
}
