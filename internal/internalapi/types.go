package internalapi

type statusResponse struct {
	Status string `json:"status"`
}

type expireServersResponse struct {
	Expired []string `json:"expired"`
}
