package server

type response struct {
	Name string `json:"name"`
	IP   string `json:"IP"`
}

type request struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	ExpireAfter int64  `json:"expireAfter"`
}

type expiredServers struct {
	Names []string `json:"names"`
}
