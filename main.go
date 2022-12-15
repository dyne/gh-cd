package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cli/go-gh"
	"gopkg.in/ini.v1"
	"log"
	"path"
)

type Config struct {
	account  string
	repo     string
	protocol string
	baseDir  string
}

func (cfg *Config) loadINI() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	cfgIni, err := ini.Load(path.Join(dirname, ".gitconfig"))
	if err != nil {
		return
	}
	baseDir := cfgIni.Section("gh-cd").Key("basedir").String()
	if baseDir != "" {
		cfg.baseDir = baseDir
	}
	protocol := cfgIni.Section("gh-cd").Key("protocol").String()
	if protocol != "" {
		cfg.protocol = protocol
	}
}

func (cfg Config) getBaseDir() string {
	repodir := fmt.Sprintf("github.com/%s/%s", cfg.account, cfg.repo)

	return path.Join(cfg.baseDir, repodir)
}
func (cfg Config) sshUrl() string {
	return fmt.Sprintf("git@github.com:%s/%s.git", cfg.account, cfg.repo)
}
func (cfg Config) httpsUrl() string {
	return fmt.Sprintf("https://github.com/%s/%s.git", cfg.account, cfg.repo)
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		return shell
	}
	//  	if runtime.GOOS == "windows" {
	//  		return os.Getenv("COMSPEC")
	//  	}
	return "/bin/sh"
}

func runShell(path string) error {
	fmt.Println("Running shell in repo")
	cmd := exec.Command(detectShell())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path
	// TODO: detect recursive executions
	//cmd.Env = append(os.Environ(), "GHQ_LOOK="+filepath.ToSlash(repo.RelPath))
	return cmd.Run()
}

func runGH(config Config) {
	if _, err := os.Stat(config.baseDir); os.IsNotExist(err) {
		if err := os.Mkdir(config.baseDir, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
	directory := config.getBaseDir()
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		// Try to create a repo (if it exists the command fails)
		repo := fmt.Sprintf("%s/%s", config.account, config.repo)
		_, _, err := gh.Exec("repo", "create", repo, "--public")
		if err != nil {
			log.Println(err)
		}
		var url string

		if config.protocol == "ssh" {
			url = config.sshUrl()
		} else if config.protocol == "https" {
			url = config.httpsUrl()
		} else {
			fmt.Println(config.protocol)
			fmt.Println("GH_PROTO must be https or ssh")
			os.Exit(-1)
		}

		// Now clone it
		gh.Exec("repo", "clone", url, directory, "--", "--recursive")
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			fmt.Println("Could not clone repository")
			os.Exit(-1)
		}
	}
}

func main() {
	client, err := gh.RESTClient(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	response := struct{ Login string }{}
	err = client.Get("user", &response)
	if err != nil {
		fmt.Println(err)
		return
	}
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	config := Config{
		account:  response.Login,
		protocol: "ssh",
		baseDir:  path.Join(dirname, "repo"),
	}
	config.loadINI()

	if len(os.Args) == 2 {
		config.repo = os.Args[1]
	} else if len(os.Args) == 3 {
		config.account = os.Args[1]
		config.repo = os.Args[2]
	} else {
		fmt.Println("Usage: gh cd [user] [repo]")
		os.Exit(-1)
	}

	runGH(config)
	runShell(config.getBaseDir())

}
