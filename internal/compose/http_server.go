package compose

import (
	"fmt"
	"net/http"
	"time"
)

func newHTTPServer(handler http.Handler, port uint16) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           handler,
	}
}
