package diff

// FileChange represents a single file's change in a diff.
type FileChange struct {
	Status     ChangeStatus
	OldPath    string
	NewPath    string
	OldContent string
	NewContent string
}

// ChangeStatus indicates the type of change.
type ChangeStatus int

const (
	Modified ChangeStatus = iota
	Added
	Deleted
	Renamed
)

func (s ChangeStatus) String() string {
	switch s {
	case Modified:
		return "modified"
	case Added:
		return "added"
	case Deleted:
		return "deleted"
	case Renamed:
		return "renamed"
	default:
		return "unknown"
	}
}

// Diff represents a complete diff with metadata.
type Diff struct {
	Files       []FileChange
	TotalLines  int
	Description string
}

// TotalChangedLines returns the sum of added and removed lines across all files.
func (d Diff) TotalChangedLines() int {
	return d.TotalLines
}

// FileCount returns the number of changed files.
func (d Diff) FileCount() int {
	return len(d.Files)
}

// IsEmpty returns true if the diff has no changes.
func (d Diff) IsEmpty() bool {
	return len(d.Files) == 0
}
