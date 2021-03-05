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

// TestHandler_UnobscuredRequestURL tests that requests issued with an
// unobscured URL are properly handled.
func TestHandler_UnobscuredRequestURL(t *testing.T) {
	// arrange.
	handled := false
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		handled = true
	})
	store := obscurer.DefaultStore
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
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
	if !handled {
		t.Error("expected for the request to be handled")
	}
	if store.Size() > 0 {
		t.Error("expected the store to be empty")
	}

	// cleanup.
	t.Cleanup(func() {
		store.Clear()
	})
}

// TestHandler_ObscuredRequestURL tests that requests issused with an
// obscured URL are properly handled.
func TestHandler_ObscuredRequestURL(t *testing.T) {
	// arrange.
	handled := false
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		handled = true
	})
	store := obscurer.DefaultStore
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	u, err := url.Parse(fmt.Sprintf("%s/this/is/the/way", server.URL))
	if err != nil {
		t.Errorf("error when creating URL: %s", err.Error())
		t.FailNow()
	}
	obscuredURL := obscurer.Default.Obscure(u)
	err = store.Load(map[*url.URL]*url.URL{
		obscuredURL: u,
	})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// action + assert.
	response, err := http.Get(obscuredURL.String())
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if response.StatusCode != 200 {
		t.Errorf("expected status code 200, got status code %d", response.StatusCode)
	}
	if !handled {
		t.Error("expected for the request to be handled")
	}
	if store.Size() != 1 {
		t.Error("expected the store to have one entry")
	}
	if _, ok := store.Get(obscuredURL); !ok {
		t.Error("expected the store to have entry for the obscured URL")
	}

	// cleanup.
	t.Cleanup(func() {
		store.Clear()
	})
}

// TestHandler_ObscuredRequestURL tests that requests issused with an
// obscured URL are properly handled.
func TestHandler_404Request(t *testing.T) {
	// arrange.
	mux := http.NewServeMux()
	store := obscurer.DefaultStore
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	u, err := url.Parse(fmt.Sprintf("%s/this/is/not/the/way", server.URL))
	if err != nil {
		t.Errorf("error when creating URL: %s", err.Error())
		t.FailNow()
	}
	obscuredURL := obscurer.Default.Obscure(u)

	// action + assert.
	response, err := http.Get(obscuredURL.String())
	if err != nil {
		t.Error(err)
	}
	if response.StatusCode != 404 {
		t.Errorf("expected status code 404, got status code %d", response.StatusCode)
	}
	if store.Size() > 0 {
		t.Error("expected the store to be empty")
	}

	// cleanup.
	t.Cleanup(func() {
		store.Clear()
	})
}
