package emojis

type Emoji struct {
	ID       string
	Name     string
	Animated bool
}

func (e Emoji) String() string {
	if e.Animated {
		return "<a:" + e.Name + ":" + e.ID + ">"
	}
	return "<:" + e.Name + ":" + e.ID + ">"
}

var (
	EmojiCrown        = Emoji{ID: "1418014546462773348", Name: "crown", Animated: true}
	EmojiQuestionMark = Emoji{ID: "1418015866695581708", Name: "question", Animated: true}
	EmojiChat         = Emoji{ID: "1418013205992575116", Name: "chat", Animated: true}
	EmojiNote         = Emoji{ID: "1418015996651765770", Name: "note", Animated: true}
	EmojiTop          = Emoji{ID: "1418012513584283709", Name: "top", Animated: true}
	EmojiStar         = Emoji{ID: "1418011800724705310", Name: "star", Animated: true}
	EmojiFire         = Emoji{ID: "1418017773354881156", Name: "fire", Animated: true}
	EmojiMic          = Emoji{ID: "1418021307089551471", Name: "mic", Animated: true}
	EmojiMic2         = Emoji{ID: "1418021315708981258", Name: "mic2", Animated: true}
	EmojiPlay         = Emoji{ID: "1418021326228295692", Name: "play", Animated: true}
	EmojiAlbum        = Emoji{ID: "1418021336110075944", Name: "album", Animated: true}
	EmojiCalendar     = Emoji{ID: "1418022075527860244", Name: "calendar", Animated: true}

	// status
	EmojiCross   = Emoji{ID: "1418016016642080848", Name: "cross", Animated: true}
	EmojiCheck   = Emoji{ID: "1418016005732565002", Name: "check", Animated: true}
	EmojiUpdate  = Emoji{ID: "1418014272415469578", Name: "update", Animated: true}
	EmojiWarning = Emoji{ID: "1418013632293507204", Name: "warning", Animated: true}

	// rank
	EmojiRankOne   = Emoji{ID: "1418015934312087582", Name: "rank1", Animated: true}
	EmojiRankTwo   = Emoji{ID: "1418015960862036139", Name: "rank2", Animated: true}
	EmojiRankThree = Emoji{ID: "1418015987562709022", Name: "rank3", Animated: true}
)
