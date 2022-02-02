package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	// Get the gitignore files
	ignores, err := GetGitignores()
	if err != nil {
		panic(err)
	}

	b, err := json.MarshalIndent(ignores, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

}
