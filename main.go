package main

import (
	"fmt"
	"log"
	"main/audio"
	disgotypes "main/disgo-types"
	"main/dlp"
	"main/util"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	guilds       = map[string]disgotypes.StreamingChannel{}
	audioService = dlp.CreateService(3610)
	APP_ID       = os.Getenv("APP_ID")
)

func createGuild(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	_, _ = s.ApplicationCommandBulkOverwrite(APP_ID, event.Guild.ID, SlashCommands)
	guilds[event.Guild.ID] = disgotypes.StreamingChannel{GuildID: event.Guild.ID}
}

func main() {
	godotenv.Load()
	token := os.Getenv("TOKEN")

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

	halt := make(chan os.Signal, 1)
	signal.Notify(halt, os.Interrupt)
	log.Println("Ctrl+C to exit")
	<-halt

	if err != nil {
		log.Panicln(err)
	}
}

const COMMAND_PREFIX = "!"

var SlashCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "hello-world",
		Description: "Say hello",
	},
}

func SlashCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	switch data.Name {
	case "hello-world":
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Hello World!",
			},
		})
		if err != nil {
			log.Println(err)
		}

	}
}

func PrefixCommandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	command := strings.Replace(strings.Split(m.Content, " ")[0], COMMAND_PREFIX, "", 1)
	log.Println(command)

	switch command {

	case "play":
		YoutubeHandler(s, m, audioService, command)
	}
}

func YoutubeHandler(s *discordgo.Session, m *discordgo.MessageCreate, as *dlp.AudioService, command string) {
	args := strings.TrimSpace(strings.Replace(m.Content, fmt.Sprintf("%s%s", COMMAND_PREFIX, command), "", 1))

	log.Println(command, args)

	sc := guilds[m.GuildID]
	if sc.IsQueueFull() {
		_, _ = s.ChannelMessageSend(m.ChannelID, "The music queue is full")
		return
	}

	vcID, err := util.GetAuthorVoiceChannel(s, m)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "User is not connected to a voice channel")
		return
	}

	media, err := as.AudioServiceRunner(args)
	if err != nil {
		log.Println(err)
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Cannot find video: %s", err))
		return
	}

	log.Println(media)
	createNewWorker := false
	if !sc.IsStreaming() {
		sc.PrepairStreaming(10)
		createNewWorker = true
	}
	sc.QueueMedia(media)
	_, _ = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		URL:         media.URL,
		Title:       media.Title,
		Description: fmt.Sprintf("Position in queue: %d", sc.GetMediaQueueSize()),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: media.Thumbnail,
		},
	})

	if createNewWorker {
		go func() {
			err = audio.Worker(s, &sc, m.GuildID, vcID)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to start voice worker: %s", err.Error()))
				os.Remove(media.FilePath)
				return
			}
		}()
	}
}
