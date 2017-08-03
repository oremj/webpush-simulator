package pushclient

type Client struct {
	UAID     string
	channels *Channels
}

func New(uaid string) *Client {
	return &Client{
		UAID:     uaid,
		channels: new(Channels),
	}
}

func (c *Client) Channels() *Channels {
	if c.channels == nil {
		c.channels = new(Channels)
	}
	return c.channels
}
