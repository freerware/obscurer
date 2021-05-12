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
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/freerware/obscurer"
	"github.com/freerware/obscurer/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_UnobscuredRequestURL tests that requests issued with an
// unobscured URL are properly handled.
func TestHandler_UnobscuredRequestURL(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctx := context.Background()
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
	require.NoError(err)
	assert.Equalf(http.StatusOK, response.StatusCode, "expected status code 200, got status code %d", response.StatusCode)
	assert.True(handled, "expected for the request to be handled")
	assert.Equalf(0, store.Size(ctx), "expected the store to be empty")

	// cleanup.
	t.Cleanup(func() {
		store.Clear(ctx)
	})
}

// TestHandler_ObscuredRequestURL tests that requests issused with an
// obscured URL are properly handled.
func TestHandler_ObscuredRequestURL(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctx := context.Background()
	handled := false
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		handled = true
	})
	store := obscurer.DefaultStore
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	u := mustParse(fmt.Sprintf("%s/this/is/the/way", server.URL))
	obscuredURL := obscurer.Default.Obscure(u)
	err := store.Load(ctx, map[*url.URL]*url.URL{
		obscuredURL: u,
	})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// action + assert.
	response, err := http.Get(obscuredURL.String())
	require.NoError(err)
	assert.Equalf(http.StatusOK, response.StatusCode, "expected status code 200, got status code %d", response.StatusCode)
	assert.True(handled, "expected for the request to be handled")
	assert.Equalf(1, store.Size(ctx), "expected the store to have one entry")
	_, ok := store.Get(ctx, obscuredURL)
	assert.True(ok, "expected the store to have entry for the obscured URL")

	// cleanup.
	t.Cleanup(func() {
		store.Clear(ctx)
	})
}

// TestHandler_404Request tests that requests issued to an unhandled URL
// results in HTTP 404.
func TestHandler_404Request(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctx := context.Background()
	mux := http.NewServeMux()
	store := obscurer.DefaultStore
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	u := mustParse(fmt.Sprintf("%s/this/is/not/the/way", server.URL))
	obscuredURL := obscurer.Default.Obscure(u)

	// action + assert.
	response, err := http.Get(obscuredURL.String())
	require.NoError(err)
	defer response.Body.Close()
	assert.Equalf(http.StatusNotFound, response.StatusCode, "expected status code 404, got status code %d", response.StatusCode)
	responseBytes, err := ioutil.ReadAll(response.Body)
	require.NoError(err)
	responseBody := string(responseBytes)
	// https://golang.org/src/net/http/server.go?s=64501:64553#L2086
	want := "404 page not found\n"
	assert.Equal(want, responseBody, "expected body to be %q, got %q", want, responseBody)
	assert.Equalf(0, store.Size(ctx), "expected the store to be empty")

	// cleanup.
	t.Cleanup(func() {
		store.Clear(ctx)
	})
}

// TestHandler_404Request_RemovalError tests that requests issued to an
// unhandled URL results in HTTP 500 when there is a failure when attempting
// to remove a mapping from the store.
func TestHandler_404Request_RemovalError(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mux := http.NewServeMux()
	store := mock.NewStore(ctrl)
	expectedErr := errors.New("whoa")
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	u := mustParse(fmt.Sprintf("%s/this/is/not/the/way", server.URL))
	obscuredURL := obscurer.Default.Obscure(u)
	store.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, false)
	store.EXPECT().Remove(gomock.Any(), gomock.Any()).Return(expectedErr)

	// action + assert.
	response, err := http.Get(obscuredURL.String())
	require.NoError(err)
	assert.Equalf(http.StatusInternalServerError, response.StatusCode, "expected status code 500, got status code %d", response.StatusCode)
	responseBytes, err := ioutil.ReadAll(response.Body)
	require.NoError(err)
	responseBody := string(responseBytes)
	want := obscurer.ErrFailedRemoval.Error() + "\n"
	assert.Equal(want, responseBody, "expected body to be %q, got %q", want, responseBody)
}

// TestHandler_LocationHeader tests that the 'Location' header is obscured.
func TestHandler_LocationHeader(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctx := context.Background()
	location := mustParse("/hey/der")
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", location.String())
		w.WriteHeader(http.StatusOK)
	})
	obscuredLocation := obscurer.Default.Obscure(location)
	store := obscurer.DefaultStore
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	require.NoError(err)
	assert.Equalf(http.StatusOK, response.StatusCode, "expected status code 200, got status code %d", response.StatusCode)
	assert.Equalf(1, store.Size(ctx), "expected the store to have one entry")
	got := response.Header.Get("Location")
	want := obscuredLocation.String()
	assert.Equal(want, got, "expected 'Location' header to be %q, not %q", obscuredLocation.String(), got)

	// cleanup.
	t.Cleanup(func() {
		store.Clear(ctx)
	})
}

// TestHandler_LocationHeader_InvalidURL tests that an HTTP 500 is returned
// when an invalid URL is provided for the 'Location' header.
func TestHandler_LocationHeader_InvalidURL(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctx := context.Background()
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", "example.com\foo")
		w.WriteHeader(http.StatusOK)
	})
	store := obscurer.DefaultStore
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	require.NoError(err)
	assert.Equalf(http.StatusInternalServerError, response.StatusCode, "expected status code 500, got status code %d", response.StatusCode)
	responseBytes, err := ioutil.ReadAll(response.Body)
	require.NoError(err)
	responseBody := string(responseBytes)
	want := obscurer.ErrLocationHeaderFailure.Error() + "\n"
	assert.Equal(want, responseBody, "expected body to be %q, got %q", want, responseBody)

	// cleanup.
	t.Cleanup(func() {
		store.Clear(ctx)
	})
}

// TestHandler_LocationHeader_PutError tests that an HTTP 500 is returned
// when an error is encountered when attempting to store the 'Location'
// header in the store.
func TestHandler_LocationHeader_PutError(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", "/hey/der")
		w.WriteHeader(http.StatusOK)
	})
	store := mock.NewStore(ctrl)
	expectedErr := errors.New("whoa")
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	store.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, false)
	store.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	require.NoError(err)
	assert.Equalf(http.StatusInternalServerError, response.StatusCode, "expected status code 500, got status code %d", response.StatusCode)
	responseBytes, err := ioutil.ReadAll(response.Body)
	require.NoError(err)
	responseBody := string(responseBytes)
	want := obscurer.ErrLocationHeaderFailure.Error() + "\n"
	assert.Equal(want, responseBody, "expected body to be %q, got %q", want, responseBody)
}

// TestHandler_ContentLocationHeader tests that the 'Content-Location'
// header is obscured.
func TestHandler_ContentLocationHeader(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctx := context.Background()
	location := mustParse("/hey/der")
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Location", location.String())
		w.WriteHeader(http.StatusOK)
	})
	obscuredLocation := obscurer.Default.Obscure(location)
	store := obscurer.DefaultStore
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	require.NoError(err)
	assert.Equalf(http.StatusOK, response.StatusCode, "expected status code 200, got status code %d", response.StatusCode)
	assert.Equalf(1, store.Size(ctx), "expected the store to have one entry")
	got := response.Header.Get("Content-Location")
	want := obscuredLocation.String()
	assert.Equal(want, got, "expected 'Content-Location' header to be %q, not %q", want, got)

	// cleanup.
	t.Cleanup(func() {
		store.Clear(ctx)
	})
}

// TestHandler_ContentLocationHeader_InvalidURL tests that an HTTP 500 is returned
// when an invalid URL is provided for the 'Content-Location' header.
func TestHandler_ContentLocationHeader_InvalidURL(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctx := context.Background()
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Location", "example.com\foo")
		w.WriteHeader(http.StatusOK)
	})
	store := obscurer.DefaultStore
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	require.NoError(err)
	assert.Equalf(http.StatusInternalServerError, response.StatusCode, "expected status code 500, got status code %d", response.StatusCode)
	responseBytes, err := ioutil.ReadAll(response.Body)
	require.NoError(err)
	responseBody := string(responseBytes)
	want := obscurer.ErrContentLocationHeaderFailure.Error() + "\n"
	assert.Equal(want, responseBody, "expected body to be %q, got %q", want, responseBody)

	// cleanup.
	t.Cleanup(func() {
		store.Clear(ctx)
	})
}

// TestHandler_ContentLocationHeader_PutError tests that an HTTP 500 is returned
// when an error is encountered when attempting to store the 'Content-Location'
// header in the store.
func TestHandler_ContentLocationHeader_PutError(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Location", "/hey/der")
		w.WriteHeader(http.StatusOK)
	})
	store := mock.NewStore(ctrl)
	expectedErr := errors.New("whoa")
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	store.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, false)
	store.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	require.NoError(err)
	assert.Equalf(http.StatusInternalServerError, response.StatusCode, "expected status code 500, got status code %d", response.StatusCode)
	responseBytes, err := ioutil.ReadAll(response.Body)
	require.NoError(err)
	responseBody := string(responseBytes)
	want := obscurer.ErrContentLocationHeaderFailure.Error() + "\n"
	assert.Equal(want, responseBody, "expected body to be %q, got %q", want, responseBody)
}

// TestHandler_LinkHeader tests that the 'Link' header is obscured.
func TestHandler_LinkHeader(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctx := context.Background()
	link := mustParse("/hey/der")
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Link", fmt.Sprintf("<%s>", link.String()))
		w.WriteHeader(http.StatusOK)
	})
	obscuredLink := obscurer.Default.Obscure(link)
	store := obscurer.DefaultStore
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	require.NoError(err)
	assert.Equalf(http.StatusOK, response.StatusCode, "expected status code 200, got status code %d", response.StatusCode)
	assert.Equalf(1, store.Size(ctx), "expected the store to have one entry")
	want := fmt.Sprintf("<%s>", obscuredLink.String())
	got := response.Header.Get("Link")
	assert.Equal(want, got, "expected 'Link' header to be %q, not %q", want, got)

	// cleanup.
	t.Cleanup(func() {
		store.Clear(ctx)
	})
}

// TestHandler_LinkHeader_InvalidURL tests that an HTTP 500 is returned
// when an invalid URL is provided for the 'Link' header.
func TestHandler_LinkHeader_InvalidURL(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctx := context.Background()
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Link", "<example.com\foo>; rel='next'")
		w.WriteHeader(http.StatusOK)
	})
	store := obscurer.DefaultStore
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	require.NoError(err)
	assert.Equalf(http.StatusInternalServerError, response.StatusCode, "expected status code 500, got status code %d", response.StatusCode)
	responseBytes, err := ioutil.ReadAll(response.Body)
	require.NoError(err)
	responseBody := string(responseBytes)
	want := obscurer.ErrLinkHeaderFailure.Error() + "\n"
	assert.Equal(want, responseBody, "expected body to be %q, got %q", want, responseBody)

	// cleanup.
	t.Cleanup(func() {
		store.Clear(ctx)
	})
}

// TestHandler_LinkHeader_PutError tests that an HTTP 500 is returned
// when an error is encountered when attempting to store the 'Link'
// header in the store.
func TestHandler_LinkHeader_PutError(t *testing.T) {
	// arrange.
	assert := assert.New(t)
	require := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Link", "</hey/der>; rel='next'")
		w.WriteHeader(http.StatusOK)
	})
	store := mock.NewStore(ctrl)
	expectedErr := errors.New("whoa")
	handler := obscurer.NewHandler(obscurer.Default, store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	store.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, false)
	store.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedErr)

	// action + assert.
	response, err := http.Get(fmt.Sprintf("%s/this/is/the/way", server.URL))
	require.NoError(err)
	assert.Equalf(http.StatusInternalServerError, response.StatusCode, "expected status code 500, got status code %d", response.StatusCode)
	responseBytes, err := ioutil.ReadAll(response.Body)
	require.NoError(err)
	responseBody := string(responseBytes)
	want := obscurer.ErrLinkHeaderFailure.Error() + "\n"
	assert.Equal(want, responseBody, "expected body to be %q, got %q", want, responseBody)
}

func mustParse(str string) *url.URL {
	u, err := url.Parse(str)
	if err != nil {
		panic(err)
	}
	return u
}
