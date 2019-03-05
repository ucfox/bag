<?php

class HttpRequest{
    private $_logtype = null;

    public function __construct($logtype){
        $this->_logtype = $logtype;
    }

    public function http($url, $data = "", $headers = array(), $timeout = 4){
        // requestid
        if (!isset($headers["X-REQUEST-ID"])) {
            $headers["X-REQUEST-ID"] = empty($_SERVER['HTTP_X_REQUEST_ID']) ? substr(md5(uniqid(mt_rand(), true)), 0, 10) : $_SERVER['HTTP_X_REQUEST_ID'];
        }
        $headerArr = array();
        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_HEADER, true);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, $timeout);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
        curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);
        curl_setopt($ch, CURLOPT_ENCODING, "");
        if($headers){
            foreach($headers as $key => $value){
                $headerArr[] = $key . ': ' . $value;
            }
            curl_setopt ($ch, CURLOPT_HTTPHEADER, $headerArr);

        }
        if($data){  // 自动POST数据
            curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "POST");
            curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        }
        $starttime = microtime(true);
        $response = curl_exec($ch);
        $usetime = (int)round((microtime(true) - $starttime) * 1000, 1);
        $error    = curl_error($ch);
        if ($error) {
            XLogKit::ins($this->_logtype)->error(json_encode(array(504, $url, $usetime, $data, $headerArr, $error)));
            return false;
        }
        // Split the full response in its headers and body
        $curl_info   = curl_getinfo($ch);
        $header_size = $curl_info["header_size"];
        $header      = substr($response, 0, $header_size);
        $body        = substr($response, $header_size);
        $httpCode    = $curl_info["http_code"];
        if ($httpCode >= 300) {
            XLogKit::ins($this->_logtype)->error(json_encode(array($httpCode, $url, $usetime, $data, $headerArr, $body)));
            return false;
        }else{
            XLogKit::ins($this->_logtype)->info(json_encode(array($httpCode, $url, $usetime, $data, $headerArr, $body)));
        }

        return $body;
    }
}
