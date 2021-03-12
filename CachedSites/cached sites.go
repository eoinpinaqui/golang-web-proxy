package CachedSites

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// The CachedSites struct keeps copies of recently served HTTP requests
type CachedSites struct {
	cache map[url.URL]*cachedSiteInfo
	mu    sync.Mutex
}

// The cachedSiteInfo struct holds information about a cached site
type cachedSiteInfo struct {
	resp     *http.Response
	respBody string
	cachedAt time.Time
}

// This function creates a new CachedSites struct and begins the refresh process
func New() *CachedSites {
	c := &CachedSites{
		cache: make(map[url.URL]*cachedSiteInfo),
		mu:    sync.Mutex{},
	}
	go c.keepCacheClean()
	return c
}

// This function adds a response to the cache
func (c *CachedSites) AddToCache(host url.URL, response *http.Response, body string) {
	c.mu.Lock()
	c.cache[host] = &cachedSiteInfo{
		resp:     response,
		respBody: body,
		cachedAt: time.Now(),
	}
	c.mu.Unlock()
}

// This function returns information about a cached website
func (c *CachedSites) GetFromCache(host url.URL) (bool, *http.Response, string) {
	if k, exists := c.cache[host]; exists {
		return true, k.resp, k.respBody
	}
	return false, nil, ""
}

// This function returns a string representation of all the cached sites
func (c *CachedSites) List() string {
	sites := ""
	for site := range c.cache {
		sites += site.Host + "\n"
	}
	return sites
}

// This function continuously removes out of date responses from the cache
func (c *CachedSites) keepCacheClean() {
	for {
		c.mu.Lock()
		for k, v := range c.cache {
			if time.Since(v.cachedAt) > 10*time.Second {
				delete(c.cache, k)
				log.Infof("Removed %s from the cache\n", k.Host)
			}
		}
		c.mu.Unlock()
	}
}
