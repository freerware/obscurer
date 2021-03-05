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
	"net/url"
	"sync"
)

// DefaultStore represents the default store.
var DefaultStore = &memoryStore{}

// Store stores mappings between obscured URLs and their original form.
type Store interface {
	Put(obscured, original *url.URL) error
	Get(obscured *url.URL) (*url.URL, bool)
	Remove(obscured *url.URL) error
	Clear() error
	Size() int
	Load(map[*url.URL]*url.URL) error
}

// memoryStore stores all obscured URL mappings in memory.
type memoryStore struct {
	store sync.Map
}

// Put places the mapping between the provided obscured URL and it's original
// form into the store.
func (s *memoryStore) Put(obscured, original *url.URL) error {
	if _, ok := s.store.Load(obscured.Path); !ok {
		s.store.Store(obscured.Path, *original)
	}
	return nil
}

// Get retrieves the original form of the provided obscured URL.
func (s *memoryStore) Get(obscured *url.URL) (*url.URL, bool) {
	original, ok := s.store.Load(obscured.Path)
	if ok {
		originalURL := original.(url.URL)
		return &originalURL, ok
	}
	return nil, ok
}

// Remove deletes the entry in the store for the provided obscured URL.
func (s *memoryStore) Remove(obscured *url.URL) error {
	s.store.Delete(obscured.Path)
	return nil
}

// Clear removes all entries in the store.
func (s *memoryStore) Clear() error {
	s.store.Range(func(key, value interface{}) bool {
		s.store.Delete(key)
		return true
	})
	return nil
}

// Size computes the size of the store.
func (s *memoryStore) Size() (size int) {
	s.store.Range(func(key, value interface{}) bool {
		size = size + 1
		return true
	})
	return
}

// Load loads the store with the provided map, where the keys are
// obscured URLs and the values are their corresponding originals.
func (s *memoryStore) Load(mappings map[*url.URL]*url.URL) error {
	for obscured, unobscured := range mappings {
		if err := s.Put(obscured, unobscured); err != nil {
			return err
		}
	}
	return nil
}
