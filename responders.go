package ray

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
)

func Error(err error) ResponderFunc {
	return func(response Response) Response {
		return Response{
			Status:  response.Status,
			Headers: cloneHeaders(response.Headers),
			Body:    response.Body,
			Error:   err,
		}
	}
}

func Status(status int) ResponderFunc {
	return func(response Response) Response {
		return Response{
			Status:  status,
			Headers: cloneHeaders(response.Headers),
			Body:    response.Body,
			Error:   response.Error,
		}
	}
}

func Header(key, value string) Headers {
	return Headers(http.Header{key: []string{value}})
}

type Headers http.Header

func (h Headers) Apply(response Response) Response {
	return Response{
		Status:  response.Status,
		Headers: mergeHeaders(response.Headers, http.Header(h)),
		Body:    response.Body,
		Error:   response.Error,
	}
}

func Body(body io.WriterTo) ResponderFunc {
	return func(response Response) Response {
		return Response{
			Status:  response.Status,
			Headers: cloneHeaders(response.Headers),
			Body:    body,
			Error:   response.Error,
		}
	}
}

func BodyBytes(body []byte) Responder {
	return Body(bytes.NewBuffer(body))
}

func BodyString(body string) Responder {
	return BodyBytes([]byte(body))
}

func BodyError(err error) Responder {
	return BodyString(err.Error())
}

func BodyPlain(body string) Responders {
	return Responders{
		Header("Content-Type", "text/plain"),
		BodyString(body),
	}
}

type BodyQuery url.Values

func (b BodyQuery) Apply(response Response) Response {
	return Responders{
		Header("Content-Type", "application/x-www-form-urlencoded"),
		BodyString(url.Values(b).Encode()),
	}.Apply(response)
}

func BodyReader(r io.Reader) Responder {
	return Body(writerToFunc(func(w io.Writer) (int64, error) {
		return io.Copy(w, r)
	}))
}

func BodyFile(filename string) Responder {
	return Body(writerToFunc(func(w io.Writer) (int64, error) {
		file, err := os.Open(filename)
		if err != nil {
			return 0, err
		}

		return io.Copy(w, file)
	}))
}

func BodyJSON(body any) Responders {
	return Responders{
		Header("Content-Type", "application/json"),
		Body(writerToFunc(func(w io.Writer) (int64, error) {
			cw := newCountWriter(w)
			if err := json.NewEncoder(cw).Encode(body); err != nil {
				return 0, err
			}

			return cw.n, nil
		})),
	}
}

func BodyXML(body any) Responders {
	return Responders{
		Header("Content-Type", "application/xml"),
		Body(writerToFunc(func(w io.Writer) (int64, error) {
			cw := newCountWriter(w)
			if err := xml.NewEncoder(cw).Encode(body); err != nil {
				return 0, err
			}

			return cw.n, nil
		})),
	}
}

func BodyHTMLTemplate(template *template.Template, data any) Responders {
	return Responders{
		Header("Content-Type", "text/html"),
		Body(writerToFunc(func(w io.Writer) (int64, error) {
			cw := &countWriter{w: w}
			if err := template.Execute(cw, data); err != nil {
				return 0, err
			}

			return cw.n, nil
		})),
	}
}

func cloneHeaders(headers http.Header) http.Header {
	return mergeHeaders(headers)
}

func mergeHeaders(headers ...http.Header) http.Header {
	output := make(http.Header)
	for _, header := range headers {
		for key, values := range header {
			for _, value := range values {
				output.Add(key, value)
			}
		}
	}

	return output
}

func newCountWriter(w io.Writer) *countWriter {
	return &countWriter{w: w}
}

type countWriter struct {
	w io.Writer
	n int64
}

func (c *countWriter) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	c.n += int64(n)
	return
}

type writerToFunc func(w io.Writer) (int64, error)

func (f writerToFunc) WriteTo(w io.Writer) (int64, error) {
	return f(w)
}
