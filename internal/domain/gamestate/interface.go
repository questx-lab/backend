package gamestate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"github.com/questx-lab/backend/internal/model"
)

type Action interface {
	Apply(*GameState) error
}

func SerializeAction(id int, a Action) model.GameActionResponse {
	resp := model.GameActionResponse{
		ID:    id,
		Type:  reflect.TypeOf(a).Name(),
		Value: structs.Map(a),
	}

	return resp
}

func DeserializeAction(req model.GameActionRequest) (Action, error) {
	switch req.Type {
	case strings.ToLower(reflect.TypeOf(Move{}).Name()):
		action := Move{}
		err := mapstructure.Decode(req.Value, &action)
		return &action, err
	}

	return nil, fmt.Errorf("invalid game action type %s", req.Type)
}
