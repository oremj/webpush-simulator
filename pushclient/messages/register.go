package messages

import "github.com/google/uuid"

type RegisterResp struct {
	MessageType  string `json:"messageType"`
	ChannelID    string `json:"channelID"`
	PushEndpoint string `json:"pushEndpoint"`
	Status       int    `json:"status"`
}

type Register struct {
	MessageType string `json:"messageType"`
	ChannelID   string `json:"channelID"`
}

func NewRegister() *Register {
	return &Register{
		MessageType: "register",
		ChannelID:   uuid.New().String(),
	}
}
