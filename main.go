package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"html"
	"io"
	"net/http"
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

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func handlerAgg(s *State, cmd command) error {
	feedUrl := "https://www.wagslane.dev/index.xml"
	feedPointer, err := fetchFeed(context.Background(), feedUrl)
	if err != nil {
		return err
	}
	feed := decodeHTML(feedPointer)

	// print feed
	fmt.Printf("Channel Title: %s\n", feed.Channel.Title)
	fmt.Printf("Channel Link: %s\n", feed.Channel.Link)
	fmt.Printf("Channel Description: \n%s\n", feed.Channel.Description)
	fmt.Printf("Channel Items: \n\n")

	for i, item := range feed.Channel.Item {
		fmt.Printf("\n- Channel Item #%d -\n", i)
		fmt.Printf("- Title: #%s\n", item.Title)
		fmt.Printf("- Link: #%s\n", item.Link)
		fmt.Printf("- Publication Date: #%s\n", item.PubDate)
		fmt.Printf("- Description: \n#%s\n", item.Description)

	}
	return nil
}

// Function to decode escaped HTML entities (like &ldquo)
func decodeHTML(rssFeed *RSSFeed) *RSSFeed {
	(*rssFeed).Channel.Title = html.UnescapeString((*rssFeed).Channel.Title)
	(*rssFeed).Channel.Description = html.UnescapeString((*rssFeed).Channel.Description)

	for i, _ := range rssFeed.Channel.Item {
		// must be "&rssFeed", because 'item := rssFeed' creates a copy.
		item := &rssFeed.Channel.Item[i] // & gives a *RSSItem pointing at the real element
		// using pointer to that item, to reassign its underlying value to be "unescapedString()"
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)

	}
	return rssFeed
}

// call API to fetch the RSS Feed and unmarshal into XML
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// Create HTTP GET Request. Shorthand //	res, err := http.Get(apiConf.Next)
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("could not create request: %w", err)
	}
	// Set Headers
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("User-Agent", "gator")

	// Create a Client object, make the HTTP Request and receive the HTTP Response
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("network Error: %v", err)
	}
	defer res.Body.Close()

	// Verify successful HTTP GET Request
	if res.StatusCode != http.StatusOK {
		return &RSSFeed{}, fmt.Errorf("could not retrieve those location areas. non-OK HTTP status: %s", res.Status)
	}
	if res.StatusCode > 299 {
		return &RSSFeed{}, fmt.Errorf("Response failed with status code: %d and\nbody: %v\n", res.StatusCode, res.Body)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("could not read response body: %w", err)
	}

	var rssFeed RSSFeed
	if err := xml.Unmarshal(data, &rssFeed); err != nil {
		return &RSSFeed{}, fmt.Errorf("could not unmarshal: %w", err)
	}
	return &rssFeed, nil
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
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddFeed)

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

// Function to handle command <register some_username>. Update State struct, does not yet exist
func handlerRegister(s *State, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("error, command <register> expects a single argument")
	}
	username := cmd.args[0]

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

func handlerUsers(s *State, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error, failed to get users from database")
	}
	if len(users) == 0 {
		return fmt.Errorf("error, no users in database to display")
	}

	for _, user := range users {
		if user.Name == (*s).config.CurrentUsername {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAddFeed(s *State, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("error, command <addfeed> expects a two argument. Usage: addfeed my_feed https://example.org")
	}
	// Get current logged (from .gatorconfig.json), as mark as author of the feed
	username := (*s).config.CurrentUsername
	// Retrieve that username from table 'users' and get ID
	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return err
	}
	// 'addfeed' command requires two arguments, to set the new Feed entry in database
	feedName := cmd.args[0]
	feedUrl := cmd.args[1]
	// connect the feed to the username
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    user.ID, // retrieve above, from 'users' table
	})
	if err != nil {
		return fmt.Errorf("error: could not create feed due to failed SQL operation. ERR: %v", err)
	}

	fmt.Printf("- Feed -\n")
	fmt.Printf("ID: %v\n", feed.ID)
	fmt.Printf("Created At: %v\n", feed.CreatedAt)
	fmt.Printf("Updated At: %v\n", feed.UpdatedAt)
	fmt.Printf("Name: %v\n", feed.Name)
	fmt.Printf("URL: %v\n", feed.Url)
	fmt.Printf("Author ID: %v\n", feed.UserID)

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
