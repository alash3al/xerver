xerver v2.0
============
just a light and fast reverse proxy for fastcgi based processes .

Features
============
* Cross platform .  
* Accelerated and optimized without modules hell.  
* No configurations needed .  
* Standalone, Tiny & Lightweight .  
* Supports both http and https .  
* Automatically use HTTP/2 "in https" .  
* Control the whole webserver just with your preferred programming language .  
* Tell xerver to perform some operations  using http-headers, i.e "send-file, proxy-pass, ... etc" .  
* More is coming, just stay tuned .

How It Works 
=============
* A request hits the `xerver` .  
* `xerver` handles the request .  
* `xserver` send it to the backend `fastcgi` process and the main controller file .  
* the controller file contains your own logic .    
* `fastcgi` process reply to `xerver` with the result . 
* `xerver` parse the result and then prepare it to be sent to the client .  

Installation
==============
1-	Download the right binary for your os from [here](#) .
2- Extract the downloaded file contents to any directory say `./xerver/` "current directory" .
3- Using your `Terminal` `cd ./xerver/` .
4- run the following command to display the available options `./xerver --help`.

Example (1)
==============
**Only acts as a static file server** 
```bash
./xerver --static-dir=/path/to/www/ --addr=0.0.0.0:80
```

Example (2)
==============
**Listen on address `0.0.0.0:80`** and send the requests to `./controller.php`  
```bash
./xerver --fcgi-proto=unix --fcgi-addr=/path/to/php-fpm.socks --fcgi-controller=./controller.php --http-addr=:80
```
** OR Listen on address `0.0.0.0:80` & ``0.0.0.0:443`` ** and send the requests to `./controller.php` 
```bash
./xerver --fcgi-proto=unix --fcgi-addr=/path/to/php-fpm.socks --fcgi-controller=./controller.php --http-addr=:80 --https-addr=:443 --https-cert=./cert.pem --https-key=./key.pem
```

**Open your ./controller.php** and :
```php
<?php
	
	// uncomment any of the following to test it .

	// here you perform your own logic
	// echo "<pre>" . print_r($_SERVER, 1);

	// some xerver internal header for some operations
	// 1)- tell xerver to serve a file/directory to the client .
	// header("Xerver-Internal-FileServer: " . __DIR__ . "/style.css");
	
	// 2)- tell xerver to serve from another server "act as reverse proxy" .
	// header("Xerver-Internal-ProxyPass: http://localhost:8080/");

	// 3)- tell xerverto hide its own tokens "A.K.A Server header"
	// header("Xerver-Internal-ServerTokens: off");

	// the above headers won't be sent to the client .
```

**Open your browser** and go to `localhost` or any `localhost` paths/subdomains .  


Building from source
==================
1- make sure you have `Golang` installed .
2- `go get github.com/alash3al/xerver`
3- `go install github.com/alash3al/xerver`
4- enjoy .
