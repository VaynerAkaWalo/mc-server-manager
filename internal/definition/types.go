package definition

type Option int

const (
	OPS = iota
	VERSION
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

var optsKeys = map[Option]string{
	OPS:                 "OPS",
	VERSION:             "VERSION",
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

var requiredOpts = map[Option]string{
	EULA:            "true",
	USE_AIKAR_FLAGS: "true",
	DIFFICULTY:      "2",
}

var DefaultQuota = ResourceQuota{
	CpuRequest: 3,
	CpuLimit:   4,
	MemoryInMb: 7500,
}

type ResourceQuota struct {
	CpuRequest int
	CpuLimit   int
	MemoryInMb int
}

const defaultImage = "itzg/minecraft-server:latest"

type ServerDefinition struct {
	Name        string
	Options     map[Option]string
	Quota       ResourceQuota
	ExpireAfter int64
}
