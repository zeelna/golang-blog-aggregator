package main

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/zeelna/golang-blog-aggregator/internal/database"
)

import (
	//"bufio"
	"fmt"
	"log"
	"os"
	//"github.com/zeelna/golang-blog-aggregator/internal/cmds"
	"github.com/zeelna/golang-blog-aggregator/internal/config"
)

type State struct {
	db     *database.Queries
	config *config.Config
}

func main() {
	// Step #0. Database config
	// Read from ~/.gatorconfig.json (add it's filepath to .gitignore to avoid credential leak)
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Failed to read filepath ~/.gatorconfig.json")
		return
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		fmt.Print("Failed to open database")
		return
	}
	// use the generated 'database' package to create new *database.Queries and store into 'state' struct
	dbQueries := database.New(db)

	// Set state with the database.Queries and database configuration (URL and username)
	var state = State{
		db:     dbQueries,
		config: &cfg,
	}

	/* Step #2: Read command-line arguments and verify if correct */
	// DEBUG: Print struct before we write:
	//fmt.Printf("%+v\n", cfg)
	if len(os.Args) < 2 {
		log.Fatalf("error: not enough arguments provided")
	}

	argsWithoutProg := os.Args[1:]
	cleanedInputs := cleanArgs(argsWithoutProg)

	// Step #3: Create the command, once user's CLI input verified
	// initialize 'command' object
	var cmd command
	cmd, err = makeCommand(cleanedInputs)
	if err != nil {
		fmt.Print(err)
	}
	// Create 'commands' that holds map of allowed 'command' we can run
	cmds := commands{
		allowedCommands: make(map[string]func(*State, command) error),
	}

	// Step #4: Register the created command, to be run / allowed
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	// DEBUG:
	//fmt.Printf("%v\n", cmds)

	// Step #5: Run the command
	if err := cmds.run(&state, cmd); err != nil {
		log.Fatal(err)
	}

	// DEBUG:
	//fmt.Printf("%+v\n", state.config)

	return

} // end of main

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

func handlerReset(s *State, cmd command) error {
	if err := s.db.ResetUsers(context.Background()); err != nil {
		return fmt.Errorf("error, failed to reset database of users")
	}
	if err := (*s).config.SetUser(""); err != nil {
		return err
	}

	fmt.Println("Successfully deleted the database of users.")
	return nil
}

// Function to handle command <login some_username>. Update State struct, if some_username passed
func handlerRegister(s *State, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("error, command <register> expects a single argument")
	}
	username := cmd.args[0]
	//if err := (*s).config.SetUser(username); err != nil {
	//	return err
	//}
	//fmt.Println(fmt.Sprintf("User '%s' has been set", username))

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})

	if err != nil {
		return fmt.Errorf("error: user already exists in table")
	}

	if err := (*s).config.SetUser(user.Name); err != nil {
		return err
	}

	fmt.Printf("User %s created\n", user.Name)
	fmt.Printf(
		"User ID: %v\nUser name: %v\nCreated_At: %v\nUpdated_At: %v\n",
		user.ID, user.Name, user.CreatedAt, user.UpdatedAt,
	)
	return nil
}

// Function to handle command <login some_username>. Update State struct, if some_username passed
func handlerLogin(s *State, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("error, command <login> expects a single argument")
	}
	username := cmd.args[0]

	// must run "sqlc generate" via CLI each time we update ./sql/queries/*.sql
	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return err
	}

	if err := (*s).config.SetUser(user.Name); err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("User '%s' has been set", user.Name))
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
	// lowercase the command only
	args[0] = strings.ToLower(args[0])
	for i, arg := range args {
		args[i] = strings.TrimSpace(arg)
	}
	return args
}

func makeCommand(inputs []string) (command, error) {
	commandName := inputs[0]
	arguments := inputs[1:]
	return command{name: commandName, args: arguments}, nil
}
