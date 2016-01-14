// XErver, a simple tiny static http(s) webserver .
package main

import("io"; "os"; "log"; "fmt"; "path"; "flag"; "strings"; "net/http"; "compress/gzip")

var (
	httpAddr	=	(flag.String("http", ":80", "the local address to use for http"))
	httpsAddr	=	(flag.String("https", "", "the local address to use for https, empty to disable it"))
	sslCert		=	(flag.String("sslcert", "", "the path to sslcert"))
	sslKey		=	(flag.String("sslkey", "", "the path to sslkey"))
	Root 		=	(flag.String("root", ".", "the public directory to serve"))
	Default 	=	(flag.String("default", "404", "set the default url.path when there is a 404 error"))
	Methods 	=	(flag.String("methods", "GET", "the allowed request methods"))
	TTL		=	(flag.Int("ttl", -1, "how many seconds will the cache live ?"))
	GZLevel		=	(flag.Int("gzip", -1, "the gzip compression level, -1/0 for none 9 best compression"))
	MethodsArray []string
)

type GzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (this GzipResponseWriter) Write(data []byte) (int, error) {
	return this.Writer.Write(data)
}

func init() {
	flag.Parse()
	MethodsArray = strings.Split(strings.TrimSpace(strings.ToUpper(*Methods)), "|")
	if *GZLevel > 9 {
		*GZLevel = 9
	}
	fmt.Println(`Welcom to XErver (v1.13)`)
	fmt.Printf(`(*) HTTPServerOn: 	%s %s`, *httpAddr, "\n")
	if *httpsAddr != "" && *sslKey != "" && *sslCert != "" {
		fmt.Println(`(*) HTTPSServerOn: %s %s`, *httpsAddr, "\n")
		fmt.Println(`(*) SSLCert: %s %s`, *sslCert, "\n")
		fmt.Println(`(*) SSLKey: %s %s`, *sslKey, "\n")
	}
	fmt.Printf(`(*) RootDirectory: 	%s %s`, *Root, "\n")
	fmt.Printf(`(*) GZipLevel: 		%d %s`, *GZLevel, "\n")
	fmt.Printf(`(*) CacheTTL: 		%d %s`, *TTL, "\n")
	fmt.Printf(`(*) DefaultURLPath: 	%s %s`, *Default, "\n")
	fmt.Printf(`(*) AllowedMethods: 	%s %s`, *Methods, "\n")
	fmt.Println(``)
}

func main() {
	block := make(chan struct{})
	go log.Fatal(http.ListenAndServe(*httpAddr, serve()))
	if *httpsAddr != "" && *sslKey != "" && *sslCert != "" {
		go log.Fatal(http.ListenAndServeTLS(*httpsAddr, *sslCert, *sslKey, serve()))
	}
	<- block
}

func serve() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Close = true
		w.Header().Set("Connection", "close")
		allowed := false
		for i := 0; i < len(MethodsArray); i ++ {
			if MethodsArray[i] == r.Method {
				allowed = true
				break
			}
		}
		if ! allowed {
			http.Error(w, "MethodNotAllowed :)", http.StatusMethodNotAllowed)
			return
		}
		if *TTL > -1 {
			w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", *TTL))
		}
		if _, e := os.Stat(path.Join(*Root, r.URL.Path)); e != nil && (*Default != "404") {
			r.URL.Path = *Default
		}
		if *GZLevel > -1 {
			w.Header().Set("Content-Encoding", "gzip")
			gz, _ := gzip.NewWriterLevel(w, *GZLevel)
			defer gz.Close()
			w = GzipResponseWriter{w, gz}
		}
		http.FileServer(http.Dir(*Root)).ServeHTTP(w, r)
	})
}
