package definition

type Option string

const (
	OPS               = "OPS"
	VERSION           = "VERSION"
	EULA              = "EULA"
	ONLINE_MODE       = "ONLINE_MODE"
	TYPE              = "TYPE"
	MOTD              = "MOTD"
	USE_AIKAR_FLAGS   = "USE_AIKAR_FLAGS"
	MAX_PLAYERS       = "MAX_PLAYERS"
	DIFFICULTY        = "DIFFICULTY"
	MODRINTH_PROJECTS = "MODRINTH_PROJECTS"
	MODS              = "MODS"
	MEMORY            = "MEMORY"
	JVM_XX_OPTS       = "JVM_XX_OPTS"
	VIEW_DISTANCE     = "VIEW_DISTANCE"
)

var optsKeys = map[Option]string{
	OPS:               "OPS",
	VERSION:           "VERSION",
	EULA:              "EULA",
	ONLINE_MODE:       "ONLINE_MODE",
	TYPE:              "TYPE",
	MOTD:              "MOTD",
	USE_AIKAR_FLAGS:   "USE_AIKAR_FLAGS",
	MAX_PLAYERS:       "MAX_PLAYERS",
	DIFFICULTY:        "DIFFICULTY",
	MODRINTH_PROJECTS: "MODRINTH_PROJECTS",
	MODS:              "MODS",
	MEMORY:            "MEMORY",
	JVM_XX_OPTS:       "JVM_XX_OPTS",
	VIEW_DISTANCE:     "VIEW_DISTANCE",
}

var RequiredOpts = map[Option]string{
	EULA:            "true",
	USE_AIKAR_FLAGS: "true",
	DIFFICULTY:      "2",
	JVM_XX_OPTS:     "-XX:MaxRAMPercentage=75",
	MEMORY:          "",
	VIEW_DISTANCE:   "20",
}

var DefaultQuota = ResourceQuota{
	CpuRequest: 3,
	CpuLimit:   4,
	MemoryInMb: 7000,
}

type ResourceQuota struct {
	CpuRequest int
	CpuLimit   int
	MemoryInMb int
}

const defaultImage = "itzg/minecraft-server:latest"

type ServerDefinition struct {
	Name        string            `json:"name"`
	Options     map[Option]string `json:"options"`
	Quota       ResourceQuota     `json:"quota"`
	ExpireAfter int64             `json:"expireAfter"`
}
