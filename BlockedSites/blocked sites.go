package BlockedSites

import (
	"errors"
	"fmt"
	"github.com/golang-collections/collections/set"
	"net/url"
	"strings"
	"sync"
)

// The BlockedSites struct keeps track of what sites have been blocked by the administrator
type BlockedSites struct {
	sites set.Set
	mu    sync.Mutex
}

// This function creates a new BlockedSites struct
func New() *BlockedSites {
	return &BlockedSites{
		sites: *set.New(),
		mu:    sync.Mutex{},
	}
}

// This function adds a site to the BlockedSites struct
func (b *BlockedSites) Add(host string) error {
	b.mu.Lock()
	if b.sites.Has(host) {
		b.mu.Unlock()
		return errors.New(fmt.Sprintf("[BlockedSites] %s is already blocked", host))
	}
	b.sites.Insert(host)
	b.mu.Unlock()
	return nil
}

// This function removes a site from the BlockedSites struct
func (b *BlockedSites) Remove(host string) error {
	b.mu.Lock()
	if b.sites.Has(host) {
		b.sites.Remove(host)
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()
	return errors.New(fmt.Sprintf("[BlockedSites] %s is already unblocked", host))
}

// This function returns whether a url should in the BlockedSites struct
func (b *BlockedSites) IsBlocked(site *url.URL) bool {
	// Extract the host and remove any port present
	host := site.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	if strings.Contains(host, ".") {
		splitUrl := strings.Split(host, ".")
		host = splitUrl[len(splitUrl)-2] + "." + splitUrl[len(splitUrl)-1]
	}
	return b.sites.Has(host)
}

// This function returns a list of all the blocked sites
func (b *BlockedSites) List() []interface{} {
	var list []interface{}
	b.sites.Do(func(i interface{}) {
		list = append(list, i)
	})
	return list
}
