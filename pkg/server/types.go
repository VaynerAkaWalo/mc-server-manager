package server

type Response struct {
	Name          string `json:"name"`
	IP            string `json:"IP"`
	RemainingTime string `json:"remainingTime"`
	Status        string `json:"status"`
}

type Request struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	ExpireAfter int64  `json:"expireAfter"`
}
