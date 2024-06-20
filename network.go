package manifesto

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	headerContentType = "Content-Type"
	headerUserAgent   = "User-Agent"

	mimeJson = "application/json"
	mimeText = "plain/text"
)

// HandleManifest adds a POST endpoint for the given path and spec/status types.
func HandleManifest(path string, pool *Pool, spec any, status any) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("this endpoint only allows POST requests"))
			return
		}

		manifest := ParseReader(r.Body, spec, status)
		pool.Apply(*manifest)

		b, err := json.Marshal(manifest)
		if err != nil {
			w.Header().Add(headerContentType, mimeText)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		} else {
			w.Header().Add(headerContentType, mimeJson)
			w.WriteHeader(http.StatusOK)
			w.Write(b)
			return
		}
	})
}

// SendManifest sends a manifest to a given endpoint (blocking) and returns the new manifest.
func SendManifest(endpoint string, manifest *Manifest, userAgent string) (*Manifest, error) {
	b, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint+"/"+manifest.ApiVersion+"/"+manifest.Kind, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set(headerContentType, mimeJson)
	req.Header.Set(headerUserAgent, userAgent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	return ParseReader(res.Body, manifest.Spec, manifest.Status), nil
}
