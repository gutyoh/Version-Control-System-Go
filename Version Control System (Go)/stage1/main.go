package main

/*
[Version Control System - Stage 1/4: Help page](https://hyperskill.org/projects/177/stages/909/implement)
-------------------------------------------------------------------------------
[String formatting](https://hyperskill.org/learn/step/16860)
[Control statements](https://hyperskill.org/learn/step/16235)
[Working with slices](https://hyperskill.org/learn/step/15935)
[Operations with maps](https://hyperskill.org/learn/step/17206)
[Command-line arguments and flags](https://hyperskill.org/learn/step/17863)
*/

import (
	"flag"
	"fmt"
)

const (
	CommandConfig   = "config"
	CommandAdd      = "add"
	CommandLog      = "log"
	CommandCommit   = "commit"
	CommandCheckout = "checkout"

	DescriptionConfig   = "Get and set a username."
	DescriptionAdd      = "Add a file to the index."
	DescriptionLog      = "Show commit logs."
	DescriptionCommit   = "Save changes."
	DescriptionCheckout = "Restore a file."

	HelpMessage = `These are SVCS commands:
config     Get and set a username.
add        Add a file to the index.
log        Show commit logs.
commit     Save changes.
checkout   Restore a file.`

	HelpCommand = "--help"

	DefaultMessage = "'%s' is not a SVCS command.\n"
)

func main() {
	flag.Usage = func() {
		fmt.Println(HelpMessage)
	}

	flag.Parse()
	args := flag.Args()

	if len(args) == 0 || args[0] == HelpCommand {
		fmt.Println(HelpMessage)
		return
	}

	commands := map[string]string{
		CommandConfig:   DescriptionConfig,
		CommandAdd:      DescriptionAdd,
		CommandLog:      DescriptionLog,
		CommandCommit:   DescriptionCommit,
		CommandCheckout: DescriptionCheckout,
	}

	command := args[0]
	if description, ok := commands[command]; ok {
		fmt.Println(description)
	} else {
		fmt.Printf(DefaultMessage, command)
	}
}
