package main

import (
	"log"
	disgotypes "main/disgo-types"
	"main/dlp"
	"os"
	"os/signal"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Guilds struct {
	guildStreams map[string]*disgotypes.StreamingChannel
	mu           sync.Mutex
}

var (
	_      = godotenv.Load()
	guilds = Guilds{
		guildStreams: make(map[string]*disgotypes.StreamingChannel),
	}
	audioService = dlp.CreateService(3610)
	token        = os.Getenv("TOKEN")
	appID        = os.Getenv("APP_ID")
)

func createGuild(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	log.Println(token, appID)
	_, _ = s.ApplicationCommandBulkOverwrite(appID, event.Guild.ID, SlashCommands)
	guilds.mu.Lock()
	guilds.guildStreams[event.Guild.ID] = &disgotypes.StreamingChannel{GuildID: event.Guild.ID}
	guilds.mu.Unlock()
}

func updateStatus(s *discordgo.Session) {
	err := s.UpdateStreamingStatus(0, "!help or /help", os.Getenv("STATUS_URL"))
	if err != nil {
		log.Println(err)
	}
}

func main() {

	if token == "" {
		log.Println("Invalid Discord API Token... ")
		return
	}
	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Println("when creating a session", err)
	}

	sess.AddHandler(createGuild)
	sess.AddHandler(SlashCommandHandler)
	sess.AddHandler(PrefixCommandHandler)

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	err = sess.Open()
	if err != nil {
		log.Fatal("while opening a session: ", err)
	}
	defer sess.Close()

	updateStatus(sess)

	halt := make(chan os.Signal, 1)
	signal.Notify(halt, os.Interrupt)
	log.Println("Ctrl+C to exit")
	<-halt

}
