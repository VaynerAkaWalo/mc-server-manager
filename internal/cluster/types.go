package cluster

type ServerRequest struct {
	Name        string
	Image       string
	Env         map[string]string
	ExpireAfter int64
}
