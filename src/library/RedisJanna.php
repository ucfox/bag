<?php

// 加载janna
require_once '/home/q/php/phpbase/JannaRedis.php';

class RedisJanna
{

    /**
     * 日志
     */
    protected $logger = null;

    /**
     * redis链接资源
     * @var
     */
    protected $_redis = null;

    /**
     * magic: __construct
     */
    public function __construct($callname, $target, $passwd = '')
    {
        // 日志实例
        $this->logger = XLogKit::logger('janna_redis');
        try{
            $this->_redis = JannaRedis::getIns($callname, $target, $passwd);
            if($this->_redis == false) {
                throw new Exception('redis connect failed');
                $this->logger->error('redis connect failed ' . __METHOD__ . " callname={$callname}&target={$target}&passwd={$passwd}");
            }
        } catch(Exception $e) {
            throw $e;
        }
    }

    /**
     * 调用redis命令
     *
     * @param string $method 命令名
     * @param array  $args   参数列表
     * @return mixed
     */
    public function __call($method, $args)
    {
        return call_user_func_array([$this->_redis, $method], $args);
    }

}
