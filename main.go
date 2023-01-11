package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cli/go-gh"
	"gopkg.in/ini.v1"
	"log"
	"path"
	"strings"
)

type Config struct {
	account    string
	repo       string
	protocol   string
	baseDir    string
	createRepo bool
	shellCmd   []string
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
	section := cfgIni.Section("gh-cd")
	baseDir := section.Key("basedir").String()
	if baseDir != "" {
		cfg.baseDir = baseDir
	}
	protocol := section.Key("protocol").String()
	if protocol != "" {
		cfg.protocol = protocol
	}
	createRepo, err := section.Key("create-repo").Bool()
	if err == nil {
		cfg.createRepo = createRepo
	}
	shellCmd := section.Key("shell-cmd").String()
	if shellCmd != "" {
		splitFn := func(c rune) bool {
			return c == ' '
		}
		cfg.shellCmd = strings.FieldsFunc(shellCmd, splitFn)
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

func runShell(config Config) error {
	var cmdStr []string
	if len(config.shellCmd) > 0 {
		cmdStr = config.shellCmd
	} else {
		cmdStr = []string{detectShell()}
	}
	fmt.Println("🍻 Running shell in repo")
	cmd := exec.Command(cmdStr[0], cmdStr[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = config.getBaseDir()
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
		if config.createRepo {
			// Try to create a repo (if it exists the command fails)
			repo := fmt.Sprintf("%s/%s", config.account, config.repo)
			if _, _, err := gh.Exec("repo", "create", repo, "--public"); err != nil {
				fmt.Println("❌ Repository not created")
			} else {
				fmt.Println("✅ Repository created")
			}
		}
		var url string

		if config.protocol == "ssh" {
			url = config.sshUrl()
		} else if config.protocol == "https" {
			url = config.httpsUrl()
		} else {
			fmt.Println(config.protocol)
			fmt.Println("❓ GH_PROTO must be https or ssh")
			os.Exit(-1)
		}

		// Now clone it
		fmt.Println("⏳ Please wait...")
		gh.Exec("repo", "clone", url, directory, "--", "--recursive")
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			fmt.Println("🛑 Could not clone repository")
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
		account:    response.Login,
		protocol:   "ssh",
		baseDir:    path.Join(dirname, "repo"),
		createRepo: true,
		shellCmd:   []string{},
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
	runShell(config)

}
