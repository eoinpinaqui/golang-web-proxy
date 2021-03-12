# golang-web-proxy
A web proxy written in Golang with site blocking and caching functionality.

## Project Description
The objective of this project is to implement a web proxy server. A web proxy server is a local 
server, which fetches items from the internet on behalf of a web client instead of the client 
fetching them directly. This allows for caching of pages and access control.
This proxy can:
1. Respond to HTTP & HTTPS requests and displays each request on a management console. It 
forwards the request to the web server and relays the response to the browser.
2. Handle Websocket connections.
3. Dynamically block selected URLs via the management console.
4. Efficiently cache HTTP requests locally and thus saves bandwidth. It also gathers timing and 
bandwidth data to prove the efficiency of the proxy.
5. Handle multiple requests simultaneously because of its threaded implementation.

## Project Implementation
This web proxy is written in Golang using some built in networking libraries, as well as some open-source 
packages for formatted logging to the console. The web proxy listens on a port of the local 
machine, the address of which is specified through a command line argument. Once the proxy has 
started, a thread is started that accepts user input from the management console to dynamically 
block and unblock given domains and list performance data. Another thread is started that 
ensures the cache is filled with up-to-date information, using mutexes to safely delete any out-of-date entries in the cache.

## Project Requirements Implementations
### Request Handling
The Proxy struct contains a http server which listens on the TCP network address specified by the 
user. This server handles all requests and incoming connections through its handler function, 
`requestHandler()`, ensuring that blocked and cached requests are handled appropriately. All 
requests, except for connection requests, are performed by a temporary http client created by the 
web proxy. The body and headers of the response are forwarded back to the web client and are 
relayed to the browser where appropriate e.g., if the user agent for the request is curl, the 
response will be opened in the browser as well as written to the terminal.

### Websocket Connections
In the case of a connection request, the web proxy hijacks control of the client connection using 
the Hijacker interface in the golang http package. The proxy then establishes a TCP connection 
with the desired web server and asynchronously transfers the data between the client and the 
server. Once the data transfer is complete, the connections are closed.

### Dynamic Blocking
The BlockedSites struct contains a set of domains which have been blocked by an administrator 
through the management console. An administrator can block domains by entering a block 
command e.g., `block tcd.ie`. Similarly, a user can unblock domains using an unblock command 
e.g., `unblock tcd.ie`. When a domain has been blocked, all sub-domains are also blocked e.g., 
when tcd.ie has been blocked, www.scss.tcd.ie will also be blocked, as well as www.ahss.tcd.ie, 
etc. The set of blocked sites is protected by a mutex to prevent the corruption of data by different 
threads. The list of all blocked sites can be viewed using a list command in the management 
console e.g., `list blocked`.

### Caching of HTTP Requests
The CachedSites struct contains a set of cached HTTP responses that can be used to efficiently 
serve multiple requests to the same URL. This data structure stores a copy of the response body as 
a string and copy of the response headers. In this manner, when a request is sent for a cached site, 
the response can be used from the cache in an efficient manner, rather than continuously 
performing the same requests repeatedly. The list of all cached responses can be viewed using a 
list command in the management console e.g., `list cached`.

The PerformanceData struct contains a set of both timing and bandwidth data for requests to both 
uncached and cached URLs. Every time a HTTP request is made, the time taken to perform that 
request is noted and added to the appropriate entry in the PerformanceData struct. An 
administrator can view such data by using a list command in the management console e.g., `list 
timing` to see the average time for uncached and cached URLs and `list bandwidth` to see the 
average bandwidth for uncached and cached URLs. The set of cached responses is protected by a 
mutex to prevent the corruption of data by different threads.

### Asynchronous Implementation
This server uses many goroutines to handle asynchronous requests effectively. The proxy listens 
for requests using the `http.ListenAndServe()` function, which creates a goroutine for every 
request that is received. Therefore, the server can handle multiple asynchronous requests at once



