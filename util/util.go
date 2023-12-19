package util

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

func GetAuthorVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate) (string, error) {
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		return "", err
	}

	for _, voiceInstace := range guild.VoiceStates {
		if voiceInstace.UserID == m.Author.ID {
			return voiceInstace.ChannelID, nil
		}
	}

	return "", errors.New("user not connected to channel")
}
