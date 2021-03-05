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

// Errors that can occur during middleware execution.
var (
	ErrLocationHeaderFailure        = errors.New("obscurer: unable to obscure 'Location' header")
	ErrContentLocationHeaderFailure = errors.New("obscurer: unable to obscure 'Content-Location' header")
	ErrLinkHeaderFailure            = errors.New("obscurer: unable to obscure 'Link' header")
)

// headerParser parses the URL portion of a particular header value.
type headerParser func(string) string

// middleware represents API middleware responsible for obscuring
// URLs in HTTP response headers.
type middleware struct {
	next     http.Handler
	obscurer Obscurer
	store    Store
}

// NewMiddleware constructs the obscurer middleware.
func NewMiddleware(o Obscurer, s Store, h http.Handler) http.Handler {
	return &middleware{next: h, obscurer: o, store: s}
}

// ServeHTTP handles the HTTP request.
func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw := &responseWriter{ResponseWriter: w}
	defer rw.Flush()

	// invoke the next middleware in the chain.
	m.next.ServeHTTP(rw, r)

	// obscure 'Location'.
	// see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Location
	if err := m.obscureHeader(w, "Location", defaultParseHeader); err != nil {
		http.Error(w, ErrLocationHeaderFailure.Error(), 500)
	}

	// obscure 'Content-Location'.
	// see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Location
	if err := m.obscureHeader(w, "Content-Location", defaultParseHeader); err != nil {
		http.Error(w, ErrContentLocationHeaderFailure.Error(), 500)
	}

	// obscure 'Link'.
	// see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Link
	if err := m.obscureHeader(w, "Link", parseLinkHeader); err != nil {
		http.Error(w, ErrLinkHeaderFailure.Error(), 500)
	}
}

// obscureHeader obscures the header with the provided key using the provided
// header parser.
func (m *middleware) obscureHeader(w http.ResponseWriter, key string, parse headerParser) error {
	// grab the header value.
	h := w.Header()
	header := h.Get(key)
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
	obscured := m.obscurer.Obscure(url)
	if obscured != nil {
		obscuredHeader := strings.ReplaceAll(header, url.String(), obscured.String())
		h.Set(key, obscuredHeader)
	}
	return m.store.Put(obscured, url)
}
