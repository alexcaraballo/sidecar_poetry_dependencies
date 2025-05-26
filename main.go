package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cli/go-gh/v2"
)

func getRepositoryURL(token string, organization string, repository string) string {
	return fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, organization, repository)
}

func cloneRepository(token string, organization string, repository string) {
	repoURL := getRepositoryURL(token, organization, repository)
	fmt.Println("Running: gh repo clone", repoURL)
	stdout, stderr, err := gh.Exec("repo", "clone", repoURL)
	if err != nil {
		fmt.Println("Error cloning repository:", err)
		fmt.Println("STDERR:", stderr.String())
		return
	}
	fmt.Println("STDOUT:", stdout.String())
	fmt.Println("Repository cloned successfully.")
}

func poetryAdd(dep, version, extras string) error {
	args := []string{"add", fmt.Sprintf("%s==%s", dep, version)}

	extraList := strings.Split(strings.TrimSpace(extras), ",")
	for _, extra := range extraList {
		extra = strings.TrimSpace(extra)
		if extra != "" {
			args = append(args, "-E", extra)
		}
	}

	fmt.Println("Running: poetry", strings.Join(args, " "))
	cmd := exec.Command("poetry", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func commitAndPushChange(branch string) error {
	cmds := [][]string{
		{"git", "checkout", "-b", branch},
		{"git", "config", "user.name", "github-actions"},
		{"git", "config", "user.email", "github-actions@github.com"},
		{"git", "add", "pyproject.toml", "poetry.lock"},
		{"git", "commit", "-m", ":arrow_up: chore(deps): update dependency via action"},
		{"git", "push", "--set-upstream", "origin", branch},
	}

	for _, cmdArgs := range cmds {
		fmt.Printf("Running: %s\n", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	token := os.Getenv("INPUT_TOKEN")
	organization := os.Getenv("INPUT_ORGANIZATION")
	repository := os.Getenv("INPUT_REPOSITORY")
	branch := os.Getenv("INPUT_BRANCH")
	dep := os.Getenv("INPUT_PACKAGE")
	version := os.Getenv("INPUT_PACKAGE_VERSION")
	extras := os.Getenv("INPUT_EXTRA_POETRY_ARGS")

	cloneRepository(token, organization, repository)

	err := os.Chdir(repository)
	if err != nil {
		fmt.Println("Error changing directory:", err)
		return
	}

	err = poetryAdd(dep, version, extras)
	if err != nil {
		fmt.Println("Error running poetry add:", err)
		return
	}
	fmt.Println("Dependency added successfully.")

	err = commitAndPushChange(branch)
	if err != nil {
		fmt.Println("Error committing and pushing changes:", err)
		return
	}
	fmt.Println("Changes committed and pushed successfully.")
}
