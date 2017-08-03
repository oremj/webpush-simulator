package pushclient

import "github.com/oremj/webpush-simulator/pushclient/messages"

type Handler interface {
	HandleRegister(*Conn, messages.RegisterResp)
	HandleNotification(*Conn, messages.NotificationResp)
}
