package main

import (
	//"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	//"github.com/zeelna/golang-blog-aggregator/internal/cmds"
	"github.com/zeelna/golang-blog-aggregator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Failed to read filepath ~/.gatorconfig.json")
	}

	/* Task 1 */
	// DEBUG: Print struct before we write:
	//fmt.Printf("%+v\n", cfg)

	// Set state with the configuration
	var state = State{
		config: &cfg,
	}

	if len(os.Args) < 2 {
		log.Fatalf("error: not enough arguments provided")
	}

	/*
		input := strings.Join(os.Args[1:], " ")
		cleanedInputs := cleanInput(input)
	*/

	argsWithoutProg := os.Args[1:]
	cleanedInputs := cleanArgs(argsWithoutProg)

	// initialize 'command' object
	var cmd command
	cmd, err = makeCommand(cleanedInputs)
	if err != nil {
		fmt.Print(err)
	}

	// Initialize commands object.
	cmds := commands{
		allowedCommands: make(map[string]func(*State, command) error),
	}

	cmds.register("login", handlerLogin)
	// DEBUG:
	//fmt.Printf("%v\n", cmds)

	if err := cmds.run(&state, cmd); err != nil {
		log.Fatal(err)
	}

	// DEBUG:
	//fmt.Printf("%+v\n", state.config)

	return

} // end of main

type State struct {
	config *config.Config
}

type command struct {
	name string   // login
	args []string // ['some_username']
}

type commands struct {
	// This will be a map of command names to their handler functions.
	allowedCommands map[string]func(*State, command) error // handler___ function (ex: 'handlerLogin')
}

// method registers a new handler function for a command name.
func (c *commands) register(name string, f func(*State, command) error) {
	if _, ok := (*c).allowedCommands[name]; ok {
		fmt.Println("error, already exists")
	}
	(*c).allowedCommands[name] = f
}

// Function to handle command <login some_username>. Update State struct, if some_username passed
func handlerLogin(s *State, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("error, command <login> expects a single argument")
	}
	username := cmd.args[0]
	if err := (*s).config.SetUser(username); err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("User '%s' has been set", username))
	return nil
}

// Method run's given command with the provided State if it exists
func (c *commands) run(s *State, cmd command) error {
	callback, ok := c.allowedCommands[cmd.name]
	if !ok {
		return fmt.Errorf("error: run() does not find command")
	}
	if err := callback(s, cmd); err != nil {
		return err
	}
	return nil
}

// Verify the input is trimmed of whitespaced, lowercased and split into slice by each whitespace between the words.
func cleanInput(text string) []string {
	trimmed := strings.TrimSpace(text)
	lowered := strings.ToLower(trimmed)
	if lowered == "" {
		return []string{}
	}
	collection := strings.Fields(lowered)
	return collection
}

func cleanArgs(args []string) []string {
	for i, arg := range args {
		args[i] = strings.ToLower(strings.TrimSpace(arg))
	}
	return args
}

func makeCommand(inputs []string) (command, error) {
	commandName := inputs[0]
	arguments := inputs[1:]
	return command{name: commandName, args: arguments}, nil
}
