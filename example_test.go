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

	"github.com/freerware/obscurer"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
)

func Example_http() {
	// create your mux.
	mux := http.NewServeMux()
	mux.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("we made it!")
	})

	// add obscured URL support.
	handler := obscurer.NewHandler(obscurer.Default, obscurer.DefaultStore, mux)

	// create your server.
	server := httptest.NewServer(handler)
	defer server.Close()

	original, _ := url.Parse(fmt.Sprintf("%s/this/is/the/way", server.URL))
	obscured := obscurer.Default.Obscure(original)

	// load the store.
	store := obscurer.DefaultStore
	store.Load(map[*url.URL]*url.URL{obscured: original})

	// issue the request.
	http.Get(obscured.String())
	// Output:
	// we made it!
}

func Example_gorilla() {
	// create your mux.
	router := mux.NewRouter()
	router.HandleFunc("/this/is/the/way", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("we made it!")
	})

	// add obscured URL support.
	handler := obscurer.NewHandler(obscurer.Default, obscurer.DefaultStore, router)

	// create your server.
	server := httptest.NewServer(handler)
	defer server.Close()

	original, _ := url.Parse(fmt.Sprintf("%s/this/is/the/way", server.URL))
	obscured := obscurer.Default.Obscure(original)

	// load the store.
	store := obscurer.DefaultStore
	store.Load(map[*url.URL]*url.URL{obscured: original})

	// issue the request.
	http.Get(obscured.String())
	// Output:
	// we made it!
}

func Example_gin() {
	// create your mux.
	router := gin.New()
	router.GET("/this/is/the/way", func(ctx *gin.Context) {
		fmt.Println("we made it!")
	})

	// add obscured URL support.
	handler := obscurer.NewHandler(obscurer.Default, obscurer.DefaultStore, router)

	// create your server.
	server := httptest.NewServer(handler)
	defer server.Close()

	original, _ := url.Parse(fmt.Sprintf("%s/this/is/the/way", server.URL))
	obscured := obscurer.Default.Obscure(original)

	// load the store.
	store := obscurer.DefaultStore
	store.Load(map[*url.URL]*url.URL{obscured: original})

	// issue the request.
	http.Get(obscured.String())
	// Output:
	// we made it!
}
