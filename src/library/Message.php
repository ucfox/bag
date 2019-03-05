<?php

/**
 * 消息推送系统控制模块
 *
 * @author Kami
 */
class Message
{

    /**
     * caller
     * @var string
     */
    protected static $caller = 'bag';

    /**
     * 站点连接地址
     * @var string
     */
    protected static $baseurl = 'message.pdtv.io:8360';

    /**
     * 向用户推送消息
     *
     * @param string $logtype   日志类型
     * @param int    $uid       用户ID
     * @param int    $cat       消息类型
     * @param string $title     消息标题
     * @param string $content   消息内容
     *
     * @return bool
     */
    public static function sendMessageToUser($logtype, $uid, $cat, $title, $content)
    {
        // http client
        $httpclient = new HttpRequest($logtype);
        // send
        $ret = $httpclient->http(self::getUrl('/Message/sendMessageToUser'),
                          self::buildQuery([
                            'title'     => $title,
                            'cat'       => $cat,
                            'to_uid'    => $uid,
                            'content'   => $content,
                            '_caller'   => self::$caller,
                          ]));
        // decode
        $json = json_decode($ret, true);

        // json格式数据不正确
        if(json_last_error() !== JSON_ERROR_NONE) {
            XLogKit::logger($logtype)->error("send message fail sendMessageToUser logtype={$logtype}&uid={$uid}&cat={$cat}&title={$title}&content={$content}&ret={$ret}");
            return false;
        }
        // 包含不正常的返回，不正确的格式，错误的响应码
        if(!$json || !isset($json['errno']) || $json['errno'] != 0) {
            // 失败
            XLogKit::logger($logtype)->error("send message fail sendMessageToUser logtype={$logtype}&uid={$uid}&cat={$cat}&title={$title}&content={$content}");
            return false;
        }

        return true;
    }

    /**
     * 获取访问路径
     *
     * @param string $path
     *
     * @return string
     */
    protected static function getUrl($path)
    {
        $prefix = $_SERVER['ENV'] !='online' ? 'beta.' : '';
        $baseurl = "http://" . $prefix . self::$baseurl;
        return sprintf('%s%s', $baseurl, $path);
    }

    /**
     * 生成请求字符串序列
     *
     * @param array $data
     *
     * @return string
     */
    protected static function buildQuery($data = array())
    {
        if(count($data) < 1) {
            return '';
        }
        // urlencode
        foreach($data as $key => $value) {
            $data[$key] = rawurlencode($value);
        }
        return http_build_query($data);
    }

}
