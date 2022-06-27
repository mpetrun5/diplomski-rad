package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mpetrun5/diplomski-rad/util"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Communication interface {
	Broadcast(peers peer.IDSlice, msg []byte, msgType MessageType, sessionID string)
	Subscribe(topic MessageType, sessionID string, channel chan *WrappedMessage)
	UnSubscribe(topic MessageType, sessionID string)
}

type communication struct {
	h                   host.Host
	fullAddressAsString string
	protocolID          protocol.ID
	streamManager       *StreamManager
	logger              zerolog.Logger
	subscribers         map[MessageType]*MessageIDSubscriber
	subscriberLocker    *sync.Mutex
}

func NewCommunication(h host.Host, protocolID protocol.ID) (Communication, error) {
	fullAddr := util.GetHostAddress(h)
	c := communication{
		h:                   h,
		fullAddressAsString: fullAddr,
		protocolID:          protocolID,
		streamManager:       NewStreamMgr(),
		logger:              log.With().Str("peer", h.ID().Pretty()).Logger(),
		subscribers:         make(map[MessageType]*MessageIDSubscriber),
		subscriberLocker:    &sync.Mutex{},
	}
	c.listen()
	return &c, nil
}

// Broadcast writes a message to an outgoing stream for each peer requested.
func (c *communication) Broadcast(peers peer.IDSlice, msg []byte, msgType MessageType, sessionID string) {
	hostID := c.h.ID().Pretty()

	wMsg := WrappedMessage{
		MessageType: msgType,
		SessionID:   sessionID,
		Payload:     msg,
		From:        c.h.ID(),
	}
	bMsg, err := json.Marshal(wMsg)
	if err != nil {
		c.logger.Error().Err(err).Msg("unable to marshal message")
		return
	}

	c.logger.Info().Str("sessionID", sessionID).Str("msgType", msgType.String()).Msg("broadcasting message")

	for _, p := range peers {
		if hostID != p.Pretty() {
			stream, err := c.h.NewStream(context.TODO(), p, c.protocolID)
			if err != nil {
				c.logger.Error().Err(err).Msg(
					fmt.Sprintf("unable to open stream toward %s", p.Pretty()),
				)
				return
			}

			c.logger.Info().Str("sessionID", sessionID).Str("to", p.Pretty()).Msg("sending message")
			err = WriteStreamWithBuffer(bMsg, stream)
			if err != nil {
				c.logger.Error().Str("to", p.Pretty()).Err(err).Msg("unable to send message")
				return
			}
			c.streamManager.AddStream(sessionID, stream)
		}
	}
}

// Subscribe stars listening on given topic and session ID and sends messages to provided channel.
func (c *communication) Subscribe(topic MessageType, sessionID string, channel chan *WrappedMessage) {
	c.subscriberLocker.Lock()
	defer c.subscriberLocker.Unlock()

	messageIDSubscriber, ok := c.subscribers[topic]
	if !ok {
		messageIDSubscriber = NewMessageIDSubscriber()
		c.subscribers[topic] = messageIDSubscriber
	}

	c.logger.Info().Str("sessionID", sessionID).Msgf("subscribed to topic %s", topic.String())
	messageIDSubscriber.Subscribe(sessionID, channel)
}

func (c communication) UnSubscribe(topic MessageType, sessionID string) {
	c.subscriberLocker.Lock()
	defer c.subscriberLocker.Unlock()

	sessionIDSubscribers, ok := c.subscribers[topic]
	if !ok {
		c.logger.Debug().Msgf("cannot find the given channels %s", topic.String())
		return
	}
	if nil == sessionIDSubscribers {
		return
	}

	sessionIDSubscribers.UnSubscribe(sessionID)
}

func (c communication) getSubscriber(topic MessageType, sessionID string) chan *WrappedMessage {
	c.subscriberLocker.Lock()
	defer c.subscriberLocker.Unlock()

	messageIDSubscriber, ok := c.subscribers[topic]
	if !ok {
		c.logger.Debug().Msgf("fail to find subscribers for %s", topic)
		return nil
	}

	return messageIDSubscriber.GetSubscriber(sessionID)
}

func (c *communication) listen() {
	c.h.SetStreamHandler(c.protocolID, func(s network.Stream) {
		msg, err := c.processMessageFromStream(s)
		if err != nil {
			c.logger.Error().Err(err).Msg("unable to process message")
			return
		}

		subscriber := c.getSubscriber(msg.MessageType, msg.SessionID)
		subscriber <- msg
	})
}

func (c *communication) processMessageFromStream(s network.Stream) (*WrappedMessage, error) {
	msgBytes, err := ReadStreamWithBuffer(s)
	if err != nil {
		c.streamManager.AddStream("UNKNOWN", s)
		return nil, err
	}

	var wrappedMsg WrappedMessage
	if err := json.Unmarshal(msgBytes, &wrappedMsg); nil != err {
		c.streamManager.AddStream("UNKNOWN", s)
		return nil, err
	}

	c.streamManager.AddStream(wrappedMsg.SessionID, s)

	c.logger.Info().Str(
		"from", wrappedMsg.From.Pretty()).Str(
		"msg", wrappedMsg.MessageType.String()).Str(
		"sessionID", wrappedMsg.SessionID).Msg(
		"processed message",
	)

	return &wrappedMsg, nil
}
