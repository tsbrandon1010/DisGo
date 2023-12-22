package main

import (
	"fmt"
	"log"
	"main/audio"
	"main/dlp"
	"main/util"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const COMMAND_PREFIX = "!"

var helpString = "!play (/play)\n!stop (/stop)\n!start (/start)\n!clear (/clear)"

var SlashCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "help",
		Description: "List commands",
	},
	{
		Name:        "play",
		Description: "play music",
	},
	{
		Name:        "stop",
		Description: "stop (pause music)",
	},
	{
		Name:        "start",
		Description: "start (resume paused music)",
	},
	{
		Name:        "clear",
		Description: "clear queued music",
	},
}

func SlashCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	switch data.Name {
	case "help":
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: helpString,
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
	case "help":
		s.ChannelMessageSend(m.ChannelID, helpString)
	case "play":
		AudioPlayHandler(s, m, audioService, command)
	case "stop":
		AudioStopHandler(s, m)
	case "start":
		AudioStartHandler(s, m)
	case "skip":
		AudioSkipHandler(s, m)
	case "clear":
		AudioClearHandler(s, m)
	}
}

func AudioPlayHandler(s *discordgo.Session, m *discordgo.MessageCreate, as *dlp.AudioService, command string) {
	args := strings.TrimSpace(strings.Replace(m.Content, fmt.Sprintf("%s%s", COMMAND_PREFIX, command), "", 1))

	log.Println(command, args)

	sc := guilds.guildStreams[m.GuildID]
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
			err = audio.Worker(s, sc, m.GuildID, vcID)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to start voice worker: %s", err.Error()))
				os.Remove(media.FilePath)
				return
			}
		}()
	}
}

func AudioStopHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	guilds.mu.Lock()
	sc := guilds.guildStreams[m.GuildID]
	guilds.mu.Unlock()
	sc.StreamingSession.SetPaused(true)
}

func AudioSkipHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	guilds.mu.Lock()
	sc := guilds.guildStreams[m.GuildID]
	guilds.mu.Unlock()
	sc.UserActions.Skip()
}

func AudioClearHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	guilds.mu.Lock()
	sc := guilds.guildStreams[m.GuildID]
	guilds.mu.Unlock()
	sc.StopStreaming()
}

func AudioStartHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	guilds.mu.Lock()
	sc := guilds.guildStreams[m.GuildID]
	guilds.mu.Unlock()

	log.Println(sc.StreamingSession)
	sc.StreamingSession.SetPaused(false)
}
