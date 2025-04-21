package definition

type Option int

const (
	OPS = iota
	EULA
	ONLINE_MODE
	TYPE
	MODT
	USE_AIKAR_FLAGS
	MAX_PLAYERS
	DIFFICULTY
	SIMULATION_DISTANCE
	VIEW_DISTANCE
)

var OPTS = map[Option]string{
	OPS:                 "OPS",
	EULA:                "EULA",
	ONLINE_MODE:         "ONLINE_MODE",
	TYPE:                "TYPE",
	MODT:                "MODT",
	USE_AIKAR_FLAGS:     "USE_AIKAR_FLAGS",
	MAX_PLAYERS:         "MAX_PLAYERS",
	DIFFICULTY:          "DIFFICULTY",
	SIMULATION_DISTANCE: "SIMULATION_DISTANCE",
	VIEW_DISTANCE:       "VIEW_DISTANCE",
}

type ServerDefinition struct {
	Image   string
	Options []Option
}
