package deepseek

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type DeepSeek struct {
	api string
}

func NewDeepSeek(apipath string) *DeepSeek {
	api := ReadApi(apipath)
	return &DeepSeek{
		api: api,
	}
}
func (d *DeepSeek) Conversation() {
	reader := bufio.NewReader(os.Stdin)

	req := NewDSRequest(d.api)

	for {
		fmt.Print("> ")
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)
		if strings.ToLower(userInput) == "quit" || strings.ToLower(userInput) == "exit" {
			fmt.Println("Bye")
			break
		}
		if strings.ToLower(userInput) == "new" {
			req.ClearMsg()
			continue
		}
		if userInput == "" {
			continue
		}

		req.AddUserMsg(userInput)
		reader, writer := io.Pipe()
		go req.Send(writer)

		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					fmt.Println()
					// log.Println("stream closed")
					break
				}
				log.Fatal("error reading from stream", err)
			}
			content := string(buf[:n])
			fmt.Print(content)
		}
	}
}

func (d *DeepSeek) QueryOnce() {

	req := NewDSRequest(d.api)

	req.AddUserMsg("hello, how are you")
	reader, writer := io.Pipe()
	go req.Send(writer)

	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				// log.Println("stream closed")
				break
			}
			log.Fatal("error reading from stream", err)
		}
		content := string(buf[:n])
		fmt.Print(content)
	}

}
