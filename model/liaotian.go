package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"sync"
)

type Session struct {
	Target
	To      Target
	NatsUrl string
	Stream  nats.JetStreamContext
	Topic   string
}

func NewSession(target Target, to Target, natsUrl string) (*Session, error) {
	s := &Session{Target: target, To: to, NatsUrl: natsUrl}
	connect, err := nats.Connect(s.NatsUrl)
	if err != nil {
		return nil, err
	}
	stream, err := connect.JetStream()
	if err != nil {
		return nil, err
	}
	s.Stream = stream
	UUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	s.Topic = UUID.String()
	config := &nats.StreamConfig{Name: s.Topic}
	_, err = s.Stream.AddStream(config)
	return s, err
}

type Target struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type MsgType string

const (
	TextMsg  MsgType = "text"
	PicMsg   MsgType = "pic"
	VideoMsg MsgType = "video"
	HrefMsg  MsgType = "href"
)

type Msg struct {
	Type MsgType
	Body []byte
}

func (r *Session) Send(msg Msg) error {
	var da = struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}{
		string(msg.Type),
		string(msg.Body),
	}
	marshal, err := json.Marshal(da)
	if err != nil {
		return err
	}
	_, err = r.Stream.Publish(r.Topic, marshal)
	return err
}

func (r *Session) Recv() (<-chan Msg, error) {
	ch := make(chan Msg)
	_, err := r.Stream.Subscribe(r.Topic, func(msg *nats.Msg) {
		var da = struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		}{}
		err := json.Unmarshal(msg.Data, &da)
		if err != nil {
			fmt.Println(err)
			return
		}
		ch <- Msg{Type: MsgType(da.Type), Body: []byte(da.Content)}
	})
	return ch, err
}

type liaotian map[string][]*Session

var Liaotian = make(liaotian)

var mut sync.Mutex

var DefaultUrl = nats.DefaultURL

func (r liaotian) Add(to, from Target) (string, error) {
	mut.Lock()
	defer mut.Unlock()
	ss, ok := r[from.Id]
	var se *Session
	if ok {
		fl := true
		for _, s := range ss {
			if s.To.Id == to.Id {
				se = s
				fl = false
				break
			}
		}
		if fl {
			ssi, err := NewSession(from, to, DefaultUrl)
			if err != nil {
				return "", err
			}
			se = ssi
			ss = append(ss, se)
		}
	} else {
		ssi, err := NewSession(from, to, DefaultUrl)
		if err != nil {
			return "", err
		}
		se = ssi
		ss = make([]*Session, 0)
		ss = append(ss, se)
	}
	r[from.Id] = ss
	return se.Topic, nil
}

func (r liaotian) Del(to, from string) error {
	mut.Lock()
	defer mut.Unlock()
	ss, ok := r[from]
	if !ok {
		return errors.New("not liaotian")
	}
	fl := false
	ns := make([]*Session, 0)
	for _, s := range ss {
		if s.To.Id == to {
			fl = true
		} else {
			ns = append(ns, s)
		}
	}
	if !fl {
		return errors.New(fmt.Sprintf("not to %s session", to))
	}
	if len(ns) == 0 {
		delete(r, from)
		return nil
	}
	r[from] = ns
	return nil
}

func (r liaotian) Get(to, from string) (*Session, error) {
	mut.Lock()
	defer mut.Unlock()
	ss, ok := r[from]
	if !ok {
		return nil, errors.New("not liaotian")
	}
	for _, s := range ss {
		if s.To.Id == to {
			return s, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("not to %s session", to))
}
