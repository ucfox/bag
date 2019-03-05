<?php
/**
 * @brief library目录下面的文件也会被自动加载
 *        如果要改变自动加载路径，在ini里面设置application.library(可改变本地类库的默认加载路径)
 *        类名同文件名
 *        如果类名不同文件名，在文件之外是不可见的
 */
class RedisInsFac
{
    //读写分离的实例
    public static function getInsByApp($app)
    {
        $port = RedisConf::$app2InsMap[$app];
        if(!$port)
        {
            $logger = XLogKit::ins("redis");
            $logger->error(__CLASS__."::".__FUNCTION__."::".__LINE__." redis app error. do not have $app port");
            return false;
        }
        try
        {
            $conf = RedisConf::getRedisConfByPort($port);
            //兼容以前
            $masterConf = $conf['m']['domain'].':'.$conf['m']['port'].':'.$conf['m']['auth'];
            $slaveConf  = $conf['s']['domain'].':'.$conf['s']['port'].':'.$conf['s']['auth'];
            return new RedisProxy($masterConf,$slaveConf);
        }
        catch(Exception $e)
        {
            $logger = XLogKit::ins("redis");
            $logger->error(__CLASS__."::".__FUNCTION__."::".__LINE__." redis error. [{$e->getCode()}] [{$e->getMessage()}]");
            return false;
        }
    }
}

class RedisProxy
{
    const MASTER_FUNC = 1;
    const SLAVE_FUNC  = 2;
    protected static $funcs = array(
        'append'=>self::MASTER_FUNC,
        'decr'=>self::MASTER_FUNC,
        'decrby'=>self::MASTER_FUNC,
        'del'=>self::MASTER_FUNC,
        'delete'=>self::MASTER_FUNC,
        'hdel'=>self::MASTER_FUNC,
        'hincrby'=>self::MASTER_FUNC,
        'hset'=>self::MASTER_FUNC,
        'hmset'=>self::MASTER_FUNC,
        'hsetnx'=>self::MASTER_FUNC,
        'incr'=>self::MASTER_FUNC,
        'incrby'=>self::MASTER_FUNC,
        'linsert'=>self::MASTER_FUNC,
        'lpop'=>self::MASTER_FUNC,
        'lpush'=>self::MASTER_FUNC,
        'lpushx'=>self::MASTER_FUNC,
        'llen'=>self::MASTER_FUNC,
        'lrem'=>self::MASTER_FUNC,
        'lremove'=>self::MASTER_FUNC,
        'lset'=>self::MASTER_FUNC,
        'ltrim'=>self::MASTER_FUNC,
        'listtrim'=>self::MASTER_FUNC,
        'move'=>self::MASTER_FUNC,
        'mset'=>self::MASTER_FUNC,
        'msetnx'=>self::MASTER_FUNC,
        'rpop'=>self::MASTER_FUNC,
        'rpush'=>self::MASTER_FUNC,
        'rename'=>self::MASTER_FUNC,
        'renamekey'=>self::MASTER_FUNC,
        'renamenx'=>self::MASTER_FUNC,
        'sadd'=>self::MASTER_FUNC,
        'sdiff'=>self::MASTER_FUNC,
        'sdiffstore'=>self::MASTER_FUNC,
        'srem'=>self::MASTER_FUNC,
        'sremove'=>self::MASTER_FUNC,
        'sunionstore'=>self::MASTER_FUNC,
        'set'=>self::MASTER_FUNC,
        'setex'=>self::MASTER_FUNC,
        'setnx'=>self::MASTER_FUNC,
        'sort'=>self::MASTER_FUNC,
        'expire'=>self::MASTER_FUNC,
        'settimeout'=>self::MASTER_FUNC,
        'expireat'=>self::MASTER_FUNC,
        'zadd'=>self::MASTER_FUNC,
        'zdelete'=>self::MASTER_FUNC,
        'zrem'=>self::MASTER_FUNC,
        'zincrby'=>self::MASTER_FUNC,
        'zremrangebyrank'=>self::MASTER_FUNC,
        'zdeleterangebyrank'=>self::MASTER_FUNC,
        'zremrangebyscore'=>self::MASTER_FUNC,
        'zdeleterangebyscore'=>self::MASTER_FUNC,
        'exists'=>self::SLAVE_FUNC,
        'get'=>self::SLAVE_FUNC,
        'getbit'=>self::SLAVE_FUNC,
        'getmultiple'=>self::SLAVE_FUNC,
        'ttl'=>self::SLAVE_FUNC,
        'hget'=>self::SLAVE_FUNC,
        'hgetall'=>self::SLAVE_FUNC,
        'hkeys'=>self::SLAVE_FUNC,
        'hlen'=>self::SLAVE_FUNC,
        'hmget'=>self::SLAVE_FUNC,
        'hexists'=>self::SLAVE_FUNC,
        'keys'=>self::SLAVE_FUNC,
        'getkeys'=>self::SLAVE_FUNC,
        'lindex'=>self::SLAVE_FUNC,
        'lget'=>self::SLAVE_FUNC,
        'lrange'=>self::SLAVE_FUNC,
        'lgetrange'=>self::SLAVE_FUNC,
        'lsize'=>self::SLAVE_FUNC,
        'ismember'=>self::SLAVE_FUNC,
        'scontains'=>self::SLAVE_FUNC,
        'smembers'=>self::SLAVE_FUNC,
        'sgetmembers'=>self::SLAVE_FUNC,
        'sunion'=>self::SLAVE_FUNC,
        'sismember'=>self::SLAVE_FUNC,
        'type'=>self::SLAVE_FUNC,
        'zcount'=>self::SLAVE_FUNC,
        'zrange'=>self::SLAVE_FUNC,
        'zrangebyscore'=>self::SLAVE_FUNC,
        'zrevrangebyscore'=>self::SLAVE_FUNC,
        'zrank'=>self::SLAVE_FUNC,
        'zrevrank'=>self::SLAVE_FUNC,
        'zrevrange'=>self::SLAVE_FUNC,
        'zscore'=>self::SLAVE_FUNC,
        'zsize'=>self::SLAVE_FUNC,
    );
    protected $read   = null;
    protected $write  = null;
    protected $mdomain = null;
    protected $mport   = null;
    protected $mauth   = null;
    protected $sdomain = null;
    protected $sport   = null;
    protected $sauth   = null;
    public function __construct($masterConf,$slaveConf)
    {
        list($this->mdomain,$this->mport,$this->mauth) = explode(':',$masterConf);
        list($this->sdomain,$this->sport,$this->sauth) = explode(':',$slaveConf);
    }
    public function __call($func, $params)
    {
        $func = strtolower($func);
        try {
            if(self::$funcs[$func] == self::MASTER_FUNC)
            {
                if(!$this->write)
                {
                    $this->write = new Redis();
                    $this->write->connect($this->mdomain,$this->mport,0.5);
                    if($this->mauth)
                    {
                        $this->write->auth($this->mauth);
                    }
                }
                $ret = call_user_func_array(array($this->write, $func), $params);
                if ($ret===NULL)
                {
                    XLogKit::ins("redis")->error(__CLASS__."::".__FUNCTION__."::".__LINE__."redis write error.[{$this->mdomain}-{$this->mport}][$func]".json_encode($params));
                }
            }
            else if(self::$funcs[$func] == self::SLAVE_FUNC)
            {
                if(!$this->read)
                {
                    $this->read = new Redis();
                    $this->read->connect($this->sdomain,$this->sport,0.5);
                    if($this->sauth)
                    {
                        $this->read->auth($this->sauth);
                    }
                }
                $ret = call_user_func_array(array($this->read, $func), $params);
                if ($ret===NULL)
                {
                    XLogKit::ins("redis")->error(__CLASS__."::".__FUNCTION__."::".__LINE__." redis read error.[{$this->sdomain}-{$this->sport}][$func]".json_encode($params));
                }
            }
            else
            {
                throw new Exception("method $func not implemented");
            }
        }
        catch (Exception $e)
        {
            XLogKit::ins("redis")->error(__CLASS__."::".__FUNCTION__."::".__LINE__." redis error. [mport:{$this->mport}] [sport:{$this->sport}] [$func] [{$e->getCode()}] [{$e->getMessage()}]");
        }
        return $ret;
    }

    public function  __destruct()
    {
        if($this->read)
        {
            try
            {
                $this->read->close();
            }
            catch(Exception $e)
            {
                $logger = XLogKit::ins("redis");
                $logger->error(__CLASS__."::".__FUNCTION__."::".__LINE__."close read redis error.[{$this->sdomain}-{$this->sport}][{$e->getCode()}] [{$e->getMessage()}]");
            }
        }
        if($this->write)
        {
            try
            {
                $this->write->close();
            }
            catch(Exception $e)
            {
                $logger = XLogKit::logger("redis");
                $logger->error(__CLASS__."::".__FUNCTION__."::".__LINE__."close write redis error. [{$this->mdomain}-{$this->mport}][{$e->getCode()}] [{$e->getMessage()}]");
            }
        }
    }
}

class RedisConf
{
    public static function getofflineMap()
    {
        return array(
            'm' => array('domain'=>$_SERVER['REDIS_HOST'],'port'=>$_SERVER['REDIS_PORT'],'auth'=>''),
            's' => array('domain'=>$_SERVER['REDIS_HOST'],'port'=>$_SERVER['REDIS_PORT'],'auth'=>''),
        );
    }
    public static function getbetaMap()
    {
        return array(
            'm' => array('domain'=>$_SERVER['REDIS_HOST'],'port'=>$_SERVER['REDIS_PORT'],'auth'=>''),
            's' => array('domain'=>$_SERVER['REDIS_HOST'],'port'=>$_SERVER['REDIS_PORT'],'auth'=>''),
        );
    }

    /**
        * @brief 根据端口返回相关的redis配置信息
        *
        * @param $port
        *
        * @return
        * array(
            'm' => array('domain'=>'10.138.230.23','port'=>'6099','auth'=>'76801e5d341e87d2'),
            's' => array('domain'=>'10.138.230.24','port'=>'6099','auth'=>'76801e5d341e87d2'),
        ),

        * PS: 目前线上已经使用的Redis端口: $allPorts = array(6948,6949,6099,6088,6075);
     */
    public static function getRedisConfByPort($port)
    {
        if($_SERVER['ENV'] != 'online')
        {
            if($_SERVER['ENV'] == 'beta')
            {
                return RedisConf::getbetaMap();
            }
            else
            {
                return RedisConf::getofflineMap();
            }
        }

        $conf = array();
        $conf['m']['domain'] = $_SERVER['REDIS_HOST_MASTER_'.$port];
        $conf['s']['domain'] = $_SERVER['REDIS_HOST_SLAVE_'.$port];

        $conf['m']['port'] = $conf['s']['port'] = $port;

        $authKey = 'REDIS_PWD_'.$port;
        $conf['m']['auth'] = $conf['s']['auth'] = $_SERVER[$authKey];

        return $conf;
    }
    public static $app2InsMap = array(
        'hostinfo' => '6948',
    );
}
