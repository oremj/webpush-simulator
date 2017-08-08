package messages

type Update struct {
	ChannelID string `json:"channelID"`
	Version   string `json:"version"`
}

type NotificationResp struct {
	MessageType string `json:"messageType"`
	ChannelID   string `json:"channelID"`
	Version     string `json:"version"`
	Data        string `json:"data"`
	Headers     struct {
		Encoding   string `json:"encoding"`
		Encryption string `json:"encryption"`
		CryptoKey  string `json:"crypto_key"`
	} `json:"headers"`
}

type Ack struct {
	MessageType string   `json:"messageType"`
	Updates     []Update `json:"updates"`
}

func NewAck() *Ack {
	return &Ack{
		MessageType: "ack",
		Updates:     make([]Update, 0),
	}
}

func (a *Ack) Add(channelID string, version string) {
	a.Updates = append(a.Updates, Update{
		ChannelID: channelID,
		Version:   version,
	})
}
