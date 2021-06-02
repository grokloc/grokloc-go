// Package env contains environment-designating symbols and functions
package env

import "errors"

// Level is an integer representing a run env
type Level int

// env levels
const (
	Unit  = Level(0)
	Dev   = Level(1)
	Stage = Level(2)
	Prod  = Level(3)
)

// NewLevel evaluates env var strings to Level instances
func NewLevel(level string) (Level, error) {
	switch level {
	case "UNIT":
		return Unit, nil
	case "DEV":
		return Dev, nil
	case "STAGE":
		return Stage, nil
	case "PROD":
		return Prod, nil
	default:
		return -1, errors.New("unknown level")
	}
}
