package server

import (
	"../utils/log"
	"encoding/json"
)

type Message struct {
	Payload []byte
	Client  *Client
}

type Command struct {
	Name    string `json:"name"`
	Payload string `json:"payload"`
}

func NewMessage(client *Client, payload []byte) *Message {
	m := &Message{
		Payload: payload,
		Client:  client,
	}
	return m
}

func (m *Message) Decode(v interface{}) bool {
	if err := json.Unmarshal(m.Payload, v); err != nil {
		log.Error(err)
		return false
	}
	return true
}

func (m *Message) Encode(v interface{}) bool {
	b, err := json.Marshal(v)
	if err != nil {
		log.Error(err)
		return false
	}
	m.Payload = b
	return true
}

func (m *Message) Send() {
	m.Client.Send(m)
}
