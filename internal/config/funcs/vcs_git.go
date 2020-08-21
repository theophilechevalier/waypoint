package funcs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

func VCSGitFuncs(path string) map[string]function.Function {
	state := &VCSGit{Path: path}

	return map[string]function.Function{
		"gitrefpretty": state.RefPrettyFunc(),
		"gitrefhash":   state.RefHashFunc(),
		"gitreftag":    state.RefTagFunc(),
	}
}

type VCSGit struct {
	// Path of the git repository. Parent directories will be searched for
	// a ".git" folder automatically.
	Path string

	initErr error
	repo    *git.Repository
}

// RefPrettyFunc returns a string format of the current Git ref. This function
// takes some liberties to humanize the output: it will use a tag if the
// ref matches a tag, it will append "+CHANGES" to the commit if there are
// uncommitted changed files, etc.
//
// You may use direct functions such as `gitrefhash` if you want the direct
// hash. Or `gitreftag` to get the current tag.
//
// waypoint:gitrefpretty
func (s *VCSGit) RefPrettyFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{},
		Type:   function.StaticReturnType(cty.String),
		Impl:   s.refPrettyFunc,
	})
}

func (s *VCSGit) refPrettyFunc(args []cty.Value, retType cty.Type) (cty.Value, error) {
	if err := s.init(); err != nil {
		return cty.UnknownVal(cty.String), err
	}

	ref, err := s.repo.Head()
	if err != nil {
		return cty.UnknownVal(cty.String), err
	}
	result := ref.Hash().String()

	// Get the tags
	iter, err := s.repo.Tags()
	if err != nil {
		return cty.UnknownVal(cty.String), err
	}
	defer iter.Close()
	for {
		tagRef, err := iter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}
		if tagRef.Hash() == ref.Hash() {
			result = tagRef.Name().Short()
			break
		}
	}

	// To determine if there are changes we subprocess because go-git's Status
	// function is really, really slow sadly. On the waypoint repo at the time
	// of this commit, go-git took 12s on my machine vs. 50ms for `git`.
	cmd := exec.Command("git", "diff", "--quiet")
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard
	if err := cmd.Run(); err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			return cty.UnknownVal(cty.String), fmt.Errorf("error executing git: %s", err)
		}

		if exitError.ExitCode() != 0 {
			result += fmt.Sprintf("_CHANGES_%d", time.Now().Unix())
		}
	}

	return cty.StringVal(result), nil
}

// RefHashFunc returns the full hash of the HEAD ref.
//
// waypoint:gitrefhash
func (s *VCSGit) RefHashFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{},
		Type:   function.StaticReturnType(cty.String),
		Impl:   s.refHashFunc,
	})
}

func (s *VCSGit) refHashFunc(args []cty.Value, retType cty.Type) (cty.Value, error) {
	if err := s.init(); err != nil {
		return cty.UnknownVal(cty.String), err
	}

	ref, err := s.repo.Head()
	if err != nil {
		return cty.UnknownVal(cty.String), err
	}

	return cty.StringVal(ref.Hash().String()), nil
}

// RefTagFunc returns the tag of the HEAD ref or empty if not tag is found.
//
// waypoint:gitreftag
func (s *VCSGit) RefTagFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{},
		Type:   function.StaticReturnType(cty.String),
		Impl:   s.refTagFunc,
	})
}

func (s *VCSGit) refTagFunc(args []cty.Value, retType cty.Type) (cty.Value, error) {
	if err := s.init(); err != nil {
		return cty.UnknownVal(cty.String), err
	}

	ref, err := s.repo.Head()
	if err != nil {
		return cty.UnknownVal(cty.String), err
	}

	// Get the tags
	iter, err := s.repo.Tags()
	if err != nil {
		return cty.UnknownVal(cty.String), err
	}
	defer iter.Close()
	for {
		tagRef, err := iter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}
		if tagRef.Hash() == ref.Hash() {
			return cty.StringVal(tagRef.Name().Short()), nil
		}
	}

	return cty.StringVal(""), nil
}

func (s *VCSGit) init() error {
	// If we initialized already return
	if s.initErr != nil {
		return s.initErr
	}
	if s.repo != nil {
		return nil
	}

	// Check if `git` is installed. We'll use this sometimes.
	if _, err := exec.LookPath("git"); err != nil {
		s.initErr = fmt.Errorf("git was not found on the system and is required")
		return s.initErr
	}

	// Open the repo
	repo, err := git.PlainOpenWithOptions(s.Path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		s.initErr = err
		return err
	}
	s.repo = repo
	return nil
}