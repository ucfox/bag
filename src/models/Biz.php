<?php
/**
 * User: Kami <chanyanbing@panda.tv>
 * Date: 2018/1/18
 */

class BizModel
{

    /**
     * Application
     */
    const APP = 'pandaren';

    /**
     * @var string
     */
    private $_env = null;

    /**
     * @var
     */
    private $_logger     = 'biz';

    /**
     * @var string
     */
    private $_address    = 'biz.gate.pdtv.io:8360';

    /**
     * @var HttpRequest
     */
    private $_httpclient = null;

    /**
     * @var string
     */
    private $_caller     = "bag";

    /**
     * 单例
     * @var BizModel
     */
    private static $ins = null;

    /**
     * 获取单例
     *
     * @param string $env
     * @return BizModel
     */
    public static function ins($env = '')
    {
        if(!(self::$ins instanceof self)){
            self::$ins = new self($env);
        }
        return self::$ins;
    }

    /**
     * BizModel constructor.
     */
    public function __construct($env = '')
    {
        // 环境变量
        $this->_env        = $env ?: $_SERVER['ENV'];
        $this->_httpclient = new HttpRequest($this->_logger);
    }

    /**
     * 使用弹幕道具卡
     *
     * @param int $uid
     * @param array $conf
     * @return bool
     */
    public function useCard($uid, array $conf)
    {

        $callerstrace = Context::get('callerstrace');
        // 请求参数
        $params = [
            'rid'     => $uid,
            'conf'    => json_encode($conf),
            'app'     => self::APP,
            'callerstrace' => $callerstrace,
            '_caller' => $this->_caller,
        ];

        $url = $this->_BuildUrl('/barrage/usecard?' . http_build_query($params));
        $ret = $this->_httpclient->http($url);

        $ret_json = json_decode($ret, true);
        // json格式数据不正确
        if(json_last_error() !== JSON_ERROR_NONE) {
            XLogKit::logger($this->_logger)->error("useCard fail uid={$uid}&conf=" . json_encode($conf) . "&ret={$ret}");
            return false;
        }
        // 包含不正常的返回，不正确的格式，错误的响应码
        if(!$ret_json || !isset($ret_json['errno']) || $ret_json['errno'] != 0) {
            XLogKit::logger($this->_logger)->error("useCard fail uid={$uid}&conf=" . json_encode($conf) . "&ret={$ret}");
            return false;
        }
        return true;
    }

    /**
     * 获取域名URL
     *
     * @return string
     */
    protected function _getBaseUrl()
    {
        if ($this->_env === 'online') {
            $prefix = '';
        } else {
            $prefix = 'beta.';
        }
        return "http://{$prefix}{$this->_address}";
    }

    /**
     * 处理请求URL
     *
     * @param string $apiurl
     * @return string
     */
    protected function _BuildUrl($apiurl) {
        return $this->_getBaseUrl() . $apiurl . "&_caller={$this->_caller}";
    }
}
