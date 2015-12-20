
package auth

import (
	"crypto/sha1"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// Set of known-to-be-valid {password, bcryt-hash} pairs.
// Keys are of the form SHA1 digest of password + bcrypt'ed hash of password
var cachedHashes = map[string]struct{}{}
var cacheLock sync.Mutex

// The maximum number of pairs to keep in the above cache
const kMaxCacheSize = 10000

// Optimized wrapper around bcrypt.CompareHashAndPassword that caches successful results in
// memory to avoid the _very_ high overhead of calling bcrypt.
func compareHashAndPassword(hash []byte, password []byte) bool {
	// Actually we cache the SHA1 digest of the password to avoid keeping passwords in RAM.
	s := sha1.New()
	s.Write(password)
	digest := string(s.Sum(nil))
	key := digest + string(hash)

	cacheLock.Lock()
	_, valid := cachedHashes[key]
	cacheLock.Unlock()
	if valid {
		return true
	}

	// Cache missed; now we make the very slow (~100ms) bcrypt call:
	if err := bcrypt.CompareHashAndPassword(hash, password); err != nil {
		// Note: It's important to only cache successful matches, not failures.
		// Failure is supposed to be slow, to make online attacks impractical.
		return false
	}

	cacheLock.Lock()
	if len(cachedHashes) >= kMaxCacheSize {
		cachedHashes = map[string]struct{}{}
	}
	cachedHashes[key] = struct{}{}
	cacheLock.Unlock()
	return true
}