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
	"crypto/md5"
	"fmt"
	"hash"
	"net/url"
	"strings"
)

// Default represents the default obscurer.
var Default = &md5Obscurer{}

// Interface represents the interface an obscurer needs to abide by.
type Interface = Obscurer

// Obscurer obscures URLs.
type Obscurer interface {
	Obscure(*url.URL) *url.URL
}

// md5Obscurer obscures URLs using the MD5 hashing algorithm.
type md5Obscurer struct {
	hash hash.Hash
}

// Obscure obscures the provided URL.
func (o *md5Obscurer) Obscure(url *url.URL) *url.URL {
	var empty hash.Hash
	if o.hash == empty {
		o.hash = md5.New()
	}
	obscuredPathBytes := o.hash.Sum([]byte(strings.TrimLeft(url.Path, "/")))
	obscuredPath := fmt.Sprintf("%x", obscuredPathBytes)
	result := *url
	result.Path = "/" + obscuredPath
	return &result
}
