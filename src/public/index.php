<?php
define("APP_PATH",  dirname(dirname(__FILE__)));
ob_start();
// 初始化应用
$app  = new Yaf_Application(APP_PATH . "/conf/app.ini");
$app->bootstrap()->run();
