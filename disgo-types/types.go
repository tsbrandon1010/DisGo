package disgotypes

type Media struct {
	Title     string
	FilePath  string
	Uploader  string
	URL       string
	Thumbnail string
	Duration  int
}

type StreamingChannel struct {
	GuildID      string
	MediaChannel chan *Media
	UserActions  *UserActions
}

func (sc *StreamingChannel) PrepairStreaming(maxQueueSize int) {
	sc.MediaChannel = make(chan *Media, maxQueueSize)
	sc.UserActions = &UserActions{
		SkipChannel: make(chan bool, 1),
		StopChannel: make(chan bool, 1),
	}
}

func (sc *StreamingChannel) IsStreaming() bool {
	return sc.MediaChannel != nil
}

func (sc *StreamingChannel) QueueMedia(media *Media) {
	sc.MediaChannel <- media
}

func (sc *StreamingChannel) IsQueueFull() bool {
	return len(sc.MediaChannel) != 0 && cap(sc.MediaChannel) == len(sc.MediaChannel)
}

func (sc *StreamingChannel) GetMediaQueueSize() int {
	return len(sc.MediaChannel)
}

func (sc *StreamingChannel) StopStreaming() {
	close(sc.MediaChannel)
	sc.MediaChannel = nil
	sc.UserActions = nil
}

type UserActions struct {
	SkipChannel chan bool
	StopChannel chan bool
	Stopped     bool
}

func (ua *UserActions) Stop() {
	ua.Stopped = true
	ua.StopChannel <- true
}

func (ua *UserActions) Skip() {
	ua.SkipChannel <- true
}
