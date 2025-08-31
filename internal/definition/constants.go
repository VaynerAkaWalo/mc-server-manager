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

const defaultImage = "itzg/minecraft-server:latest"

type Tier string

const (
	Wooden  Tier = "wooden"
	Iron    Tier = "iron"
	Diamond Tier = "diamond"
)
