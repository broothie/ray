package ray

import (
	"io"
	"net/http"
)

type Response struct {
	Status  int
	Headers http.Header
	Body    io.WriterTo
	Error   error
}

type Handler func(r *http.Request) Response

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := h(r)
	for key, values := range response.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	status := response.Status
	if status == 0 {
		status = http.StatusOK
	}

	w.WriteHeader(status)
	if response.Body == nil {
		return
	}

	if _, err := response.Body.WriteTo(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
