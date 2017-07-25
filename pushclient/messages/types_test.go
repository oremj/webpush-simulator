package messages

import "testing"

func TestMessageType(t *testing.T) {
	cases := []struct {
		Msg      []byte
		Expected Type
	}{
		{[]byte(`{"messageType": "hello"}`), TypeHello},
	}

	for _, c := range cases {
		if MessageType(c.Msg) != c.Expected {
			t.Errorf("messageType failed for: %s", c.Msg)
		}
	}
}
