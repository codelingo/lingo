package git

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/codelingo/platform/controller/dotlingo"
	"github.com/codelingo/platform/controller/dotlingo/clql/ast"
	"github.com/juju/errors"
)

// This file is responsible for building up a set of queries from a git repository.
// That includes reading reading .lingo files, determining which ones are applicable based on directory
// from which the review is started (cascade logic), and filling out a CLQL template.
// TODO(BlakeMScurr) pull out generic query building logic. Define

// BuildQueries finds all the relevant .lingo files from the repo and builds out the vcs section tenet to
// produce a valid query for the platform.
func (r *Repo) BuildQueries(host string) ([]string, error) {
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
		q, err := r.buildQuery(host, dir, lingoSrc)
		if err != nil {
			return nil, errors.Trace(err)
		}
		queries = append(queries, q)
	}

	return queries, nil
}

// buildQuery is responsible for augmenting tenets and building workable queries:
//  - Scope queries to the dir of the .lingo file they are found in
//  - add basic vcs facts to the base of the match statement
// New fact that introduces a new lexicon should add a corresponding import. New facts that do *not* introduce
// a new lexicon should use that lexicon's short identifier (if it exists)..
func (r *Repo) buildQuery(host, dir, lingoSrc string) (string, error) {
	dl, err := dotlingo.Parse(lingoSrc)
	if err != nil {
		return "", errors.Trace(err)
	}

	var common, git *dotlingo.Lexicon
	common, dl.Lexicons = getLex(dl.Lexicons, "codelingo", "common")
	git, dl.Lexicons = getLex(dl.Lexicons, "codelingo", "git")

	for i, tenet := range dl.Tenets {
		// Exclude root fact
		userQuery := tenet.Match.Arg.(ast.FactList)[0].DeepCopy()

		// Nest the existing facts from the .lingo file inside new sets of facts generated
		// from the state of the repository.
		dirFacts, err := r.addDirFacts(userQuery, common, dir)
		if err != nil {
			return "", errors.Trace(err)
		}

		patchFacts, err := r.addPatchFacts(dirFacts, git)
		if err != nil {
			return "", errors.Trace(err)
		}

		fullTree, err := r.addRepoFacts(patchFacts, git, host)
		if err != nil {
			return "", errors.Trace(err)
		}

		dl.Tenets[i].Match = &ast.Fact{
			Name:      "root",
			Namespace: "!",
			LexiconID: "!",
			Arg: ast.FactList{
				fullTree,
			},
		}
	}
	str, err := dl.String()
	return str, errors.Trace(err)
}

// getLex returns a lexicon of a given owner and name, with the correct ident.
func getLex(lexicons []*dotlingo.Lexicon, owner, name string) (*dotlingo.Lexicon, []*dotlingo.Lexicon) {
	for _, lex := range lexicons {
		if lex.Owner == owner && lex.Name == name {
			return lex, lexicons
		}
	}
	newLex := &dotlingo.Lexicon{
		Owner: owner,
		Name:  name,
		Ident: name,
		// TODO(BlakeMScurr) default to latest
		Version: "0.0.0",
	}
	return newLex, append(lexicons, newLex)
}

// addDirFacts nests a set of facts from the tenet inside tenets facts based on the directory
// where the .lingo file was found. This has the effect of scoping the query to that directory.
// The commonLex exicon ensures that we namespace the facts according to the user's short idents.
func (r *Repo) addDirFacts(child *ast.Fact, commonLex *dotlingo.Lexicon, dir string) (*ast.Fact, error) {
	dirs := []string{}

	for dir != "." {
		err := child.IncrementLevel(1)
		if err != nil {
			return nil, errors.Trace(err)
		}

		dir = filepath.Dir(dir)
		dirs = append(dirs, filepath.Base(dir))
		child = &ast.Fact{
			Name:      "dir",
			Namespace: commonLex.Ident,
			LexiconID: fmt.Sprintf("%s/%s/%s", commonLex.Owner, commonLex.Name, commonLex.Version),
			Level:     0,
			Arg: ast.FactList{
				child,
			},
		}
	}
	return child, nil
}

// addPatchFacts nests the directory scoped query inside facts representing any unstaged changes.
func (r *Repo) addPatchFacts(child *ast.Fact, gitLex *dotlingo.Lexicon) (*ast.Fact, error) {
	patches, err := r.Patches()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if len(patches) == 0 {
		return child, nil
	}

	lexId := fmt.Sprintf("%s/%s/%s", gitLex.Owner, gitLex.Name, gitLex.Version)

	patchString := strings.Join(patches, "\n")
	patchString = "__$```$__" + strings.Replace(patchString, "\n", "\\\\n", -1) + "__$```$__"

	err = child.IncrementLevel(1)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &ast.Fact{
		Name:      "patch",
		Namespace: gitLex.Ident,
		LexiconID: lexId,
		Level:     0,
		Arg: ast.FactList{
			&ast.Fact{
				Name:  "diff",
				Arg:   ast.StringArg(patchString),
				Level: 1,
			},
			child,
		},
	}, nil
}

// addRepoFacts nests the whole query inside a set of facts representing the current repo.
func (r *Repo) addRepoFacts(child *ast.Fact, gitLex *dotlingo.Lexicon, host string) (*ast.Fact, error) {
	owner, repoName, err := r.OwnerAndNameFromRemote()
	if err != nil {
		return nil, errors.Annotate(err, "\nlocal vcs error")
	}

	sha, err := r.CurrentCommitId()
	if err != nil {
		return nil, errors.Trace(err)
	}

	lexId := fmt.Sprintf("%s/%s/%s", gitLex.Owner, gitLex.Name, gitLex.Version)

	err = child.IncrementLevel(2)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// lexiconID is unnecessary if the dotlingo will be stringified and parsed again, as it is determined at
	// parse time from the imported lexicon. But the eventual goal is to only parse once.
	return &ast.Fact{
		Name:      "repo",
		Namespace: gitLex.Ident,
		LexiconID: lexId,
		Level:     0,
		Arg: ast.FactList{
			&ast.Fact{
				Name:  "name",
				Arg:   ast.StringArg(repoName),
				Level: 1,
			},
			&ast.Fact{
				Name:  "host",
				Arg:   ast.StringArg(host),
				Level: 1,
			},
			&ast.Fact{
				Name:  "owner",
				Arg:   ast.StringArg(owner),
				Level: 1,
			},
			&ast.Fact{
				Name:      "commit",
				Namespace: gitLex.Ident,
				LexiconID: lexId,
				Level:     1,
				Arg: ast.FactList{
					&ast.Fact{
						Name:  "sha",
						Arg:   ast.StringArg(sha),
						Level: 2,
					},
					child,
				},
			},
		},
	}, nil
}

// Get all .lingo files that apply to the current review.
func (r *Repo) getDotlingoFiles(workingDir string) (map[string]string, error) {
	paths, err := r.getDotlingoFilepaths(workingDir)
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
			lingoSrc, err := r.ReadFile(path)
			if err != nil {
				return nil, errors.Trace(err)
			}
			lingoFiles[path] = lingoSrc
		}
	}
	return lingoFiles, nil
}

// gets the file paths of all the .lingo files in the repo
func (r *Repo) getDotlingoFilepaths(directory string) ([]string, error) {
	staged, err := gitCMD("ls-tree", "-r", "--name-only", "--full-tree", "HEAD")
	if err != nil {
		return nil, errors.Trace(err)
	}

	unstaged, err := gitCMD("ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, errors.Trace(err)
	}

	files := strings.Split(staged, "\n")
	files = append(files, strings.Split(unstaged, "\n")...)

	dotlingoFilepaths := []string{}
	for _, filepath := range files {
		if strings.HasSuffix(filepath, ".lingo") {
			dotlingoFilepaths = append(dotlingoFilepaths, filepath)
		}
	}

	return dotlingoFilepaths, nil
}
