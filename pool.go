package manifesto

import (
	"fmt"
	"sync"
)

// Action tells what happened to a manifest when it is passed to listeners.
type Action int8

const (
	Created Action = iota
	Updated
	Deleted
)

// Listener is a function that is called when a manifest has been changed.
type Listener func(*Pool, Action, Manifest)

// Pool holds all manifests and listeners.
type Pool struct {
	manifests map[ManifestKey]Manifest
	listeners []Listener
	wg        sync.WaitGroup
}

// CreatePool creates an empty Pool.
func CreatePool() *Pool {
	return &Pool{
		manifests: make(map[ManifestKey]Manifest),
		listeners: make([]Listener, 0),
	}
}

// Listen add a listener to the pool.
func (pool *Pool) Listen(listener Listener) {
	pool.listeners = append(pool.listeners, listener)
}

// Apply adds or updates the manifest to or in the pool and calls all listeners.
// The manifest is transferred as value, not as reference. By doing so, we
// prevent race conditions.
func (pool *Pool) Apply(manifest Manifest) {
	key := manifest.CreateKey()

	if _, ok := pool.manifests[*key]; ok {
		pool.manifests[*key] = manifest
		for _, listener := range pool.listeners {
			pool.apply(listener, Updated, &manifest)
		}
	} else {
		pool.manifests[*key] = manifest
		for _, listener := range pool.listeners {
			pool.apply(listener, Created, &manifest)
		}
	}
}

// ApplyPartial adds or updates the manifest to or in the pool and calls all
// listeners except the specified one. This is meant to be used when a listener
// changes a manifest and should not be called again for that change (that could
// result in an endless loop). The manifest is transferred as value, not as
// reference. By doing so, we prevent race conditions.
func (pool *Pool) ApplyPartial(except Listener, manifest Manifest) {
	key := manifest.CreateKey()
	exceptName := fmt.Sprintf("%v", except)

	if _, ok := pool.manifests[*key]; ok {
		pool.manifests[*key] = manifest
		for _, listener := range pool.listeners {
			if fmt.Sprintf("%v", listener) != exceptName {
				pool.apply(listener, Updated, &manifest)
			}
		}
	} else {
		pool.manifests[*key] = manifest
		for _, listener := range pool.listeners {
			pool.apply(listener, Created, &manifest)
		}
	}
}

// ApplySilent adds or updates the manifest to or in the pool WITHOUT calling
// the listeners. This function is especially useful when using Manifesto just
// as a simple database without listeners.
func (pool *Pool) ApplySilent(manifest Manifest) {
	pool.manifests[*manifest.CreateKey()] = manifest
}

func (pool *Pool) apply(listener Listener, action Action, manifest *Manifest) {
	pool.wg.Add(1)
	go func(listener Listener) {
		defer pool.wg.Done()
		listener(pool, action, *manifest)
	}(listener)
}

// Delete deletes a manifest from the pool.
func (pool *Pool) Delete(key *ManifestKey) {
	if manifest, ok := pool.manifests[*key]; ok {
		delete(pool.manifests, *key)
		for _, listener := range pool.listeners {
			pool.wg.Add(1)
			go func(listener Listener) {
				defer pool.wg.Done()
				listener(pool, Deleted, manifest)
			}(listener)
		}
	}
}

// GetByKey searches for a manifest and returns it.
func (pool *Pool) GetByKey(key *ManifestKey) (Manifest, bool) {
	manifest, ok := pool.manifests[*key]
	return manifest, ok
}

// Find goes through all existing manifests and filters for a testing function.
func (pool *Pool) Find(test func(Manifest) bool) []Manifest {
	manifests := make([]Manifest, 0)
	for _, manifest := range pool.manifests {
		if test(manifest) {
			manifests = append(manifests, manifest)
		}
	}
	return manifests
}

// Waits till all listeners have completed their work.
func (pool *Pool) Wait() {
	pool.wg.Wait()
}
