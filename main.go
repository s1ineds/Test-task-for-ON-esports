package main

import (
	"discord-bot/structs"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

func main() {
	dBot := structs.DBot{}

	// // Токен бота здесь лучше предоставлять как параметр CLI.
	// // Если токен не предоставлен, то и бот работать не будет.
	// if os.Args[1] == "" {
	// 	log.Fatal("Please specify bot token as command line parameter.")
	//

	// Здесь токен получаем из аргументов CLI.
	discord, err := discordgo.New("Bot " + getToken())
	if err != nil {
		log.Fatal(err)
	}

	openErr := discord.Open()
	if openErr != nil {
		log.Fatal(err)
	}
	defer discord.Close()

	discord.AddHandler(dBot.SendWeatherMessage)
	discord.Identify.Intents = discordgo.IntentsGuildMessages

	// Здесь горутина main работает бесконечно, пока в канал
	// не поступит прерывание.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	<-sc
}

func getToken() string {
	file, err := os.Open("token.txt")
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	tokenBytes, err := io.ReadAll(file)

	return string(tokenBytes)
}
