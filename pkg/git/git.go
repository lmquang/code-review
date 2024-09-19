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
func NewClient() IGit {
	return &Client{}
}

// GetDiff executes 'git diff' against the branch point and returns the output and changed files
func (c *Client) GetDiff() (string, []string, error) {
	currentBranch, err := c.ExecCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", nil, fmt.Errorf("failed to get current branch: %v", err)
	}

	// Get the branch that the current branch was checked out from
	baseBranch, err := c.ExecCommand("git", "rev-parse", "--abbrev-ref", "@{u}")
	if err != nil {
		// If there's an error (e.g., no upstream branch), fallback to 'develop'
		baseBranch = "develop"
	} else {
		// The result will be in the format 'origin/branch', so we need to remove 'origin/'
		baseBranch = strings.TrimPrefix(baseBranch, "origin/")
	}

	fmt.Printf("Comparing %s against %s\n", currentBranch, baseBranch)

	// Find the merge-base (common ancestor) of the current branch and the base branch
	mergeBase, err := c.ExecCommand("git", "merge-base", currentBranch, baseBranch)
	if err != nil {
		return "", nil, fmt.Errorf("failed to find merge base: %v", err)
	}

	changedFiles, err := c.ExecCommand("git", "diff", "--name-only", mergeBase, currentBranch)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get list of changed files: %v", err)
	}

	diff, err := c.ExecCommand("git", "diff", mergeBase, currentBranch)
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
