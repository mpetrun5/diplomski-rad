package p2p

import "sync"

type MessageIDSubscriber struct {
	lock        *sync.Mutex
	subscribers map[string]chan *WrappedMessage
}

func NewMessageIDSubscriber() *MessageIDSubscriber {
	return &MessageIDSubscriber{
		lock:        &sync.Mutex{},
		subscribers: make(map[string]chan *WrappedMessage),
	}
}

func (ms *MessageIDSubscriber) GetSubscriber(msgID string) chan *WrappedMessage {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	c, ok := ms.subscribers[msgID]
	if !ok {
		return nil
	}
	return c
}

func (ms *MessageIDSubscriber) Subscribe(msgID string, channel chan *WrappedMessage) {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	ms.subscribers[msgID] = channel
}

func (ms *MessageIDSubscriber) UnSubscribe(msgID string) {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	delete(ms.subscribers, msgID)
}
