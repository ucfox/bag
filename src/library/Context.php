<?php

// 上下文
Class Context {
    private static $data = array();

    public static function set($key, $value) {
        if(!array_key_exists($key, self::$data)) {
            self::$data[$key] = $value;
        }
    }

    public static function get($key) {
        if(!array_key_exists($key, self::$data)) {
            return NULL;
        }

        return self::$data[$key];
    }

    public static function clear() {
        self::$data = array();
    }
}
