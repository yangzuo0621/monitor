package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
)

const gitRepoFormat = "https://dev.azure.com/%s/%s/_git/%s"

func main() {
}

func gitCloneRepo(organization string, project string, repo string, pat string, dir string) {

	auth := fmt.Sprintf(":%s", pat)
	authBase64Token := base64.StdEncoding.EncodeToString([]byte(auth))
	gitRepo := fmt.Sprintf(gitRepoFormat, organization, project, repo)
	gitCmd := exec.Command("git", "-c", fmt.Sprintf(`http.extraHeader=Authorization: Basic %s`, authBase64Token), "clone", gitRepo)

	gitCmd.Dir = dir
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr

	if err := gitCmd.Run(); err != nil {
		fmt.Printf("err: %v", err)
	}
}
