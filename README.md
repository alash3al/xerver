# xerver
a tiny static http(s) web server written in golang

# requirements
[golang](https://golang.org/dl/)

# installation
(1)- `go get github.com/alash3al/xerver` .   
(2)- `go install github.com/alash3al/xerver` .  

**NOTE** you should add `GOPATH/bin` to your `PATH` ..  

# usage 
- `xerver -option value` or `xerver --option=value`  
- `xerver --help` for full arguments .  

# options
```
  -default string
        set the default url.path when there is a 404 error (default "404")
  -gzip int
        the gzip compression level, -1/0 for none 9 best compression (default -1)
  -http string
        the local address to use for http (default ":80")
  -https string
        the local address to use for https, empty to disable it
  -methods string
        the allowed request methods (default "GET")
  -root string
        the public directory to serve (default ".")
  -sslcert string
        the path to sslcert
  -sslkey string
        the path to sslkey
  -ttl int
        how many seconds will the cache live ? (default -1)
```

# examples

**Serving files from `/htdocs/`** `xerver -root /htdocs/`   
**Serving files from `/htdocs/` and listen on "0.0.0.0:8080"** `xerver -root /htdocs/ -http :8080`   

# author
[Mohammed Al Ashaal](http://www.alash3al.xyz)
