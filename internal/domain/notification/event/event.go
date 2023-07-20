package event

import "encoding/json"

type Event interface {
	Op() string
}

type Metadata struct {
	To string `json:"to"`
}

type EventRequest struct {
	Op       string          `json:"o"`
	Data     json.RawMessage `json:"d"`
	Metadata Metadata        `json:"m"`
}

type EventResponse struct {
	Op   string          `json:"o"`
	Seq  int64           `json:"s"`
	Data json.RawMessage `json:"d"`
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

func Format(event *EventRequest, seq int64) *EventResponse {
	return &EventResponse{
		Op:   event.Op,
		Seq:  seq,
		Data: event.Data,
	}
}
