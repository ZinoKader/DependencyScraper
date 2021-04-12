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

func RandomProxy() string {
	return PROXIES[rand.Intn(len(PROXIES))]
}
