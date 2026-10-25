package server

type expiredServers struct {
	Names []string `json:"names"`
}

type stats struct {
	Online  int `json:"online"`
	Created int `json:"created"`
}
