<?php
	// this file is a simple xerver controller that
	// acts as a configurations file i.e "nginx.conf" but in php
	// ----------------------------------------------
	// handle exceptions
	set_exception_handler(function($e){
		die($e);
	});

	// prepare the environment
	ksort($_SERVER);
	$_SERVER["SCRIPT_FILENAME"] = __DIR__ . "/index.php";
	$_SERVER["SCRIPT_NAME"] = $_SERVER["PHP_SELF"] = "/index.php";

	// deny access to this file .
	if ( basename($_SERVER["REQUEST_PATH"]) == basename(__FILE__) ) {
		exit("404 not found");
	}

	// the required file is valid php file ?
	$ext = pathinfo($_SERVER["REQUEST_PATH"])["extension"] ?? "";
	if ( $ext == "php" && is_file($filename = __DIR__ . $_SERVER["REQUEST_PATH"]) ) {
		$_SERVER["SCRIPT_FILENAME"] = __DIR__. $_SERVER["REQUEST_PATH"];
		$_SERVER["SCRIPT_NAME"] = $_SERVER["PHP_SELF"]  = $_SERVER["REQUEST_PATH"];
		require($filename);
		exit;
	}

	// the required file is avalid none-php file ?
	if ( is_file($filename = __DIR__ . $_SERVER["REQUEST_PATH"]) && ! is_dir($filename) ) {
		header("Xerver-Internal-FileServer: " . __DIR__ . $_SERVER["REQUEST_PATH"]);
		header("Cache-Control: max-age=2592000");
		exit;
	}

	// the required file is a directory ?
	if ( is_dir($dirname = __DIR__ . $_SERVER["REQUEST_PATH"]) ) {
		$_SERVER["REQUEST_URI"] = rtrim($_SERVER["REQUEST_PATH"] , "/") .  "/" . ( $_SERVER["QUERY_STRING"] ? "?" . $_SERVER["QUERY_STRING"] : "" ); 
		if ( is_file($filename = $dirname . "/index.php") ) {
        	$_SERVER["SCRIPT_FILENAME"] = $filename;
       		$_SERVER["SCRIPT_NAME"] = $_SERVER["PHP_SELF"]  = rtrim($_SERVER["REQUEST_PATH"] , "/") . "/index.php";
			require($filename);
		}
		else if ( is_file($filename = $dirname . "/index.html") ) {
			header("Xerver-Internal-FileServer: " . $filename);
		}
		else {
			echo "404 not found";
		}
		exit;
	}

	// anything else, redirect to index.php ..
	require(__DIR__ . "/index.php");

