package scraping

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/ZinoKader/KEX/pkg/data"
)

var PROXIES = data.ProxyList()

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}


func CreateRequest (URL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return &http.Request{}, nil
	} 
	req.Close = true
	return req, nil
}

func RandomProxy() string {
	return PROXIES[rand.Intn(len(PROXIES))]
}
