package entity

import "github.com/questx-lab/backend/pkg/enum"

type ThemeType string

var (
	DarkTheme  = enum.New(ThemeType("dark"))
	LightTheme = enum.New(ThemeType("light"))
)

type GameMap struct {
	Base
	Name   string
	Width  int
	Height int
	Theme  ThemeType
}
