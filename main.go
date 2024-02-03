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

	// Получаем токен из файла.
	token := getToken()

	discord, err := discordgo.New("Bot " + token)
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
	defer file.Close()

	tokenBytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	token := string(tokenBytes)

	return token
}
