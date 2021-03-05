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

package obscurer_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/freerware/obscurer"
)

func TestMiddleware_LocationHeader(t *testing.T) {
	// arrange.
	location, err := url.Parse("/hey/der")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", location.String())
		w.WriteHeader(200)
	})
	obscuredLocation := obscurer.Default.Obscure(location)
	store := obscurer.DefaultStore
	middleware := obscurer.NewMiddleware(obscurer.Default, obscurer.DefaultStore, mux)
	handler := obscurer.NewHandler(obscurer.Default, store, middleware)
	server := httptest.NewServer(handler)
	defer server.Close()

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if response.StatusCode != 200 {
		t.Errorf("expected status code 200, got status code %d", response.StatusCode)
	}
	if store.Size() == 0 {
		t.Error("expected the store to not be empty")
	}
	got := response.Header.Get("Location")
	if got != obscuredLocation.String() {
		t.Errorf("expected 'Location' header to be %q, not %q", obscuredLocation.String(), got)
	}

	// cleanup.
	t.Cleanup(func() {
		store.Clear()
	})
}

func TestMiddleware_ContentLocationHeader(t *testing.T) {
	// arrange.
	location, err := url.Parse("/hey/der")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Location", location.String())
		w.WriteHeader(200)
	})
	obscuredLocation := obscurer.Default.Obscure(location)
	store := obscurer.DefaultStore
	middleware := obscurer.NewMiddleware(obscurer.Default, obscurer.DefaultStore, mux)
	handler := obscurer.NewHandler(obscurer.Default, store, middleware)
	server := httptest.NewServer(handler)
	defer server.Close()

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if response.StatusCode != 200 {
		t.Errorf("expected status code 200, got status code %d", response.StatusCode)
	}
	if store.Size() == 0 {
		t.Error("expected the store to not be empty")
	}
	got := response.Header.Get("Content-Location")
	if got != obscuredLocation.String() {
		t.Errorf("expected 'Content-Location' header to be %q, not %q", obscuredLocation.String(), got)
	}

	// cleanup.
	t.Cleanup(func() {
		store.Clear()
	})
}

func TestMiddleware_LinkHeader(t *testing.T) {
	// arrange.
	link, err := url.Parse("/hey/der")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Link", fmt.Sprintf("<%s>", link.String()))
		w.WriteHeader(200)
	})
	obscuredLink := obscurer.Default.Obscure(link)
	store := obscurer.DefaultStore
	middleware := obscurer.NewMiddleware(obscurer.Default, obscurer.DefaultStore, mux)
	handler := obscurer.NewHandler(obscurer.Default, store, middleware)
	server := httptest.NewServer(handler)
	defer server.Close()

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if response.StatusCode != 200 {
		t.Errorf("expected status code 200, got status code %d", response.StatusCode)
	}
	if store.Size() == 0 {
		t.Error("expected the store to not be empty")
	}
	want := fmt.Sprintf("<%s>", obscuredLink.String())
	got := response.Header.Get("Link")
	if got != want {
		t.Errorf("expected 'Link' header to be %q, not %q", want, got)
	}

	// cleanup.
	t.Cleanup(func() {
		store.Clear()
	})
}
