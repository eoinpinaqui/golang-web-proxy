package PerformanceData

import (
"fmt"
"net/url"
"sync"
"time"
)

// The PerformanceData struct keeps track of how the proxy is performing
type PerformanceData struct {
	times      map[url.URL]*TimingData
	bandwidths map[url.URL]*BandWidthData
	mu         sync.Mutex
}

// The TimingData struct keeps track of the response times for the proxy
type TimingData struct {
	uncachedTimes []time.Duration
	cachedTimes   []time.Duration
}

// The BandwidthData struct keeps track of the bandwidth data for the proxy
type BandWidthData struct {
	uncachedTimes []float64
	cachedTimes   []float64
}

// This function creates a new PerformanceData struct
func New() *PerformanceData {
	times := make(map[url.URL]*TimingData)
	bandwidths := make(map[url.URL]*BandWidthData)
	return &PerformanceData{
		times:      times,
		bandwidths: bandwidths,
		mu:         sync.Mutex{},
	}
}

// This function adds a site to the PerformanceData struct
func (p *PerformanceData) AddSite(site url.URL) {
	p.mu.Lock()
	p.times[site] = &TimingData{
		uncachedTimes: []time.Duration{},
		cachedTimes:   []time.Duration{},
	}
	p.bandwidths[site] = &BandWidthData{
		uncachedTimes: []float64{},
		cachedTimes:   []float64{},
	}
	p.mu.Unlock()
}

// This function adds an uncached time to the PerformanceData struct
func (p *PerformanceData) AddUncachedTime(site url.URL, responseTime time.Duration, contentLength int64) {
	p.mu.Lock()
	if responseTimes, exists := p.times[site]; exists {
		responseTimes.uncachedTimes = append(responseTimes.uncachedTimes, responseTime)
	} else {
		p.times[site] = &TimingData{
			uncachedTimes: []time.Duration{responseTime},
			cachedTimes:   []time.Duration{},
		}
	}
	if bandwidthTimes, exists := p.bandwidths[site]; exists {
		bandwidthTimes.uncachedTimes = append(bandwidthTimes.uncachedTimes, float64(contentLength/1000)/(float64(responseTime.Milliseconds())/float64(time.Second.Milliseconds())))
	} else {
		p.bandwidths[site] = &BandWidthData{
			uncachedTimes: []float64{float64(contentLength/1000) / (float64(responseTime.Milliseconds()) / float64(time.Second.Milliseconds()))},
			cachedTimes:   []float64{},
		}
	}
	p.mu.Unlock()
}

// This function adds a cached time to the PerformanceData struct
func (p *PerformanceData) AddCachedTime(site url.URL, responseTime time.Duration, contentLength int64) {
	p.mu.Lock()
	if responseTimes, exists := p.times[site]; exists {
		responseTimes.cachedTimes = append(responseTimes.cachedTimes, responseTime)
	} else {
		p.times[site] = &TimingData{
			uncachedTimes: []time.Duration{},
			cachedTimes:   []time.Duration{responseTime},
		}
	}
	if bandwidthTimes, exists := p.bandwidths[site]; exists {
		bandwidthTimes.cachedTimes = append(bandwidthTimes.cachedTimes, float64(contentLength/1000)/(float64(responseTime.Milliseconds())/float64(time.Second.Milliseconds())))
	} else {
		p.bandwidths[site] = &BandWidthData{
			uncachedTimes: []float64{},
			cachedTimes:   []float64{float64(contentLength/1000) / (float64(responseTime.Milliseconds()) / float64(time.Second.Milliseconds()))},
		}
	}
	p.mu.Unlock()
}

// This function returns the average response times in the PerformanceData struct in milliseconds
func (p *PerformanceData) GetAverageTimes() [][]string {
	p.mu.Lock()
	var data [][]string
	for k, v := range p.times {
		// Calculate the average response times
		uncachedAvg := int64(0)
		cachedAvg := int64(0)

		for i := 0; i < len(v.uncachedTimes); i++ {
			uncachedAvg += v.uncachedTimes[i].Milliseconds()
		}
		for i := 0; i < len(v.cachedTimes); i++ {
			cachedAvg += v.cachedTimes[i].Milliseconds()
		}
		if len(v.uncachedTimes) > 0 {
			uncachedAvg /= int64(len(v.uncachedTimes))
		} else {
			uncachedAvg = -1
		}
		if len(v.cachedTimes) > 0 {
			cachedAvg /= int64(len(v.cachedTimes))
		} else {
			cachedAvg = -1
		}
		// Add the response times as a row of the table to print
		if cachedAvg > 0 {
			data = append(data, []string{k.Host, fmt.Sprintf("%dms", uncachedAvg), fmt.Sprintf("%dms", cachedAvg)})
		} else {
			data = append(data, []string{k.Host, fmt.Sprintf("%dms", uncachedAvg), fmt.Sprintf("unused")})
		}
	}
	p.mu.Unlock()
	return data
}

// This function returns the average bandwidth values in the PerformanceData struct
func (p *PerformanceData) GetAverageBandwidths() [][]string {
	p.mu.Lock()
	var data [][]string
	for k, v := range p.bandwidths {
		// Calculate the average response times
		uncachedAvg := float64(0)
		cachedAvg := float64(0)

		for i := 0; i < len(v.uncachedTimes); i++ {
			uncachedAvg += v.uncachedTimes[i]
		}
		for i := 0; i < len(v.cachedTimes); i++ {
			cachedAvg += v.cachedTimes[i]
		}
		if len(v.uncachedTimes) > 0 {
			uncachedAvg /= float64(len(v.uncachedTimes))
		} else {
			uncachedAvg = -1
		}
		if len(v.cachedTimes) > 0 {
			cachedAvg /= float64(len(v.cachedTimes))
		} else {
			cachedAvg = -1
		}
		// Add the response times as a row of the table to print
		if cachedAvg > 0 {
			data = append(data, []string{k.Host, fmt.Sprintf("%fMb/s", uncachedAvg), fmt.Sprintf("%fMb/s", cachedAvg)})
		} else {
			data = append(data, []string{k.Host, fmt.Sprintf("%fMb/s", uncachedAvg), fmt.Sprintf("unused")})

		}
	}
	p.mu.Unlock()
	return data
}
