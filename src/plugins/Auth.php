<?php

class AuthPlugin extends Yaf_Plugin_Abstract {

    /**
     * 忽略验证模块名单
     * @var string[]
     */
    protected $ignore_modules = ["intenal"];

    // 分发循环开始之前被触发
    public function dispatchLoopStartup(Yaf_Request_Abstract $request, Yaf_Response_Abstract $response)
    {
        // The module
        $module = strtolower($request->module);
        // 检查忽略名单
        if(!in_array($module, $this->ignore_modules)) {
            $res = UserModel::ins()->auth();
            if ($res === array()) {
                // 鉴权失败
                $ret = array(
                    "errno" => RetCode::USER_NOT_LOGIN,
                    "errmsg" => RetCode::$RetMsg[RetCode::USER_NOT_LOGIN],
                    "data" => "",
                );
                exit(json_encode($ret));
            }
        }

        // 设置上下文
        if(isset($res["data"]["rid"])){
            Context::set("uid", $res["data"]["rid"]);
            Context::set("nickName", $res["data"]["nickName"]);
        }
    }
}
