package server

type response struct {
	Name          string `json:"name"`
	IP            string `json:"IP"`
	RemainingTime string `json:"remainingTime"`
}

type request struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	ExpireAfter int64  `json:"expireAfter"`
}

type expiredServers struct {
	Names []string `json:"names"`
}
