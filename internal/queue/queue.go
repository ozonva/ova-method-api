package queue

import "encoding/json"

//go:generate mockgen -source=$GOFILE -destination=./mock/queue.go -package=mock

type Queue interface {
	Connect() error
	Close() error
	Send(string, QueueMsg) error
}

type QueueMsg interface {
	json.Marshaler
}

type Body map[string]interface{}

type message struct {
	Action string      `json:"action"`
	Body   interface{} `json:"body"`
}

func NewMessage(action string, msg interface{}) QueueMsg {
	return &message{Action: action, Body: msg}
}

func (m *message) MarshalJSON() ([]byte, error) {
	return json.Marshal(*m)
}
