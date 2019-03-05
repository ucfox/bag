<?php

Class Pool {
    public static $pools = array();

    public static function getMysql($alias = 'slave') {
        $alias = "mysql_".$alias;
        if(!isset(self::$pools[$alias])) {
            if ($alias == "mysql_slave") {
                self::$pools[$alias] = new Mysql($_SERVER["DB_HOST_SLAVE"], $_SERVER["DB_USER"], $_SERVER["DB_PWD"], $_SERVER["DB_NAME"]);
            } else {
                self::$pools[$alias] = new Mysql($_SERVER["DB_HOST"], $_SERVER["DB_USER"], $_SERVER["DB_PWD"], $_SERVER["DB_NAME"]);
            }
        }
        return self::$pools[$alias];
    }
}
