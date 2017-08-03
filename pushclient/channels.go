package pushclient

type Channels struct {
	ids []string
}

func (c *Channels) Add(id string) {
	c.ids = append(c.ids, id)
}

func (c *Channels) Remove(id string) {
	for i := 0; i < len(c.ids); i++ {
		if id == c.ids[i] {
			c.ids[i] = c.ids[len(c.ids)-1]
			c.ids = c.ids[:len(c.ids)-1]
		}
	}
}

func (c *Channels) List() []string {
	res := make([]string, len(c.ids))
	for i, id := range c.ids {
		res[i] = id
	}
	return res
}

func (c *Channels) Set(ids []string) {
	c.ids = make([]string, len(ids))
	for i, id := range ids {
		c.ids[i] = id
	}
}
