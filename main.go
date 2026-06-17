package main

import (
	"fmt"

	"github.com/zeelna/golang-blog-aggregator/internal/config"
)

func main() {
	fmt.Println("Main: read file")
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Failed to successfully call Read() ")
	}
	if err = cfg.SetUser("zeelna"); err != nil {
		fmt.Print("Failed to successfully call .SetUser() ")
	}
	fmt.Println("Main: successfuly read file")

	// Print struct we read and wrote:
	fmt.Printf("%+v\n", cfg)
}
