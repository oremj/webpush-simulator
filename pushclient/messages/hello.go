package messages

type HelloResp struct {
	MessageType string   `json:"messageType"`
	UAID        string   `json:"uaid"`
	Status      int      `json:"status"`
	Ping        float64  `json:"ping"`
	Env         string   `json:"env"`
	ChannelIDs  []string `json:"channelIDs"`
	UseWebPush  bool     `json:"use_webpush"`
}

type Hello struct {
	MessageType string   `json:"messageType"`
	UAID        string   `json:"uaid"`
	ChannelIDs  []string `json:"channelIDs"`
	UseWebPush  bool     `json:"use_webpush"`
}

func NewHello() *Hello {
	return &Hello{
		MessageType: "hello",
		UseWebPush:  true,
	}
}
