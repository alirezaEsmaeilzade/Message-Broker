package broker

import (
	"context"
	"log"
	"sync"
	"therealbroker/pkg/broker"
	"time"
)
type FetchMessage struct{
	create time.Time
	msg *broker.Message
}
type IDCreator struct {
	lock sync.Mutex
	lastID int
}
func NewIDCreator()*IDCreator{
	return &IDCreator{lastID: 0}
}
func (i *IDCreator)getID() int{
	i.lock.Lock()
	defer i.lock.Unlock()
	i.lastID++
	return i.lastID
}
type Topic struct{
	lock sync.Mutex
	subject string
	subscribers []*Subscriber
	idCreator *IDCreator
	//messages map[int] *FetchMessage
	expiredMessages map[int] bool
	expiredMessageID chan int
	db DB
}
func NewTopic(name string, db DB) *Topic{
	t := &Topic{
		subject: name,
		subscribers: make([]*Subscriber, 0),
		idCreator: NewIDCreator(),
		//messages: make(map[int]*FetchMessage),
		expiredMessages: make(map[int]bool),
		expiredMessageID: make(chan int),
		db: db,
	}
	go t.FindExpiredMessages()
	return t
}
func (t * Topic) FindExpiredMessages(){
	for ID := range t.expiredMessageID{
		Id := ID
		t.db.DeleteMessage(Id)
		//t.expiredMessages[Id] = true
	}
}
func (t *Topic) SendSignalOfMessageExpired(ID int, d time.Duration){//new
	//fmt.Println("signal")
	if d == 0{
		return
	}
	time.AfterFunc(d ,func() {
		ID := ID
		go func() {// can't delete why?
			t.expiredMessageID <- ID
		}()
		//fmt.Println("func run")
		//t.db.DeleteMessage(ID)
		t.lock.Lock()
		defer t.lock.Unlock()
		t.expiredMessages[ID] = true
	})
}
func (t *Topic) AddMessage(msg *broker.Message) int{
	//t.messages[id] = &FetchMessage{msg: msg, create: creatTime}
	msgID, err :=  t.db.AddMessage(t.subject, msg)
	if err != nil{
		log.Fatalln("can't add message", err)/// must change this
		return -1
	}
	t.lock.Lock()
	t.expiredMessages[msgID] = false
	t.lock.Unlock()
	t.SendSignalOfMessageExpired(msgID, msg.Expiration)
	return msgID
	//time.AfterFunc()
}
func (t *Topic) Publish(msg broker.Message, creatTime time.Time) int{
	//t.lock.Lock()
	//defer t.lock.Unlock()
	//msgID := t.idCreator.getID()
	//fmt.Println("msgID", msgID)
	msgID := t.AddMessage(&msg)
	read := t.subscribers
	for _, j := range(read){
		j.AddMessageToQueue(msg)
		//j.receive()
	}
	return msgID
}

func (t *Topic) AddSubscriber(ctx context.Context, channel chan broker.Message){
	sub := NewSubscriber(ctx, channel)
	go sub.Listen()
	t.lock.Lock()
	defer t.lock.Unlock()
	t.subscribers = append(t.subscribers, sub)
}

func (t *Topic) FetchMessage(id int) (broker.Message, error){
	//t.lock.Lock()
	//defer t.lock.Unlock()
	////if t, ok := t.messages[id]; ok{
	//	if time.Since(t.create) < t.msg.Expiration {
	//		return *t.msg, nil
	//	}else{
	//		return broker.Message{}, broker.ErrExpiredID
	//	}
	//}
	if expiredStatus, ok := t.expiredMessages[id]; ok{
		if expiredStatus{
			return broker.Message{}, broker.ErrExpiredID
		}
		msg, err := t.db.FetchMessage(id)
		if err != nil{
			return broker.Message{}, err
		}
		return *msg, nil
	}
	return broker.Message{}, broker.ErrInvalidID
}