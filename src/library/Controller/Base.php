<?php

Class Controller_Base extends Yaf_Controller_Abstract {

    public function RespJson($errno, $errmsg, $data = array()) {
        header('Content-Type:application/json; charset=utf-8');
        $ret = array(
            "errno" => (int)$errno,
            "errmsg" => (string)$errmsg,
            "data" => $data,
        );
        exit(json_encode($ret));
    }

    public function RespError($errmsg,$errno=-1) {
        if(is_int($errmsg)){
            $errno = $errmsg;
            $errmsg = isset(RetCode::$RetMsg[$errno]) ? RetCode::$RetMsg[$errno] : '';
        }
        $this->RespJson($errno,$errmsg);
    }

    public function RespSuccess($data=array()) {
        $this->RespJson(RetCode::SUCCESS,RetCode::$RetMsg[RetCode::SUCCESS],$data);
    }

    public function getParam($key, $isFilter = false, $default = NULL){
        $val = $this->getRequest()->getQuery($key, $default);
        if ($isFilter) {
            return Tools::xssFilter($val);
        }
        return $val;
    }

    public function postParam($key, $isFilter = false, $default = NULL){
        $val = $this->getRequest()->getPost($key, $default);
        if ($isFilter) {
            return Tools::xssFilter($val);
        }
        return $val;
    }

    public function reqParam($key, $isFilter = false, $default = NULL){
        $val = isset($_REQUEST[$key]) ? $_REQUEST[$key] : $default;
        if ($isFilter) {
            return Tools::xssFilter($val);
        }
        return $val;
    }
}
