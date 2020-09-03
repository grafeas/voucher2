package grafeasos

import (
	"errors"
)

var errNoOccurrences = errors.New("no occurrences returned for image")
var errDiscoveriesUnfinished = errors.New("discoveries have not finished processing")

func projectPath(project string) string {
	return "projects/" + project
}
