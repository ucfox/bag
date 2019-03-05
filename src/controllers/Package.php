<?php

// http://bag.gate.panda.tv/package/get?type=first_charge
//此文件没有用到, 同样的实现在内网接口，由马龙给前端页面

Class PackageController extends Controller_Index {

    //没有用到
    public function getAction() {
        $type = $this->getParam("type", true);
        $lev = (int)$this->getParam("lev", true, 0);

        if (empty($type)) {
            $this->RespJson(RetCode::INVALID_PARAMS, RetCode::$RetMsg[RetCode::INVALID_PARAMS], "");
        }

        $data = PackageModel::ins()->get($type, $lev);

		//兼容老版本， 2018.6 后删除
        if($type=='first_charge'){
            $data = $data[1]['list'];
        }

        if (!$data) {
            $this->RespJson(RetCode::SYSTEM_ERROR, RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }

        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $data);
    }
}
