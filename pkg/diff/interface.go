package diff

type IDiff interface {
	Format(diff string, changedFiles []string) (string, []error)
}
