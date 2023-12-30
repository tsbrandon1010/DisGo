package util

import (
	"errors"
	disgotypes "main/disgo-types"

	"github.com/bwmarrin/discordgo"
)

func GetAuthorVoiceChannel(s *discordgo.Session, c disgotypes.ConnectionInfo) (string, error) {
	guild, err := s.State.Guild(c.GuildID)
	if err != nil {
		return "", err
	}

	for _, voiceInstace := range guild.VoiceStates {
		if voiceInstace.UserID == c.AuthorID {
			return voiceInstace.ChannelID, nil
		}
	}

	return "", errors.New("user not connected to channel")
}
