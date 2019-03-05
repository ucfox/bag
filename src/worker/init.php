<?php

try {
    date_default_timezone_set('Asia/Shanghai');
    define("APP_PATH",  dirname(dirname(__FILE__)));
    $app  = new Yaf_Application(APP_PATH . "/conf/app.ini");
    $app->bootstrap()->execute("main", $argc, $argv);
} catch (Exception $e) {
    XLogKit::logger('exception')->error(json_encode($e));
}
