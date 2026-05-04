package main

import (
	"fmt"
	"os"
	"youtube-service/Download"
	"youtube-service/Youtube"
)

// --- entry point ---
func main() {
	var arg string
	var token string
	for arg != "exit" {
		arg = ""
		print("Choices are authenticate, download, and delete : ")
		fmt.Scanln(&arg)
		switch arg {
		case "authenticate":
			token = Youtube.Init()
		case "download":
			Download.Run()
		case "delete":

			if token == "" {
				fmt.Println("You need to provide a token")
				continue
			}
			err := Youtube.ClearPlaylist(token)
			if err != nil {
				fmt.Println(err)
			}

		case "exit":
			os.Exit(0)
		default:
			fmt.Println("Unknown argument. Use: download | delete | (no arg)")

		}
	}

}
