package event

type Event interface {
	Op() string
}

type Metadata struct {
	To string `json:"to"`
}

type EventRequest struct {
	Op       string   `json:"o"`
	Data     any      `json:"d"`
	Metadata Metadata `json:"m"`
}

type EventResponse struct {
	Op   string `json:"o"`
	Seq  int64  `json:"s"`
	Data any    `json:"d"`
}

func New(ev Event, metadata *Metadata) *EventRequest {
	req := &EventRequest{
		Op:   ev.Op(),
		Data: ev,
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
