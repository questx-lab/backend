package migration

import "context"

var Migrators = map[string]func(context.Context) error{
	"0001": Migrate0001,
}
