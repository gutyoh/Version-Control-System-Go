package main

/*
[Version Control System - Stage 3/4: Log & commit](https://hyperskill.org/projects/177/stages/911/implement)
-------------------------------------------------------------------------------
[Getting file attributes](https://hyperskill.org/learn/step/18851)
[Working with file paths](https://hyperskill.org/learn/step/18961)
[Hashing strings and files](https://hyperskill.org/learn/step/19064)
[Debugging Go code in GoLand](https://hyperskill.org/learn/step/23118)
[Working with time](https://hyperskill.org/learn/step/19297)
*/

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	VcsDirectory     = "./vcs"
	ConfigFileName   = "config.txt"
	IndexFileName    = "index.txt"
	CommitsDirectory = VcsDirectory + "/commits"
	LogFileName      = VcsDirectory + "/log.txt"
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

// Commit Related Messages
const (
	NoCommitsYetMessage     = "No commits yet."
	ChangesCommittedMessage = "Changes are committed."
	NothingToCommitMessage  = "Nothing to commit."
	MessageNotPassedMessage = "Message was not passed."
	CommitLogFormat         = "commit %s\nAuthor: %s\n%s\n\n"
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
	data, err := os.ReadFile(LogFileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println(NoCommitsYetMessage)
			return
		}
		log.Printf("cannot read log file: %v\n", err)
		return
	}

	if len(data) == 0 {
		fmt.Println(NoCommitsYetMessage)
	} else {
		fmt.Print(string(data))
	}
}

func (vcs *VersionControlSystem) Commit(args []string) {
	if len(args) < 2 {
		fmt.Println(MessageNotPassedMessage)
		return
	}

	message := args[1]
	username, err := vcs.ReadConfig()
	if err != nil {
		fmt.Println(PromptWhoAreYou)
		return
	}

	indexFiles, err := vcs.ReadIndex()
	if err != nil {
		log.Printf("cannot read index file: %v\n", err)
		return
	}

	if len(indexFiles) == 0 {
		fmt.Println(NothingToCommitMessage)
		return
	}

	if !vcs.HasChanges(indexFiles) {
		fmt.Println(NothingToCommitMessage)
		return
	}

	commitID := vcs.CreateCommit(indexFiles)
	vcs.WriteLog(commitID, username, message)

	fmt.Println(ChangesCommittedMessage)
}

func (vcs *VersionControlSystem) HasChanges(files []string) bool {
	lastCommitFiles, err := vcs.GetLastCommitFiles()
	if err != nil {
		log.Printf("cannot get last commit files: %v\n", err)
	}
	for _, file := range files {
		currentHash := vcs.GetFileHash(file)
		lastCommitHash, exists := lastCommitFiles[file]
		if !exists || currentHash != lastCommitHash {
			return true
		}
	}
	return false
}

func (vcs *VersionControlSystem) CreateCommit(files []string) string {
	commitID := vcs.GenerateCommitID()
	commitPath := filepath.Join(CommitsDirectory, commitID)
	err := os.MkdirAll(commitPath, os.ModePerm)
	if err != nil {
		log.Printf("cannot create directory %s: %v\n", commitPath, err)
		return ""
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("cannot read file %s: %v\n", file, err)
		}
		commitFilePath := filepath.Join(commitPath, filepath.Base(file))
		err = os.WriteFile(commitFilePath, content, os.ModePerm)
		if err != nil {
			log.Printf("cannot write file %s: %v\n", commitFilePath, err)
			return ""
		}
	}

	return commitID
}

func (*VersionControlSystem) WriteLog(commitID, username, message string) {
	logEntry := fmt.Sprintf(CommitLogFormat, commitID, username, message)

	existingLog, err := os.ReadFile(LogFileName)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("cannot read log file: %v\n", err)
		return
	}

	newLogContent := logEntry + string(existingLog)
	err = os.WriteFile(LogFileName, []byte(newLogContent), os.ModePerm)
	if err != nil {
		log.Printf("cannot write to log file: %v\n", err)
		return
	}
}

func (*VersionControlSystem) GetFileHash(filename string) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("cannot read file %s: %v\n", filename, err)
		return ""
	}

	fileHash := sha256.New()
	fileHash.Write(data)

	return fmt.Sprintf("%x", fileHash.Sum(nil))
}

func (*VersionControlSystem) GenerateCommitID() string {
	timestamp := time.Now().UnixNano()
	randomPart := fmt.Sprintf("%d", timestamp)

	commitHash := sha256.New()
	commitHash.Write([]byte(randomPart))

	return fmt.Sprintf("%x", commitHash.Sum(nil))
}

func (vcs *VersionControlSystem) GetLastCommitFiles() (map[string]string, error) {
	entries, err := os.ReadDir(CommitsDirectory)
	if err != nil {
		return nil, err
	}

	var lastCommitDirName string
	for _, entry := range entries {
		if entry.IsDir() {
			if lastCommitDirName < entry.Name() {
				lastCommitDirName = entry.Name()
			}
		}
	}

	if lastCommitDirName == "" {
		return nil, nil
	}

	lastCommitPath := filepath.Join(CommitsDirectory, lastCommitDirName)
	files, err := os.ReadDir(lastCommitPath)
	if err != nil {
		return nil, err
	}

	fileHashes := make(map[string]string)
	for _, file := range files {
		filename := filepath.Join(lastCommitPath, file.Name())
		fileHashes[file.Name()] = vcs.GetFileHash(filename)
	}

	return fileHashes, nil
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
