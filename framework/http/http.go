package http

import (
	"github.com/go-resty/resty/v2"
	"time"
)

var HttpClient = resty.New().
	SetTimeout(5*time.Second).
	SetHeader("Content-Type", "application/json")

func Test() {
}
