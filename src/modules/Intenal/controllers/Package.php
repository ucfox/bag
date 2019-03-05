<?php

Class PackageController extends Controller_Intenal {

    // 获取背包物品
    public function getAction() {
        $uid = Context::get("uid");
        $type = $this->getParam("type", true);
        $lev = (int) $this->getParam("lev", true, 0);

//        $comb = $this->getParam("comb", true, 0);
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

    //获取移动端背包物品
    public function getMobileAction(){

        $lev = (int) $this->getParam("lev", true, 0);


        $data = PackageModel::ins()->get_first_mobile_data($lev);

        if (empty($data)) {
            $this->RespJson(RetCode::SYSTEM_ERROR, RetCode::$RetMsg[RetCode::SYSTEM_ERROR]);
        }

        $this->RespJson(RetCode::SUCCESS, RetCode::$RetMsg[RetCode::SUCCESS], $data);
    }
}
