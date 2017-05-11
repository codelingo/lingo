package git

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/juju/errors"
)

// This file is responsible for building up a set of queries from a git repository.
// That includes reading reading .lingo files, determining which ones are applicable based on directory
// from which the review is started (cascade logic), and filling out a CLQL template.
// TODO(BlakeMScurr) pull out generic query building logic. Define

// BuildQueries finds all the relevant .lingo files from the repo and builds out the vcs section tenet to
// produce a valid query for the platform.
func (r *Repo) BuildQueries() ([]string, error) {
	workingDir, err := r.WorkingDir()
	if err != nil {
		return []string{}, errors.Trace(err)
	}

	lingoFiles, err := r.getDotlingoFiles(workingDir)
	if err != nil {
		return nil, errors.Trace(err)
	}

	queries := []string{}

	for dir, lingoSrc := range lingoFiles {
		q, err := r.buildQuery(dir, lingoSrc)
		if err != nil {
			return nil, errors.Trace(err)
		}
		queries = append(queries, q)
	}

	return queries, nil
}

func (r *Repo) buildQuery(dir, lingoSrc string) (string, error) {
	// Create dir nodes
	dirs := []string{}

	for dir != "." {
		dir = filepath.Dir(dir)
		dirs = append([]string{filepath.Base(dir)}, dirs...)
	}

	vcsFacts, err := r.populateGitTemplate()
	if err != nil {
		return "", errors.Trace(err)
	}

	patches, err := r.Patches()
	if err != nil {
		return "", errors.Trace(err)
	}

	whitespace := "    "
	patchString := strings.Join(patches, "\n")
	if patchString != "" {
		vcsFacts += "\n        git.patch:\n          diff: ```" + strings.Replace(patchString, "\n", "\\n", -1) + "```"
		whitespace += "  "
	}

	// Get tenets. Here a tenet is **very loosely** defined as any sequence of lines starting with four spaces.
	matchStmts := regexp.MustCompile("(    .*\n)+")

	// Update the query string
	newLingoSrc := lingoSrc
	for _, match := range matchStmts.FindAllString(lingoSrc, -1) {
		// Add VCS facts
		newMatch := vcsFacts + "\n"

		// Limit scope of query to location of the .lingo file

		for _, dirFactName := range dirs {
			newMatch += whitespace + "    common.dir:\n"
			newMatch += whitespace + "      name: \"" + dirFactName + "\"\n"
			whitespace += "  "
		}

		// Add indentation to original string
		lines := []string{}
		for _, line := range strings.Split(match, "\n") {
			line = whitespace + line
			lines = append(lines, line)
		}

		newMatch += strings.Join(lines, "\n")

		newLingoSrc = strings.Replace(newLingoSrc, match, newMatch, -1)
	}

	return newLingoSrc, nil
}

const gitTemplate = `    git.repo:
      name: %s
      host: %s
      owner: %s
      git.commit:
        sha: "%s"`

func (r *Repo) populateGitTemplate() (string, error) {
	owner, repoName, err := r.OwnerAndNameFromRemote()
	if err != nil {
		return "", errors.Annotate(err, "\nlocal vcs error")
	}

	sha, err := r.CurrentCommitId()
	if err != nil {
		return "", errors.Trace(err)
	}

	return fmt.Sprintf(gitTemplate, repoName, "local", owner, sha), nil
}

// Get all .lingo files that apply to the current review.
func (r *Repo) getDotlingoFiles(workingDir string) (map[string]string, error) {
	sha, err := r.CurrentCommitId()
	if err != nil {
		return nil, errors.Trace(err)
	}

	paths, err := r.getDotlingoFilepaths(sha, workingDir)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// Only those paths that are either a child or parent of the working directory generate queries.
	// Go to the root of the repository to do a full test.
	lingoFiles := map[string]string{}
	for _, path := range paths {
		pathDir := filepath.Dir(path)
		if pathDir == "." {
			pathDir = ""
		}

		if strings.HasPrefix(pathDir, workingDir) || strings.HasPrefix(workingDir, pathDir) {
			lingoSrc, err := r.ReadFile(path, sha)
			if err != nil {
				return nil, errors.Trace(err)
			}
			lingoFiles[path] = lingoSrc
		}
	}
	return lingoFiles, nil
}

// gets the file paths of all the .lingo files in the repo
func (r *Repo) getDotlingoFilepaths(commitID string, directory string) ([]string, error) {
	out, err := gitCMD("ls-tree", "-r", "--name-only", "--full-tree", commitID)
	if err != nil {
		return []string{}, errors.Trace(err)
	}

	files := strings.Split(out, "\n")

	dotlingoFilepaths := []string{}
	for _, filepath := range files {
		if strings.HasSuffix(filepath, ".lingo") {
			dotlingoFilepaths = append(dotlingoFilepaths, filepath)
		}
	}

	return dotlingoFilepaths, nil
}
