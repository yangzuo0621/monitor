package git

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path"
)

// GitExecuteResult wraps git command execution result.
type GitExecuteResult interface {
	// CheckReturnCode checks execution code.
	CheckReturnCode() error
	// Stdout returns the stdout output.
	Stdout() io.Reader
	// Stderr returns the stderr output.
	Stderr() io.Reader
}

// GitExecutor runs git command execution result..
type GitExecutor interface {
	// Run runs git command.
	Run(ctx context.Context, args ...string) (GitExecuteResult, error)
}

// NewGitFromEnv create a git executor by detecting environment configurations.
func NewGitFromEnv(rootFolder string, repository string) (GitExecutor, error) {
	const envGitCmd = "AKS_E2E_GIT"

	gitCmd := "git"
	detectGit := true

	// try environment variable
	if detectGit {
		if v := os.Getenv(envGitCmd); v != "" {
			gitCmd = v
			detectGit = false
		}
	}

	// try lookpath
	if detectGit {
		path, err := exec.LookPath("git")
		if err == nil {
			gitCmd = path
			detectGit = false
		}
	}

	// TODO: download

	executor := &shellGitExecutor{
		command:    gitCmd,
		repository: repository,
		rootFolder: rootFolder,
	}
	return executor, nil
}

type shellGitExecutor struct {
	// command - the git command
	command string
	// repository - the git repository
	repository string
	// rootFolder - the git repository parent folder
	rootFolder string
}

var _ GitExecutor = (*shellGitExecutor)(nil)

func (exe *shellGitExecutor) buildCmd(
	ctx context.Context,
	args []string,
) *exec.Cmd {
	cmd := exec.CommandContext(ctx, exe.command, args...)
	cmd.Dir = path.Join(exe.rootFolder, exe.repository)

	return cmd
}

func (exe *shellGitExecutor) Run(
	ctx context.Context,
	args ...string,
) (GitExecuteResult, error) {
	cmd := exe.buildCmd(ctx, args)
	cmdStdout := new(bytes.Buffer)
	cmdStderr := new(bytes.Buffer)
	cmd.Stdout = cmdStdout
	cmd.Stderr = cmdStderr
	// cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout
	cmdErr := cmd.Run()

	result := &shellCmdResult{
		stdout:    cmdStdout,
		stderr:    cmdStderr,
		invokeErr: cmdErr,
	}

	return result, nil
}

type shellCmdResult struct {
	stdout    io.Reader
	stderr    io.Reader
	invokeErr error
}

var _ GitExecuteResult = (*shellCmdResult)(nil)

func (r *shellCmdResult) CheckReturnCode() error {
	if r.invokeErr == nil {
		return nil
	}

	if v, ok := r.invokeErr.(*exec.ExitError); ok {
		if v.ExitCode() == 0 {
			return nil
		}
	}

	return r.invokeErr
}

func (r *shellCmdResult) Stdout() io.Reader {
	return r.stdout
}

func (r *shellCmdResult) Stderr() io.Reader {
	return r.stderr
}
