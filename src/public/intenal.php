<?php
// 内网接口
require_once("/home/q/php/gsdk_base/sdk_base.php");
define("APP_PATH",  dirname(dirname(__FILE__)));
ob_start();
// 初始化应用
$app  = new Yaf_Application(APP_PATH . "/conf/intenal.ini");
$app->bootstrap()->run();
