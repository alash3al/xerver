<?php
	// already loaded ?
	if ( defined("XERVER") ) {
		exit;
	}

	// just for internal usage
	define("XERVER", true);

	// handle exceptions
	set_exception_handler(function($e){
		die($e);
	});

	// autoserve from the $root
	function serve($root, $options = []) {
		// prepare options
		$options = array_merge([
			"index"	=>	["index.php", "index.html"],
			"directory_listing" => true,
			"e404" => function(){
				exit("404 not found");
			}
		], $options);
		// set the current working directory
		$_SERVER["DOCUMENT_ROOT"] = $root;
		// get the requested file extension
		$info = pathinfo($_SERVER["REQUEST_PATH"]);
		if ( ! isset($info["extension"]) ) {
			$info = ["extension" => ""];
		}
		$ext = $info["extension"];
		// prepare the request path
		$required = rtrim(str_replace(["/", "\\"], DIRECTORY_SEPARATOR, $root), DIRECTORY_SEPARATOR) . DIRECTORY_SEPARATOR . trim($_SERVER["REQUEST_PATH"], "/");
		// 1)- a directory ?
		if ( is_dir($required) ) {
			if ( $_SERVER["REQUEST_PATH"][strlen($_SERVER["REQUEST_PATH"]) - 1] != "/" && $_SERVER["REQUEST_PATH"] != "/"&& $_SERVER["REQUEST_PATH"] != "" ) {
				header("Location: " . $_SERVER["REQUEST_PATH"] . "/");
				exit;
			}
			chdir($required);
			$_SERVER["REQUEST_URI"] = rtrim($_SERVER["REQUEST_PATH"], "/") . "/" . ($_SERVER["QUERY_STRING"] ? ("?" . $_SERVER["QUERY_STRING"]) : "");
			$_SERVER["REQUEST_PATH"] = $_SERVER["PATH_INFO"] .= "/";
			foreach ( $options["index"] as $i ) {
				if ( is_file($filename = $required . "/" . $i) ) {
					$_SERVER["SCRIPT_FILENAME"] = $filename;
					$_SERVER["SCRIPT_NAME"] = $_SERVER["PHP_SELF"] = rtrim($_SERVER["REQUEST_PATH"], "/") . "/" . $i;
					ksort($_SERVER);
					require($filename);
					exit;
				}
			}
			if ( $options["directory_listing"] ) {
				header("Xerver-Internal-FileServer: " . $required);
				exit;
			}
			exit("You don't have permissions to view this directory");
		}
		// 2)- is file ?
		else if ( is_file($required) ) {
			chdir(dirname($required));
			$_SERVER["SCRIPT_FILENAME"] = $required;
			$_SERVER["SCRIPT_NAME"] = $_SERVER["PHP_SELF"] = $_SERVER["REQUEST_PATH"];
			ksort($_SERVER);
			if ( $ext != "php" ) {
				header("Xerver-Internal-FileServer: " . $required);
				exit;
			}
			require($required);
			exit;
		}
		// 3)- not found ?
		else {
			$options["e404"]();
		}
	}

	// path rewriter
	function rewrite($old, $new) {
		$_SERVER["REQUEST_PATH"] = "/" . ltrim(preg_replace("~{$old}~", $new, $_SERVER["REQUEST_PATH"]), "/");
		$_SERVER["PATH_INFO"] = $_SERVER["REQUEST_PATH"];
	}

	// path router
	function on(string $pattern, Closure $fn) {
		$path = preg_replace("~/+~", "/", "/" . $_SERVER["PATH_INFO"] . "/");
		$pattern = preg_replace("~/+~", "/", "/" . $pattern . "/");
		if ( preg_match("~^{$pattern}$~", $path, $m) ) {
			array_shift($m);
			call_user_func_array($fn, $m);
		}
	}

	// vhost router
	function vhost($pattern, Closure $fn) {
		$host = explode(":", $_SERVER["HTTP_HOST"]);
		if ( ! isset($host[0]) ) {
			$host = "localhost";
		} else {
			$host = $host[0];
		}
		if ( preg_match("~^{$pattern}$~i", $host, $m) ) {
			array_shift($m);
			call_user_func_array($fn, $m);
		}
	}
