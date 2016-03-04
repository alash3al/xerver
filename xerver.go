// xerver 2.0, a tiny and light fastcgi reverse proxy only,
// copyright 2016, (c) Mohammed Al Ashaal <http://www.alash3al.xyz>,
// published uner MIT licnese .
package main

import "os"
import "io"
import "fmt"
import "log"
import "net"
import "flag"
import "strconv"
import "strings"
import "net/url"
import "net/http"
import "path/filepath"
import "net/http/httputil"
import "github.com/tomasen/fcgi_client"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
    VERSION         string      =   "xerver/v2.0"
    STATIC_DIR		*string 	=	flag.String("static-dir", "none", "the static directory to serve static files")
    FCGI_PROTO      *string     =   flag.String("fcgi-proto", "none", "the fastcgi protocol [unix, tcp, none]")
    FCGI_ADDR       *string     =   flag.String("fcgi-addr", "none", "the fastcgi address/location i.e '/run/php/php-fpm.sock'")
    FCGI_CONTROLLER *string     =   flag.String("fcgi-controller", "none", "the main fascgi controller i.e '/root/main.php'")
    HTTP_ADDR       *string     =   flag.String("http-addr", ":80", "the xerver http address")
    HTTPS_ADDR      *string     =   flag.String("https-addr", "none", "the xerver https address")
    HTTPS_KEY       *string     =   flag.String("https-key", "none", "the xerver https ssl key filename")
    HTTPS_CERT      *string     =   flag.String("https-cert", "none", "the xerver https ssl cert filename")
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func ServePHP(res http.ResponseWriter, req *http.Request) {
    // we only allow [GET, POST, PUT, DELETE, HEAD, OPTIONS, PATCH]
    if req.Method != "GET" && req.Method != "POST" && req.Method != "PUT" && req.Method != "DELETE" && req.Method != "HEAD" && req.Method != "OPTIONS" && req.Method != "PATCH" {
        http.Error(res, "we don't support the requested action", 405)
        return
    }
    // a helper function that will split the provided input
    // using the specified delemiter into 2 parts
    // but return them as two returns .
    split := func(in string, del string) (string, string) {
        s := strings.SplitN(in, del, 2)
        if len(s) < 1 {
            s = append(s, "", "")
        } else if len(s) < 2 {
            s = append(s, "")
        }
        return s[0], s[1]
    }
    // connect to the fastcgi backend,
    // and check whether there is an error or not .
    fcgi, err := fcgiclient.Dial(*FCGI_PROTO, *FCGI_ADDR)
    if err != nil {
        log.Println(err)
        http.Error(res, "Unable to connect to the backend", 502)
        return
    }
    // automatically close the fastcgi connection and the requested body at the end .
    defer fcgi.Close()
    defer req.Body.Close()
    // prepare some vars :
    // -- http[addr, port]
    // -- https[addr, port]
    // -- remote[addr, host, port]
    // -- stat of the requested filename
    // -- environment variables
    http_addr, http_port := split(*HTTP_ADDR, ":")
    https_addr, https_port := split(*HTTPS_ADDR, ":")
    remote_addr, remote_port := split(req.RemoteAddr, ":")
    hosts, _ := net.LookupAddr(remote_addr)
    stat, err := os.Stat(req.URL.Path)
    env := map[string]string {
        "DOCUMENT_ROOT"             :   filepath.Dir(*FCGI_CONTROLLER),
        "SCRIPT_FILENAME"           :   *FCGI_CONTROLLER,
        "SCRIPT_NAME"               :   "/index.php",
        "REQUEST_METHOD"            :   req.Method,
        "REQUEST_FILE_NAME"         :   req.URL.Path,
        "REQUEST_FILE_EXISTS"       :   fmt.Sprintf("%t", stat != nil && os.IsExist(err)),
        "REQUEST_FILE_IS_DIR"       :   fmt.Sprintf("%t", stat != nil && stat.IsDir()),
        "REQUEST_FILE_EXTENSION"    :   filepath.Ext(req.URL.Path),
        "REQUEST_URI"               :   req.URL.RequestURI(),
        "REQUEST_PATH"              :   req.URL.Path,
        "PATH_INFO"                 :   req.URL.Path,
        "ORIG_PATH_INFO"            :   req.URL.Path,
        "PATH_TRANSLATED"           :   *FCGI_CONTROLLER,
        "PHP_SELF"                  :   *FCGI_CONTROLLER,
        "CONTENT_LENGTH"            :   fmt.Sprintf("%d", req.ContentLength),
        "CONTENT_TYPE"              :   req.Header.Get("Content-Type"),
        "REMOTE_ADDR"               :   remote_addr,
        "REMOTE_PORT"               :   remote_port,
        "REMOTE_HOST"               :   hosts[0],
        "QUERY_STRING"              :   req.URL.Query().Encode(),
        "SERVER_SOFTWARE"           :   VERSION,
        "SERVER_NAME"               :   req.Host,
        "SERVER_ADDR"               :   http_addr,
        "SERVER_PORT"               :   http_port,
        "SERVER_PROTOCOL"           :   req.Proto,
        "SERVER_TEMP_DIR"           :   os.TempDir(),
        "SCHEME"                    :   "http",
        "HTTPS"                     :   "",
        "HTTP_HOST"                 :   req.Host,
    }
    // tell fastcgi backend that, this connection is done over https connection if detected .
    if req.TLS != nil {
        env["SCHEME"] = "https"
        env["HTTPS"] = "on"
        env["SERVER_PORT"] = https_port
        env["SERVER_ADDR"] = https_addr
    }
    // iterate over request headers and append them to the environment varibales in the valid format .
    for k, v := range req.Header {
        env["HTTP_" + strings.Replace(strings.ToUpper(k), "-", "_", -1)] = strings.Join(v, ";")
    }
    // fethcing the response from the fastcgi backend,
    // and check for errors .
    resp, err := fcgi.Request(env, req.Body)
    if err != nil {
        log.Println("err> ", err.Error())
        http.Error(res, "Unable to fetch the response from the backend", 502)
        return
    }
    // parse the fastcgi status .
    resp.Status = resp.Header.Get("Status")
    resp.StatusCode, _ = strconv.Atoi(strings.Split(resp.Status, " ")[0])
    if resp.StatusCode < 100 {
        resp.StatusCode = 200
    }
    // automatically close the fastcgi response body at the end .
    defer resp.Body.Close()
    // read the fastcgi response headers,
    // exclude "Xerver-Internal-*" headers from the response,
    // and apply the actions related to them .
    for k, v := range resp.Header {
        if ! strings.HasPrefix(k, "Xerver-Internal-") {
            for i := 0; i < len(v); i ++ {
                if res.Header().Get(k) == "" {
                    res.Header().Set(k, v[i])
                } else {
                    res.Header().Add(k, v[i])
                }
            }
        }
    }
    // remove server tokens from the response
    if resp.Header.Get("Xerver-Internal-ServerTokens") != "off" {
        res.Header().Set("Server", VERSION)
    }
    // serve the provided filepath using theinternal fileserver
    if resp.Header.Get("Xerver-Internal-FileServer") != "" {
        res.Header().Del("Content-Type")
        http.ServeFile(res, req, resp.Header.Get("Xerver-Internal-FileServer"))
        return
    }
    // serve the response from another backend "http-proxy"
    if resp.Header.Get("Xerver-Internal-ProxyPass") != "" {
        u, e := url.Parse(resp.Header.Get("Xerver-Internal-ProxyPass"))
        if e != nil {
            log.Println("err> ", e.Error())
            http.Error(res, "Invalid internal-proxypass value", 502)
            return
        }
        httputil.NewSingleHostReverseProxy(u).ServeHTTP(res, req)
        return
    }
    // fix the redirect issues by fetching the fastcgi response location header
    // then redirect the client, then ignore any output .
    if resp.Header.Get("Location") != "" {
        http.Redirect(res, req, resp.Header.Get("Location"), resp.StatusCode)
        return
    }
    // write the response status code .
    res.WriteHeader(resp.StatusCode)
    // only sent the header if the request method isn't HEAD .
    if req.Method != "HEAD" {
        io.Copy(res, resp.Body)
    }
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// - parse the cmd flags
// - check for the required flags
// - display welcome messages
func init() {
    flag.Parse()
    if *STATIC_DIR == "none" && (*FCGI_PROTO == "none" || *FCGI_ADDR == "none" || *FCGI_CONTROLLER == "none") {
        log.Fatal("You must configure xerver to only act as a reverse/static server")
    }
    if strings.HasPrefix(*HTTP_ADDR, ":") {
        *HTTP_ADDR = "0.0.0.0" + *HTTP_ADDR
    }
    if strings.HasPrefix(*HTTPS_ADDR, ":") {
        *HTTPS_ADDR = "0.0.0.0" + *HTTPS_ADDR
    }
    if *FCGI_CONTROLLER != "none" {
        _, err := os.Stat(*FCGI_CONTROLLER)
        if err != nil {
            log.Fatal("err> ", err.Error())
        }
    }
    fmt.Println("# Welcome to ",          VERSION)
    fmt.Println("# Static Dir: ",         *STATIC_DIR)
    fmt.Println("# FCGI Address: ",       *FCGI_ADDR)
    fmt.Println("# FCGI Controller: ",    *FCGI_CONTROLLER)
    fmt.Println("# HTTP Address: ",       *HTTP_ADDR)
    fmt.Println("# HTTPS Address: ",      *HTTPS_ADDR)
    fmt.Println("# HTTPS Key: ",          *HTTPS_KEY)
    fmt.Println("# HTTPS Cert: ",         *HTTPS_CERT)
    fmt.Println("")
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// let's play :)
func main() {
	// handle any panic
	defer (func(){
		if err := recover(); err != nil {
			log.Println("err> ", err)
		}
	})()
	// the handler
	handler := func(res http.ResponseWriter, req *http.Request) {
		if *STATIC_DIR == "none" {
			ServePHP(res, req)
			return
		}
		http.FileServer(http.Dir(*STATIC_DIR)).ServeHTTP(res, req)
	}
    // an error channel to catch any error
    err := make(chan error)
    // run a http server in a goroutine
    go (func(){
        err <- http.ListenAndServe(*HTTP_ADDR, http.HandlerFunc(handler))
    })()
    // run a https server in another goroutine
    go (func(){
        if *HTTPS_ADDR != "none" && *HTTPS_CERT != "none" && *HTTPS_KEY != "none" {
            err <- http.ListenAndServeTLS(*HTTPS_ADDR, *HTTPS_CERT, *HTTPS_KEY, http.HandlerFunc(handler))
        }
    })()
    // there is an error occurred, 
    // let's catch it, then exit .
    log.Fatal(<- err)
}
