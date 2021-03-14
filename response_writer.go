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

import "net/http"

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
