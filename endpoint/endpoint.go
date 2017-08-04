package endpoint

import (
	"fmt"

	"github.com/googlechrome/push-encryption-go/webpush"
)

type Endpoint struct {
	sub *webpush.Subscription
}

func New(endpoint string, key, auth []byte) Endpoint {
	return Endpoint{
		sub: &webpush.Subscription{endpoint, key, auth},
	}
}

func (e Endpoint) Notify(msg string) error {
	resp, err := webpush.Send(nil, e.sub, msg, "")
	if err != nil {
		return fmt.Errorf(`webpush.Send(nil, %v, %s, ""): %v`, e.sub, msg, err)
	}
	resp.Body.Close()
	return nil
}
