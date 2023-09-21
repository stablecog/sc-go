package store

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
)

// A singleton that caches the features available to free users
// Avoids having to query the database every time a user requests the features
type Cache struct {
	authClients []*ent.AuthClient
	sync.RWMutex
}

var lock = &sync.Mutex{}
var singleCache *Cache

func newCache() *Cache {
	return &Cache{}
}

func GetCache() *Cache {
	if singleCache == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleCache == nil {
			singleCache = newCache()
		}
	}
	return singleCache
}

func (f *Cache) UpdateAuthClients(clients []*ent.AuthClient) {
	f.Lock()
	defer f.Unlock()
	f.authClients = clients
}

func (f *Cache) AuthClients() []*ent.AuthClient {
	f.RLock()
	defer f.RUnlock()
	return f.authClients
}

func (f *Cache) IsValidClientID(clientId string) (*ent.AuthClient, error) {
	uuidParsed, err := uuid.Parse(clientId)
	if err != nil {
		return nil, err
	}
	f.RLock()
	defer f.RUnlock()
	for _, client := range f.authClients {
		if client.ID == uuidParsed {
			return client, nil
		}
	}
	return nil, fmt.Errorf("not_found")
}
