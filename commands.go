package main

import (
	"fmt"
	"log"
	"main/audio"
	disgotypes "main/disgo-types"
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
		Options: []*discordgo.ApplicationCommandOption{{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "video",
			Description: "The video (YouTube URL or fuzzy search) to play.",
			Required:    true,
		}},
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

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(data.Options))
	for _, opt := range data.Options {
		optionMap[opt.Name] = opt
	}

	if i.User != nil {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This command only works in a Discord server",
			},
		})
		if err != nil {
			log.Println(err)
		}

		return
	}

	connectionInfo := disgotypes.ConnectionInfo{
		GuildID:   i.GuildID,
		ChannelID: i.GuildID,
		AuthorID:  i.Member.User.ID,
	}

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
	case "play":
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "playing...",
			},
		})
		if err != nil {
			log.Println(err)
		}
		if query, ok := optionMap["video"]; ok {
			AudioPlayHandler(s, connectionInfo, audioService, query.StringValue())
		}

	case "stop":
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "stopping...",
			},
		})
		if err != nil {
			log.Println(err)
		}
		AudioStopHandler(s, connectionInfo)

	case "start":
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "starting...",
			},
		})
		if err != nil {
			log.Println(err)
		}
		AudioStartHandler(s, connectionInfo)

	case "skip":
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "skipping...",
			},
		})
		if err != nil {
			log.Println(err)
		}
		AudioSkipHandler(s, connectionInfo)

	case "clear":
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "clearing...",
			},
		})
		if err != nil {
			log.Println(err)
		}
		AudioClearHandler(s, connectionInfo)

	}
}

func PrefixCommandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	connectionInfo := disgotypes.ConnectionInfo{
		GuildID:   m.GuildID,
		ChannelID: m.ChannelID,
		AuthorID:  m.Author.ID,
	}

	command := strings.Replace(strings.Split(m.Content, " ")[0], COMMAND_PREFIX, "", 1)
	args := strings.TrimSpace(strings.Replace(m.Content, fmt.Sprintf("%s%s", COMMAND_PREFIX, command), "", 1))
	log.Println(command, args)

	switch command {
	case "help":
		s.ChannelMessageSend(m.ChannelID, helpString)
	case "play":
		AudioPlayHandler(s, connectionInfo, audioService, args)
	case "stop":
		AudioStopHandler(s, connectionInfo)
	case "start":
		AudioStartHandler(s, connectionInfo)
	case "skip":
		AudioSkipHandler(s, connectionInfo)
	case "clear":
		AudioClearHandler(s, connectionInfo)
	}
}

func AudioPlayHandler(s *discordgo.Session, c disgotypes.ConnectionInfo, as *dlp.AudioService, query string) {

	sc := guilds.guildStreams[c.GuildID]
	if sc.IsQueueFull() {
		_, _ = s.ChannelMessageSend(c.ChannelID, "The music queue is full")
		return
	}

	vcID, err := util.GetAuthorVoiceChannel(s, c)
	if err != nil {
		_, _ = s.ChannelMessageSend(c.ChannelID, "User is not connected to a voice channel")
		return
	}

	media, err := as.AudioServiceRunner(query)
	if err != nil {
		log.Println(err)
		_, _ = s.ChannelMessageSend(c.ChannelID, fmt.Sprintf("Cannot find video: %s", err))
		return
	}

	log.Println(media)
	createNewWorker := false
	if !sc.IsStreaming() {
		sc.PrepairStreaming(10)
		createNewWorker = true
	}
	sc.QueueMedia(media)
	_, _ = s.ChannelMessageSendEmbed(c.ChannelID, &discordgo.MessageEmbed{
		URL:         media.URL,
		Title:       media.Title,
		Description: fmt.Sprintf("Position in queue: %d", sc.GetMediaQueueSize()),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: media.Thumbnail,
		},
	})

	if createNewWorker {
		go func() {
			err = audio.Worker(s, sc, c.GuildID, vcID)
			if err != nil {
				_, _ = s.ChannelMessageSend(c.ChannelID, fmt.Sprintf("Unable to start voice worker: %s", err.Error()))
				os.Remove(media.FilePath)
				return
			}
		}()
	}
}

func AudioStopHandler(s *discordgo.Session, c disgotypes.ConnectionInfo) {
	guilds.mu.Lock()
	sc, ok := guilds.guildStreams[c.GuildID]
	if !ok || !sc.IsStreaming() {
		log.Println("Audio is not playing")
		return
	}
	guilds.mu.Unlock()
	sc.StreamingSession.SetPaused(true)
}

func AudioSkipHandler(s *discordgo.Session, c disgotypes.ConnectionInfo) {
	guilds.mu.Lock()
	sc, ok := guilds.guildStreams[c.GuildID]
	if !ok || !sc.IsStreaming() {
		return
	}
	guilds.mu.Unlock()
	sc.UserActions.Skip()
}

func AudioClearHandler(s *discordgo.Session, c disgotypes.ConnectionInfo) {
	guilds.mu.Lock()
	sc, ok := guilds.guildStreams[c.GuildID]
	if !ok || !sc.IsStreaming() {
		return
	}
	guilds.mu.Unlock()
	sc.StopStreaming()
}

func AudioStartHandler(s *discordgo.Session, c disgotypes.ConnectionInfo) {
	guilds.mu.Lock()
	sc, ok := guilds.guildStreams[c.GuildID]
	if !ok || !sc.IsStreaming() {
		return
	}
	guilds.mu.Unlock()
	sc.StreamingSession.SetPaused(false)
}
