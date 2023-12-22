package main

/*
[Version Control System - Stage 2/4: Add & config](https://hyperskill.org/projects/177/stages/910/implement)
-------------------------------------------------------------------------------
[Operations with strings](https://hyperskill.org/learn/step/18548)
[Advanced Input](https://hyperskill.org/learn/step/18567)
[Errors](https://hyperskill.org/learn/step/16774)
[Reading files](https://hyperskill.org/learn/step/16702)
[Writing data to files](https://hyperskill.org/learn/step/17627)
[Anonymous functions and Closures](https://hyperskill.org/learn/step/22032)
[Functional decomposition](https://hyperskill.org/learn/step/17506)
[Public and private scopes](https://hyperskill.org/learn/step/17514)
[Advanced usage of structs](https://hyperskill.org/learn/step/17498)
[Methods](https://hyperskill.org/learn/step/17739)
[Debugging Go code](https://hyperskill.org/learn/step/23076)
*/

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// Command constants
const (
	CommandConfig   = "config"
	CommandAdd      = "add"
	CommandLog      = "log"
	CommandCommit   = "commit"
	CommandCheckout = "checkout"
)

// Directory and File Configuration
const (
	VcsDirectory   = "./vcs"
	ConfigFileName = "config.txt"
	IndexFileName  = "index.txt"
)

// User Interface Messages
const (
	HelpMessage = `These are SVCS commands:
config     Get and set a username.
add        Add a file to the index.
log        Show commit logs.
commit     Save changes.
checkout   Restore a file.`

	HelpCommand = "--help"

	DefaultMessage = "'%s' is not a SVCS command.\n"

	PromptWhoAreYou      = "Please, tell me who you are."
	PromptFileNotFound   = "Can't find '%s'.\n"
	PromptFileTracked    = "The file '%s' is tracked.\n"
	PromptTrackedFiles   = "Tracked files:"
	PromptUsernameIs     = "The username is %s.\n"
	PromptAddFileToIndex = "Add a file to the index."
)

func parseArguments() []string {
	flag.Usage = func() {
		fmt.Println(HelpMessage)
	}

	flag.Parse()
	return flag.Args()
}

type VersionControlSystem struct {
	ConfigFilePath string
	IndexFilePath  string
}

func (vcs *VersionControlSystem) Run(args []string) {
	commands := map[string]func([]string){
		CommandConfig:   vcs.Config,
		CommandAdd:      vcs.Add,
		CommandLog:      vcs.Log,
		CommandCommit:   vcs.Commit,
		CommandCheckout: vcs.Checkout,
	}

	command := args[0]
	if action, exists := commands[command]; exists {
		action(args)
	} else {
		fmt.Printf(DefaultMessage, command)
	}
}

func (vcs *VersionControlSystem) Config(args []string) {
	if len(args) == 1 {
		username, err := vcs.ReadConfig()
		if err != nil {
			fmt.Println(PromptWhoAreYou)
		} else {
			fmt.Printf(PromptUsernameIs, username)
		}
		return
	}

	username := args[1]
	err := vcs.WriteConfig(username)
	if err != nil {
		log.Printf("cannot write config: %v\n", err)
		return
	}
	fmt.Printf(PromptUsernameIs, username)
}

func (vcs *VersionControlSystem) ReadConfig() (string, error) {
	data, err := os.ReadFile(vcs.ConfigFilePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (vcs *VersionControlSystem) WriteConfig(username string) error {
	err := os.MkdirAll(VcsDirectory, os.ModePerm)
	if err != nil {
		return err
	}
	return os.WriteFile(vcs.ConfigFilePath, []byte(username), os.ModePerm)
}

func (vcs *VersionControlSystem) Add(args []string) {
	if len(args) == 1 {
		files, err := vcs.ReadIndex()
		if err != nil {
			fmt.Println(PromptAddFileToIndex)
		} else {
			fmt.Println(PromptTrackedFiles)
			for _, file := range files {
				fmt.Println(file)
			}
		}
		return
	}

	filename := args[1]
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf(PromptFileNotFound, filename)
		return
	}

	err := vcs.AddFileToIndex(filename)
	if err != nil {
		log.Printf("cannot add file to index: %v\n", err)
		return
	}
	fmt.Printf(PromptFileTracked, filename)
}

func (vcs *VersionControlSystem) ReadIndex() ([]string, error) {
	file, err := os.Open(vcs.IndexFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var files []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}
	return files, scanner.Err()
}

func (vcs *VersionControlSystem) AddFileToIndex(filename string) error {
	files, err := vcs.ReadIndex()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	for _, file := range files {
		if file == filename {
			return nil
		}
	}

	files = append(files, filename)
	return vcs.WriteIndex(files)
}

func (*VersionControlSystem) Log([]string) {
	fmt.Println("Show commit logs.")
}

func (*VersionControlSystem) Commit([]string) {
	fmt.Println("Save changes.")
}

func (*VersionControlSystem) Checkout([]string) {
	fmt.Println("Restore a file.")
}

func (vcs *VersionControlSystem) WriteIndex(files []string) error {
	err := os.MkdirAll(VcsDirectory, os.ModePerm)
	if err != nil {
		log.Printf("cannot create directory %s: %v", VcsDirectory, err)
		return err
	}

	file, err := os.Create(vcs.IndexFilePath)
	if err != nil {
		log.Printf("cannot create file %s: %v", vcs.IndexFilePath, err)
		return err
	}
	defer file.Close()

	for _, f := range files {
		_, err = file.WriteString(f + "\n")
		if err != nil {
			log.Printf("cannot write to file %s: %v", vcs.IndexFilePath, err)
			return err
		}
	}

	return nil
}

func NewVersionControlSystem() (*VersionControlSystem, error) {
	err := os.MkdirAll(VcsDirectory, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create VCS directory: %w", err)
	}

	return &VersionControlSystem{
		ConfigFilePath: VcsDirectory + "/" + ConfigFileName,
		IndexFilePath:  VcsDirectory + "/" + IndexFileName,
	}, nil
}

func main() {
	args := parseArguments()

	if len(args) == 0 || args[0] == HelpCommand {
		fmt.Println(HelpMessage)
		return
	}

	vcs, err := NewVersionControlSystem()
	if err != nil {
		log.Fatalf("failed to initialize simple version control system: %v", err)
	}
	vcs.Run(args)
}
