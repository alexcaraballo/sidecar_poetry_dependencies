package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cli/go-gh/v2"
	"github.com/sidecar-poetry-dependencies/models"
)

func getRepositoryURL(organization string, repository string) string {
	return fmt.Sprintf("%s/%s", organization, repository)
}

func cloneRepository(organization string, repository string) {
	repoURL := getRepositoryURL(organization, repository)
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

func poetryAdd(dep, version, extras string, repo *models.PoetryRepository) error {
	args := []string{"add", fmt.Sprintf("%s==%s", dep, version)}

	if repo != nil {
		fmt.Printf("Configuring poetry repository: %s\n", repo.Name)
		cmd := exec.Command("poetry", "config", fmt.Sprintf("repositories.%s", repo.Name), repo.URL)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}

		cmd = exec.Command("poetry", "config", fmt.Sprintf("http-basic.%s", repo.Name), repo.User, repo.Password)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

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

func commitAndPushChange(branch string, url string) error {
	
	cmd := exec.Command("git", "remote", "set-url", "origin", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err

	}
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

	stdout, stderr, err := gh.Exec(
		"pr", 
		"create", 
		"-d",
		"-t ':arrow_up: chore(deps): update dependency via action'",
		"-b 'This PR was automatically created by the Sidecar Poetry Dependencies Action.'",
	)

	if err != nil {
		fmt.Println("Error pr create:", err)
		fmt.Println("STDERR:", stderr.String())
		return err
	}
	fmt.Println("STDOUT:", stdout.String())
	fmt.Println("Pr created successfully.")

	return nil
}

func main() {
	token := os.Getenv("INPUT_TOKEN")
	organization := os.Getenv("INPUT_ORGANIZATION")
	repositories := os.Getenv("INPUT_REPOSITORY")
	branch := os.Getenv("INPUT_BRANCH")
	dep := os.Getenv("INPUT_PACKAGE")
	version := os.Getenv("INPUT_PACKAGE_VERSION")
	extras := os.Getenv("INPUT_EXTRA_POETRY_ARGS")
	repository_name := os.Getenv("INPUT_REPOSITORY_NAME")
	repository_url := os.Getenv("INPUT_REPOSITORY_URL")
	repository_user := os.Getenv("INPUT_REPOSITORY_USERNAME")
	repository_password := os.Getenv("INPUT_REPOSITORY_PASSWORD")

	var repo *models.PoetryRepository = nil

	if repository_name != "" && repository_url != ""  && repository_user != "" && repository_password != "" {
		repo := models.PoetryRepository{
				Name:     repository_name,
				URL:      repository_url,
				User:     repository_user,
				Password: repository_password,
			}
			fmt.Printf("Using custom poetry repository: %+v\n", repo)
	}

	os.Setenv("GH_TOKEN", token)

	stdout, stderr, err := gh.Exec("auth", "status")
	if err != nil {
		fmt.Println("Error checking GH auth status:", err)
		os.Exit(1)
	}
	fmt.Println("GH AUTH STATUS:")
	fmt.Println(stdout.String())
	fmt.Println(stderr.String())

	repositoriesList := strings.Split(strings.TrimSpace(repositories), ",")
	for _, repository := range repositoriesList {
		repository = strings.TrimSpace(repository)
		if repository != "" {
			cloneRepository(organization, repository)

			err = os.Chdir(repository)
			if err != nil {
				fmt.Println("Error changing directory:", err)
				os.Exit(1)
			}

			err = poetryAdd(dep, version, extras, repo)
			if err != nil {
				fmt.Println("Error running poetry add:", err)
				os.Exit(1)
			}
			fmt.Println("Dependency added successfully.")

			err = commitAndPushChange(branch, fmt.Sprintf("https://%s@github.com/%s/%s.git", token, organization, repository))
			if err != nil {
				fmt.Println("Error committing and pushing changes:", err)
				os.Exit(1)
			}
			err = os.Chdir("..")
			if err != nil {
				fmt.Println("Error changing back to parent directory:", err)
				os.Exit(1)
			}
		}
	}
	
	fmt.Println("Changes committed and pushed successfully.")
}
