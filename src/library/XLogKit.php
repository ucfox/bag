<?php
/**
 * @brief 延续了以前pylon的命名
 *        参考pylon以前的实现
 */
class XLogKit
{
    CONST DEBUG_LEVEL = 0;
    CONST INFO_LEVEL  = 1;
    CONST WARN_LEVEL  = 2;
    CONST ERROR_LEVEL = 3;

    /**
        * @brief
        *
        * @param $prj       项目名称,决定了日志记录到哪个文件夹,默认就是$_SERVER['PRJ_NAME']
        * @param $tag       默认一般长这样,tag[online,,1d3ab75e2f]
        * @param $logLevel  日志级别,默认是error级别,即只记录error日志
        *
        * @return
     */
    public static function init($prj="",$tag="",$logLevel=self::ERROR_LEVEL)
    {
        if(!$prj)
        {
            $prj = $_SERVER['PRJ_NAME'];
        }
        if(isset($_SERVER['USER']))
        {
            $tag =  $_SERVER['USER'] . ",$tag" ;
        }
        //添加TID
        $tag .= ','.XTid::get();
        log_kit::init($prj,$tag,$logLevel);
    }
    //HttpRequest.php 中用了 ins
    public static function ins($name)
    {
        if(!$name)
        {
            $name = '_all';
        }
        return new Logger($name);
    }
    public static function logger($name)
    {
        if(!$name)
        {
            $name = '_all';
        }
        return new Logger($name);
    }
    public function toall($toall=1)
    {
        return log_kit::toall($toall);
    }
}

//参考pylon以前的实现
class XTid
{
    static private $_tid = '';
    public static function get()
    {
        if(empty(self::$_tid))
        {
            self::$_tid = self::init();
        }
        return self::$_tid;
    }

    private static function init()
    {
        if(!empty($_SERVER['ENV']) && $_SERVER['ENV'] == 'dev' && !empty($_REQUEST['__tid']))
        {
            return $_REQUEST['__tid'];
        }
        return empty($_SERVER['HTTP_X_REQUEST_ID']) ? substr(md5(uniqid(mt_rand(), true)), 0, 10) : $_SERVER['HTTP_X_REQUEST_ID'];
    }
}
