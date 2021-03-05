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
	"crypto/md5"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/freerware/obscurer"
)

func TestObscure(t *testing.T) {
	// arrange.
	obscurer := obscurer.Default
	u, err := url.Parse("http://www.example.com/this/is/the/way/")
	if err != nil {
		t.Error(err.Error())
	}
	want := *u
	obscuredPathBytes := md5.New().Sum([]byte(strings.TrimLeft(u.Path, "/")))
	obscuredPath := fmt.Sprintf("%x", obscuredPathBytes)
	want.Path = "/" + obscuredPath

	// action + assert.
	got := obscurer.Obscure(u)
	if want != *got {
		t.Errorf("wanted: %s, got: %s", &want, got)
	}
}
