package directive

import "encoding/json"

type DirectiveOp int64

const (
	// Engine directive
	EngineRegisterCommunityDirectiveOp   DirectiveOp = 1000
	EngineUnregisterCommunityDirectiveOp DirectiveOp = 1001
	EngineRegisterUserDirectiveOp        DirectiveOp = 1002
	EngineUnregisterUserDirectiveOp      DirectiveOp = 1003
)

type ClientDirective struct {
	Op   DirectiveOp `json:"op"`
	Data any         `json:"data"`
}

type ServerDirective struct {
	Op   DirectiveOp     `json:"op"`
	Data json.RawMessage `json:"data"`
}
