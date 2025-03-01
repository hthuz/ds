package deepseek

import (
	"log"
	"os"
	"strings"
)

func ReadApi(apipath string) string {
	data, err := os.ReadFile(apipath)
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(data), "\n")

	if len(lines) < 1 {
		log.Fatal("api should only be one line")
	}
	return lines[0]
}
