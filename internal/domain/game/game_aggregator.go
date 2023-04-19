package game

import "github.com/questx-lab/backend/internal/domain/gamestate"

type GameAggregator struct {
	c   <-chan gamestate.Action
	hub GameHub
}

func NewGameAggregator(roomID string, router GameRouter, hub GameHub) (*GameAggregator, error) {
	c, err := router.Register(roomID)
	if err != nil {
		return nil, err
	}

	return &GameAggregator{
		c:   c,
		hub: hub,
	}, nil
}

func (aggregator *GameAggregator) Run() {
	isStoped := false
	var actionBundle []gamestate.Action

	// Call broadcast here to trigger game hub process.
	aggregator.hub.Broadcast()

	for !isStoped {
		select {
		case <-aggregator.hub.Done():
			// Send the action bundle to
			aggregator.hub.Broadcast(actionBundle...)
			actionBundle = nil

		case action, ok := <-aggregator.c:
			if !ok {
				isStoped = true
				break
			}
			actionBundle = append(actionBundle, action)
		}
	}
}
