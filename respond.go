package ray

import "net/http"

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
