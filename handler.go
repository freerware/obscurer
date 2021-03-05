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
)

var (
	// ErrFailedRemoval represents an error that occur
	ErrFailedRemoval = errors.New("obscurer: unable to remove URL form store")
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
	rw := &responseWriter{ResponseWriter: w, autoFlush: true}
	h.handler.ServeHTTP(rw, r)

	// remove entries for resources that don't exist.
	if rw.status == 404 {
		if err := h.store.Remove(r.URL); err != nil {
			http.Error(w, ErrFailedRemoval.Error(), 500)
		}
	}
}

// responseWriter is a decorator around the original http.ResponseWriter.
// this allows for our handler to determine the status code that is going
// to be returned to the client so we can act on it.
type responseWriter struct {
	http.ResponseWriter

	status    int
	autoFlush bool
}

// WriterHeader captures the status code being set for the response,
// and delegates to the underlying http.ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	if rw.autoFlush {
		rw.Flush()
	}
}

// Flush writes the status code to the underlying http.ResponseWriter.
func (rw *responseWriter) Flush() {
	rw.ResponseWriter.WriteHeader(rw.status)
}
