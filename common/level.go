package common

import (
	"fmt"
	"strings"
)

type Level int

const (
	VERBOSE Level = 1 + iota
	DEBUG
	INFORMATION
	WARNING
	ERROR
	FATAL
)

var levels = [...]string{
	"Verbose",
	"Debug",
	"Information",
	"Warning",
	"Error",
	"Fatal",
}

func (l Level) String() string {
	return levels[l-1]
}

func LevelFromString(levelString string) (Level, error) {
	levelString = strings.ToLower(levelString)
	for i := 0; i < len(levels); i++ {
		if levelString == strings.ToLower(levels[i]) {
			return Level(i + 1), nil
		}
	}
	return -1, fmt.Errorf("String does not represent a logging level")
}
