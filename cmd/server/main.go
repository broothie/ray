package main

import (
	"encoding/json"
	"fmt"
	"github.com/broothie/ray"
	"net/http"
	"net/url"
	"time"
)

func main() {
	http.Handle("/jsonify", Logger(func(r *http.Request) ray.Response {
		return ray.Respond(
			ray.Header("trace-iD", "asdf"),
			ray.BodyFile("go.mod"),
		)
	}))

	http.Handle("/queryify", Logger(func(r *http.Request) ray.Response {
		var body url.Values
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return ray.Respond(
				ray.Status(http.StatusBadRequest),
				ray.BodyError(err),
			)
		}

		return ray.Respond(ray.BodyQuery(body))
	}))

	http.ListenAndServe(":8888", nil)
}

func Logger(next ray.Handler) ray.Handler {
	return func(r *http.Request) ray.Response {
		start := time.Now()
		response := next(r)
		elapsed := time.Since(start)

		fmt.Printf("%v %s %s | %d | %v\n", start, r.Method, r.URL.Path, response.Status, elapsed)
		return response
	}
}
