package main

import (
	_ "github.com/joho/godotenv/autoload"

	"gitlab.bbdev.team/vh/vh-srv-events/cmd"
)

func main() {
	cmd.Execute()
}
