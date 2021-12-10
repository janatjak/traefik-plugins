package cookie

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

type Config struct {
	Name     string
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

func CreateConfig() *Config {
	return &Config{}
}

type CookieConfig struct {
	next   http.Handler
	name   string
	config Config
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}

	return &CookieConfig{
		next:   next,
		name:   name,
		config: *config,
	}, nil
}

func (a *CookieConfig) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	config := a.config

	_, notFound := req.Cookie(config.Name)
	if notFound == nil {
		a.next.ServeHTTP(rw, req)
		return
	}

	value := uuidV4()
	cookie := &http.Cookie{
		Name:     config.Name,
		Value:    value,
		MaxAge:   config.MaxAge,
		Secure:   config.Secure,
		HttpOnly: config.HttpOnly,
		Path:     "/",
	}

	req.AddCookie(cookie)
	http.SetCookie(rw, cookie)

	a.next.ServeHTTP(rw, req)
}

// internal - copy from google uuid

type UUID [16]byte

func (uuid UUID) String() string {
	var buf [36]byte
	encodeHex(buf[:], uuid)
	return string(buf[:])
}

func encodeHex(dst []byte, uuid UUID) {
	hex.Encode(dst, uuid[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], uuid[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], uuid[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], uuid[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], uuid[10:])
}

func uuidV4() string {
	var uuid UUID
	io.ReadFull(rand.Reader, uuid[:])

	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	// uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return uuid.String()
}
