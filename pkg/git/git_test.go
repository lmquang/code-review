package git

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Error("NewClient() returned nil")
	}
	if _, ok := client.(*Client); !ok {
		t.Error("NewClient() did not return a *Client")
	}
}

// MockClient is a mock implementation of IGit for testing
type MockClient struct {
	ExecCommandFunc func(name string, args ...string) (string, error)
}

func (m *MockClient) GetDiff() (string, []string, error) {
	currentBranch, err := m.ExecCommandFunc("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", nil, err
	}

	upstream, err := m.ExecCommandFunc("git", "rev-parse", "--abbrev-ref", "@{u}")
	if err != nil {
		upstream = "develop" // Fallback to develop
	}

	mergeBase, err := m.ExecCommandFunc("git", "merge-base", currentBranch, upstream)
	if err != nil {
		return "", nil, err
	}

	changedFiles, err := m.ExecCommandFunc("git", "diff", "--name-only", mergeBase, currentBranch)
	if err != nil {
		return "", nil, err
	}

	diff, err := m.ExecCommandFunc("git", "diff", mergeBase, currentBranch)
	if err != nil {
		return "", nil, err
	}

	return diff, strings.Split(changedFiles, "\n"), nil
}

func (m *MockClient) GetFileContentAtBranchPoint(file, branchPoint string) (string, error) {
	_, err := m.ExecCommandFunc("git", "cat-file", "-e", branchPoint+":"+file)
	if err != nil {
		if strings.Contains(err.Error(), "Not a valid object name") {
			return "[NEW FILE]", nil
		}
		return "", err
	}
	return m.ExecCommandFunc("git", "show", branchPoint+":"+file)
}

func (m *MockClient) ExecCommand(name string, args ...string) (string, error) {
	return m.ExecCommandFunc(name, args...)
}

func TestGetDiff(t *testing.T) {
	tests := []struct {
		name           string
		execCommandMap map[string]struct {
			output string
			err    error
		}
		wantDiff         string
		wantChangedFiles []string
		wantErr          bool
	}{
		{
			name: "Successful diff",
			execCommandMap: map[string]struct {
				output string
				err    error
			}{
				"git rev-parse --abbrev-ref HEAD":            {output: "feature-branch", err: nil},
				"git rev-parse --abbrev-ref @{u}":            {output: "origin/main", err: nil},
				"git merge-base feature-branch origin/main":  {output: "abc123", err: nil},
				"git diff --name-only abc123 feature-branch": {output: "file1.go\nfile2.go", err: nil},
				"git diff abc123 feature-branch":             {output: "diff content", err: nil},
			},
			wantDiff:         "diff content",
			wantChangedFiles: []string{"file1.go", "file2.go"},
			wantErr:          false,
		},
		{
			name: "No upstream branch, fallback to develop",
			execCommandMap: map[string]struct {
				output string
				err    error
			}{
				"git rev-parse --abbrev-ref HEAD":            {output: "feature-branch", err: nil},
				"git rev-parse --abbrev-ref @{u}":            {output: "", err: errors.New("no upstream branch")},
				"git merge-base feature-branch develop":      {output: "abc123", err: nil},
				"git diff --name-only abc123 feature-branch": {output: "file1.go\nfile2.go", err: nil},
				"git diff abc123 feature-branch":             {output: "diff content", err: nil},
			},
			wantDiff:         "diff content",
			wantChangedFiles: []string{"file1.go", "file2.go"},
			wantErr:          false,
		},
		{
			name: "Error getting current branch",
			execCommandMap: map[string]struct {
				output string
				err    error
			}{
				"git rev-parse --abbrev-ref HEAD": {output: "", err: errors.New("failed to get current branch")},
			},
			wantDiff:         "",
			wantChangedFiles: nil,
			wantErr:          true,
		},
		{
			name: "Error getting changed files",
			execCommandMap: map[string]struct {
				output string
				err    error
			}{
				"git rev-parse --abbrev-ref HEAD":            {output: "feature-branch", err: nil},
				"git rev-parse --abbrev-ref @{u}":            {output: "origin/main", err: nil},
				"git merge-base feature-branch origin/main":  {output: "abc123", err: nil},
				"git diff --name-only abc123 feature-branch": {output: "", err: errors.New("failed to get list of changed files")},
			},
			wantDiff:         "",
			wantChangedFiles: nil,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{
				ExecCommandFunc: func(name string, args ...string) (string, error) {
					key := strings.Join(append([]string{name}, args...), " ")
					if result, ok := tt.execCommandMap[key]; ok {
						return result.output, result.err
					}
					return "", errors.New("unexpected command: " + key)
				},
			}

			diff, changedFiles, err := mockClient.GetDiff()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff != tt.wantDiff {
				t.Errorf("GetDiff() diff = %v, want %v", diff, tt.wantDiff)
			}
			if !reflect.DeepEqual(changedFiles, tt.wantChangedFiles) {
				t.Errorf("GetDiff() changedFiles = %v, want %v", changedFiles, tt.wantChangedFiles)
			}
		})
	}
}

func TestGetFileContentAtBranchPoint(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		branchPoint string
		mockOutputs map[string]struct {
			output string
			err    error
		}
		want    string
		wantErr bool
	}{
		{
			name:        "Existing file",
			file:        "test.go",
			branchPoint: "main",
			mockOutputs: map[string]struct {
				output string
				err    error
			}{
				"git cat-file -e main:test.go": {"", nil},
				"git show main:test.go":        {"file content", nil},
			},
			want:    "file content",
			wantErr: false,
		},
		{
			name:        "New file",
			file:        "new.go",
			branchPoint: "main",
			mockOutputs: map[string]struct {
				output string
				err    error
			}{
				"git cat-file -e main:new.go": {"", errors.New("Not a valid object name")},
			},
			want:    "[NEW FILE]",
			wantErr: false,
		},
		{
			name:        "Error checking file existence",
			file:        "error.go",
			branchPoint: "main",
			mockOutputs: map[string]struct {
				output string
				err    error
			}{
				"git cat-file -e main:error.go": {"", errors.New("unexpected error")},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:        "File path with spaces",
			file:        "file with spaces.go",
			branchPoint: "main",
			mockOutputs: map[string]struct {
				output string
				err    error
			}{
				"git cat-file -e main:file with spaces.go": {"", nil},
				"git show main:file with spaces.go":        {"file content with spaces", nil},
			},
			want:    "file content with spaces",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{
				ExecCommandFunc: func(name string, args ...string) (string, error) {
					key := strings.Join(append([]string{name}, args...), " ")
					if result, ok := tt.mockOutputs[key]; ok {
						return result.output, result.err
					}
					return "", errors.New("unexpected command: " + key)
				},
			}

			got, err := mockClient.GetFileContentAtBranchPoint(tt.file, tt.branchPoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileContentAtBranchPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFileContentAtBranchPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecCommand(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name    string
		command string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "Echo command",
			command: "echo",
			args:    []string{"test"},
			want:    "test",
			wantErr: false,
		},
		{
			name:    "Invalid command",
			command: "invalid_command",
			args:    []string{},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.ExecCommand(tt.command, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExecCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
