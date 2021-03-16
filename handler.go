/* Copyright 2021 Freerware
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package obscurer

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// headerParser parses the URL portion of a particular header value.
type headerParser func(string) string

var (
	// defaultParseHeader represents the default header parser, which
	// takes the header value as is.
	defaultParseHeader headerParser = func(header string) string { return header }

	// parseLinkHeader represents the header parser for the Link header.
	parseLinkHeader headerParser = func(header string) string {
		r := regexp.MustCompile("^<(.+)>.*")
		if !r.MatchString(header) {
			return ""
		}

		matches := r.FindStringSubmatch(header)
		return matches[1]
	}
)

var (
	// ErrFailedRemoval represents an error that occurs when removing a URL
	// mapping from the store.
	ErrFailedRemoval = errors.New("obscurer: unable to remove URL form store")
	// ErrLocationHeaderFailure represents an error that occurs when obscuring
	// the 'Location' header.
	ErrLocationHeaderFailure = errors.New("obscurer: unable to obscure 'Location' header")
	// ErrLocationHeaderFailure represents an error that occurs when obscuring
	// the 'Content-Location' header.
	ErrContentLocationHeaderFailure = errors.New("obscurer: unable to obscure 'Content-Location' header")
	// ErrLocationHeaderFailure represents an error that occurs when obscuring
	// the 'Linkj' header.
	ErrLinkHeaderFailure = errors.New("obscurer: unable to obscure 'Link' header")
)

type handler struct {
	handler  http.Handler
	obscurer Obscurer
	store    Store
}

// NewHandler constructs an HTTP handler capable of handling requests with obscured URLs.
func NewHandler(o Obscurer, s Store, h http.Handler) http.Handler {
	return &handler{handler: h, obscurer: o, store: s}
}

// ServeHTTP handles the HTTP request.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// assume incoming request is obscured.
	if unobscured, ok := h.store.Get(r.URL); ok {
		r.URL = unobscured
	}

	// handle the request.
	rw := &responseWriter{ResponseWriter: w}
	defer func() {
		if _, err := rw.Do(); err != nil {
			http.Error(rw, err.Error(), 500)
		}
	}()
	h.handler.ServeHTTP(rw, r)

	// remove entries for resources that don't exist.
	if rw.status == 404 {
		if err := h.store.Remove(r.URL); err != nil {
			http.Error(rw, ErrFailedRemoval.Error(), 500)
		}
	}

	// obscure 'Location'.
	// see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Location
	if err := h.obscureHeader(rw, "Location", defaultParseHeader); err != nil {
		http.Error(rw, ErrLocationHeaderFailure.Error(), 500)
	}

	// obscure 'Content-Location'.
	// see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Location
	if err := h.obscureHeader(rw, "Content-Location", defaultParseHeader); err != nil {
		http.Error(rw, ErrContentLocationHeaderFailure.Error(), 500)
	}

	// obscure 'Link'.
	// see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Link
	if err := h.obscureHeader(rw, "Link", parseLinkHeader); err != nil {
		http.Error(rw, ErrLinkHeaderFailure.Error(), 500)
	}
}

// obscureHeader obscures the header with the provided key using the provided
// header parser.
func (h *handler) obscureHeader(w http.ResponseWriter, key string, parse headerParser) error {
	// grab the header value.
	headers := w.Header()
	header := headers.Get(key)
	// parse the URL data from the header.
	parsedHeader := parse(header)
	if header == "" {
		return nil
	}
	url, err := url.Parse(parsedHeader)
	if err != nil {
		return err
	}
	// obscure the URL.
	obscured := h.obscurer.Obscure(url)
	if obscured != nil {
		obscuredHeader := strings.ReplaceAll(header, url.String(), obscured.String())
		headers.Set(key, obscuredHeader)
	}
	return h.store.Put(obscured, url)
}
