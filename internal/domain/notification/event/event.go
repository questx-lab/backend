package event

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Event interface {
	Op() string
	Unmarshal(b []byte) error
}

type CommunityMetadata struct {
	ID     string `json:"id"`
	Handle string `json:"handle"`
}

type Metadata struct {
	ToCommunities []CommunityMetadata `json:"to_communities"`
	ToUsers       []string            `json:"to_users"`
}

type EventRequest struct {
	Op       string          `json:"o"`
	Seq      uint64          `json:"s"`
	Data     json.RawMessage `json:"d"`
	Metadata Metadata        `json:"m"`
}

type EventResponse struct {
	Op   string          `json:"o"`
	Seq  uint64          `json:"s"`
	Data json.RawMessage `json:"d"`
}

type EventMessaging struct {
	Op   string `json:"o"`
	Seq  string `json:"s"`
	Data string `json:"d"`
}

func New(ev Event, metadata *Metadata) *EventRequest {
	b, err := json.Marshal(ev)
	if err != nil {
		return &EventRequest{}
	}

	req := &EventRequest{
		Op:   ev.Op(),
		Data: b,
	}
	if metadata != nil {
		req.Metadata = *metadata
	}

	return req
}

func Format(event *EventRequest, seq uint64) *EventResponse {
	return &EventResponse{
		Op:   event.Op,
		Seq:  seq,
		Data: event.Data,
	}
}

func FormatMessaging(event *EventRequest) (map[string]string, Event, error) {
	message := EventMessaging{
		Op:   event.Op,
		Seq:  strconv.FormatUint(event.Seq, 10),
		Data: string(event.Data),
	}

	b, err := json.Marshal(message)
	if err != nil {
		return nil, nil, err
	}

	m := map[string]string{}
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, nil, err
	}

	// raw event
	var originEvent Event
	switch event.Op {
	case ReadyEvent{}.Op():
		originEvent = &ReadyEvent{}
	case MessageCreatedEvent{}.Op():
		originEvent = &MessageCreatedEvent{}
	case MessageDeletedEvent{}.Op():
		originEvent = &MessageDeletedEvent{}
	case MessageUpdatedEvent{}.Op():
		originEvent = &MessageUpdatedEvent{}
	case ChannelCreatedEvent{}.Op():
		originEvent = &ChannelCreatedEvent{}
	case ChannelDeletedEvent{}.Op():
		originEvent = &ChannelDeletedEvent{}
	case ChannelUpdatedEvent{}.Op():
		originEvent = &ChannelUpdatedEvent{}
	case ReactionAddedEvent{}.Op():
		originEvent = &ReactionAddedEvent{}
	case ReactionRemovedEvent{}.Op():
		originEvent = &ReactionRemovedEvent{}
	case FollowCommunityEvent{}.Op():
		originEvent = &FollowCommunityEvent{}
	case ChangeUserStatusEvent{}.Op():
		originEvent = &ChangeUserStatusEvent{}
	default:
		return nil, nil, fmt.Errorf("not support push notification with event %s", event.Op)
	}

	if err := originEvent.Unmarshal(b); err != nil {
		return nil, nil, err
	}

	return m, originEvent, nil
}
