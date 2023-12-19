package audio

import (
	"log"
	"main/dca"
	disgotypes "main/disgo-types"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Worker(s *discordgo.Session, sc *disgotypes.StreamingChannel, guildID string, channelID string) error {
	defer sc.StopStreaming()

	voice, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil
	}

	for media := range sc.MediaChannel {
		if !voice.Ready {
			voice.Disconnect()
			voice, err = s.ChannelVoiceJoin(guildID, channelID, false, true)
			if err != nil {
				return err
			}
		}

		_ = voice.Speaking(true)
		if !sc.UserActions.Stopped {
			play(voice, media, sc)
		}
		_ = os.Remove(media.FilePath)
		if len(sc.MediaChannel) == 0 {
			break
		}

		time.Sleep(500 * time.Millisecond)
		_ = voice.Speaking(false)
	}

	voice.Disconnect()
	return nil
}

func play(v *discordgo.VoiceConnection, m *disgotypes.Media, sc *disgotypes.StreamingChannel) {
	options := dca.StdEncodeOptions
	options.BufferedFrames = 100
	options.FrameDuration = 20
	options.CompressionLevel = 5
	options.Bitrate = 96

	encodeSession, err := dca.EncodeFile(m.FilePath, options)
	if err != nil {
		log.Print("could not create encoding session: ", err)
		return
	}
	defer encodeSession.Cleanup()

	time.Sleep(500 * time.Millisecond)

	done := make(chan error)
	dca.NewStream(encodeSession, v, done)

	select {
	case err := <-done:
		if err != nil {
			log.Print("error during the streaming: ", err)
			return
		}

	case <-sc.UserActions.SkipChannel:
		_ = encodeSession.Stop()
	case <-sc.UserActions.StopChannel:
		_ = encodeSession.Stop()
	}

}
