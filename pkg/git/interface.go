package git

type IGit interface {
	GetDiff() (string, []string, error)
	GetFileContentAtBranchPoint(file, branchPoint string) (string, error)
	ExecCommand(name string, args ...string) (string, error)
}
