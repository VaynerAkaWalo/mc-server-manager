package save

import (
	"net/http"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
)

func newTestRouter(handlers *Handlers) http.Handler {
	mux := http.NewServeMux()
	handlers.RegisterRoutes(&xhttp.Router{ServeMux: mux})
	return mux
}
