package port

import "net/http"

type Filter interface {
	Filter(w http.ResponseWriter, r *http.Request) bool
}
