package model

type DiscordInteractionType int

const (
	PING DiscordInteractionType = iota + 1
	APPLICATION_COMMAND
	MESSAGE_COMPONENT
	APPLICATION_COMMAND_AUTOCOMPLETE
	MODAL_SUBMIT
)

type PostDiscordInteractRequest struct {
	ID            string
	ApplicationID string
	Type          DiscordInteractionType
}

type PostDiscordInteractResponse struct{}
