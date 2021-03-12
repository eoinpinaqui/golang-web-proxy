package Proxy

import (
	blocked "../BlockedSites"
	cached "../CachedSites"
	performance "../PerformanceData"
	"fmt"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type Proxy struct {
	httpServer      *http.Server
	BlockedSites    *blocked.BlockedSites
	PerformanceData *performance.PerformanceData
	CachedSites     *cached.CachedSites
}

// This function creates a new Proxy struct
func New(port string) *Proxy {
	proxy := &Proxy{
		BlockedSites:    blocked.New(),
		PerformanceData: performance.New(),
		CachedSites:     cached.New(),
	}
	proxy.httpServer = &http.Server{
		Addr:    port,
		Handler: http.HandlerFunc(proxy.requestHandler),
	}
	return proxy
}

// This function starts the Proxy
func (p *Proxy) Start() {
	log.Infof("Listening on %s...\n\n", p.httpServer.Addr)
	err := p.httpServer.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}
}

// This function handles http and connection requests that are sent to the Proxy
func (p *Proxy) requestHandler(w http.ResponseWriter, req *http.Request) {
	if p.BlockedSites.IsBlocked(req.URL) {
		p.serveBlockedSite(w, req)
		return
	}

	// Keep track of how long the response takes
	startTime := time.Now()
	cached := false
	contentLength := int64(-1)

	// Perform the request
	if exists, x, y := p.CachedSites.GetFromCache(*req.URL); exists {
		contentLength = p.handleCachedSite(w, req, x, y)
		fmt.Println(" ")
		log.Infof("HTTP Request Received:\n%s", stringifyRequest(req))
		log.Infof("Used cached response to serve %v\n", req.URL)
		cached = true
	} else {
		req.RequestURI = ""
		if req.Method == http.MethodConnect {
			handleConnection(w, req)
		} else {
			contentLength = p.handleHTTP(w, req)
		}
	}

	// Add the response time to the PerformanceData struct
	elapsedTime := time.Since(startTime)
	if cached && contentLength > 0 {
		p.PerformanceData.AddCachedTime(*req.URL, elapsedTime, contentLength)
	} else if contentLength > 0 {
		p.PerformanceData.AddUncachedTime(*req.URL, elapsedTime, contentLength)
	}
}

// This function serves responses from the cache instead of performing a http request
func (p *Proxy) handleCachedSite(w http.ResponseWriter, req *http.Request, resp *http.Response, body string) int64 {
	size := int64(0)

	// Copy the headers over to the Response Writer
	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
			size += int64(len(v)) + int64(len(key))
		}
	}
	w.WriteHeader(resp.StatusCode)
	// Write the response to the web client
	_, err := io.WriteString(w, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
		return -1
	}

	// Send the response to the browser if user agent is curl
	agent := req.Header.Get("user-agent")
	if strings.Contains(agent, "curl") {
		err = browser.OpenReader(strings.NewReader(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err)
			return -1
		}
	}

	return int64(len(body)) + size
}

// This function handles connections between the client and the desired server
func handleConnection(w http.ResponseWriter, req *http.Request) {
	// Log information about the request
	fmt.Println(" ")
	log.Infof("Connection Request Received:\n%s\n", stringifyRequest(req))

	// Create a connection to the host specified in the request
	connection, err := net.Dial("tcp", req.URL.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Error(err)
		return
	}
	w.WriteHeader(http.StatusOK)

	// Create a hijacker for the web client
	h, created := w.(http.Hijacker)
	if !created {
		http.Error(w, "Hijacking not supported", http.StatusServiceUnavailable)
		log.Error("Hijacking not supported\n")
		return
	}
	client, _, err := h.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		log.Error(err)
		return
	}

	// Turn the connections into TCP connections and transfer the data
	dest, destCreated := connection.(*net.TCPConn)
	src, srcCreated := client.(*net.TCPConn)
	if destCreated && srcCreated {
		go transferData(dest, src)
		go transferData(src, dest)
	} else {
		http.Error(w, "Error creating tcp connection", http.StatusInternalServerError)
		log.Error("Error creating tcp connection\n")
	}
}

// This function transfers data between a web client and a server
func transferData(src, dest *net.TCPConn) {
	defer src.CloseWrite()
	defer dest.CloseRead()
	io.Copy(src, dest)
}

// This function performs http requests on behalf of the client
func (p *Proxy) handleHTTP(w http.ResponseWriter, req *http.Request) int64 {
	// Log Information about the request
	fmt.Println(" ")
	log.Infof("HTTP Request Received:\n%s\n", stringifyRequest(req))

	// Create http client and perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		log.Error(err)
		return -1
	}

	// Read the body of the response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		log.Error(err)
		return -1
	}

	size := int64(0)
	// Copy the headers over to the ResponseWriter
	for key, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(key, v)
			size += int64(len(v)) + int64(len(key))
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Add the response to the cache
	p.CachedSites.AddToCache(*req.URL, resp, string(body))

	// Write the response to the client
	_, err = io.WriteString(w, string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
	}

	// Send the response to the browser if user agent is curl
	agent := req.Header.Get("user-agent")
	if strings.Contains(agent, "curl") {
		err = browser.OpenReader(strings.NewReader(string(body)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err)
		}
	}
	return int64(len(body)) + size
}

// This function returns a string representation of a request
const indent = "           "

func stringifyRequest(req *http.Request) string {
	var requestString []string

	// Add the request string
	urlString := fmt.Sprintf("%s%v %v %v", indent, req.Method, req.URL, req.Proto)
	requestString = append(requestString, urlString)

	// Add the host
	requestString = append(requestString, fmt.Sprintf("%sHost: %v", indent, req.Host))

	// Loop through headers
	for name, headers := range req.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			requestString = append(requestString, fmt.Sprintf("%s%v: %v", indent, name, h))
		}
	}

	// If this is a POST, add post data
	if req.Method == http.MethodPost {
		req.ParseForm()
		requestString = append(requestString, "\n")
		requestString = append(requestString, fmt.Sprintf("%s%v", indent, req.Form.Encode()))
	}

	// Return the request as a single string
	return strings.Join(requestString, "\n")
}

// The following code serves appropriate responses for requests to a blocked site
const (
	blockedMessage = `
<html>
	<body>
		<h1>This host has been blocked!</h1>
	</body>
</html>`
)

func (p *Proxy) serveBlockedSite(w http.ResponseWriter, req *http.Request) {
	// Send the response to the client
	http.Error(w, "This site has been blocked", http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := io.WriteString(w, blockedMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
		return
	}

	log.Warnf("Access to %s has been blocked\n", req.URL.Host)

	// Send the response to the browser if the user agent is curl
	if strings.Contains(req.Header.Get("user-agent"), "curl") {
		err := browser.OpenReader(strings.NewReader(blockedMessage))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err)
		}
	}
}
