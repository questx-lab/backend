package gamestate

type Action interface {
	Apply(*GameState) error
	Revert(*GameState) error
}
