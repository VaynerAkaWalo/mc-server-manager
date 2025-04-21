package server

import "github.com/VaynerAkaWalo/mc-server-manager/internal/definition"

type Response struct {
	Name          string `json:"name"`
	IP            string `json:"IP"`
	RemainingTime string `json:"remainingTime"`
	Status        string `json:"status"`
}

type Request struct {
	Name        string                       `json:"name"`
	ExpireAfter int64                        `json:"expireAfter"`
	OPTS        map[definition.Option]string `json:"opts"`
}
