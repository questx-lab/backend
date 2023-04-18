package gamestate

type Action interface {
	Apply(*GameState) error
}
