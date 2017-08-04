package pushclient

type Client struct {
	UAID     string
	channels *Channels

	keys *ClientEncryption
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

func (c *Client) Keys() *ClientEncryption {
	if c.keys == nil {
		c.keys = NewClientEncryption()
	}
	return c.keys
}
