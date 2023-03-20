package entity

import "time"

type User struct {
	Base
	Name     string
	Age      int32
	BirthDay time.Time
}
