package pdf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Config struct {
	Url string
}

func CreateConfig() *Config {
	return &Config{}
}

type Pdf struct {
	next http.Handler
	name string
	url  string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.Url == "" {
		return nil, fmt.Errorf("url is empty")
	}

	return &Pdf{
		next: next,
		name: name,
		url:  config.Url,
	}, nil
}

func (a *Pdf) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	if !query.Has("generate_pdf") {
		a.next.ServeHTTP(rw, req)
		return
	}

	bufferWriter := &BufferResponseWriter{
		header: make(map[string][]string, 0),
	}

	fmt.Println("🍀 START !!!")
	a.next.ServeHTTP(bufferWriter, req)
	fmt.Println("🍀 FINISH !!!")

	convert(rw, a.url, query, bufferWriter)
}

// custom part

type H map[string]interface{}

type BufferResponseWriter struct {
	header http.Header
	buf    bytes.Buffer
	code   int
}

func (w *BufferResponseWriter) Header() http.Header {
	return w.header
}

func (w *BufferResponseWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *BufferResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

// query.content_disposition: inline | attachment (default is inline)
// query.filename (default is file.pdf)
func convert(rw http.ResponseWriter, url string, query url.Values, bufferRw *BufferResponseWriter) {
	filename := query.Get("filename")
	if filename == "" {
		filename = "file.pdf"
	}

	contentDisposition := "inline"
	if query.Get("content_disposition") == "attachment" {
		contentDisposition = "attachment"
	}

	body, _ := json.Marshal(H{
		"input": H{
			"type":    "html",
			"content": bufferRw.buf.String(),
		},
		"output": H{
			"type":        "pdf",
			"disposition": "inline",
			"filename":    filename,
			"options": H{
				"printBackground": true,
			},
		},
	})

	res, _ := http.Post(url, "application/json", bytes.NewBuffer(body))

	rw.Header().Add("Content-Type", "application/pdf")
	rw.Header().Add("Content-Disposition", contentDisposition+"; filename=\""+filename+"\"")

	defer res.Body.Close()

	io.Copy(rw, res.Body)
}
