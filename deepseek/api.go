package deepseek

import (
	"log"
	"os"
	"strings"
)

func ReadApi(apipath string) string {
	// Read the entire file
	data, err := os.ReadFile(apipath)
	if err != nil {
		log.Fatal(err)
	}

	// Split the content into lines
	lines := strings.Split(string(data), "\n")

	// Get the first line (if it exists)
	if len(lines) < 1 {
		log.Fatal("api should only be one line")
	}
	return lines[0]
}
