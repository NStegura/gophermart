package clients

import "net/http"

type HTTPClient interface {
	Get(reqURL string) (*http.Response, error)
}
