package ray

import (
	"bytes"
	"net/http"
)

type Responder interface {
	Apply(Response) Response
}

type ResponderFunc func(Response) Response

func (f ResponderFunc) Apply(response Response) Response {
	return f(response)
}

type Responders []Responder

func (r Responders) Apply(response Response) Response {
	for _, responder := range r {
		response = responder.Apply(response)
	}

	return response
}

func Respond(responders ...Responder) Response {
	return Responders(responders).Apply(Response{Status: http.StatusOK})
}

type recorder struct {
	status int
	header http.Header
	body   *bytes.Buffer
}

func (r *recorder) Header() http.Header {
	return r.header
}

func (r *recorder) Write(body []byte) (int, error) {
	if r.body == nil {
		r.body = new(bytes.Buffer)
	}

	return r.body.Write(body)
}

func (r *recorder) WriteHeader(statusCode int) {
	r.status = statusCode
}

func RespondHandler(r *http.Request, handler http.Handler) Response {
	recorder := new(recorder)
	handler.ServeHTTP(recorder, r)

	return Response{
		Status:  recorder.status,
		Headers: recorder.header,
		Body:    recorder.body,
	}
}
