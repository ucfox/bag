<?php
Yaf_Loader::import("/home/q/php/count_sdk/number_sdk.php");
Class CountModel {

    const APP_NAME       = 'pcgame_pandatv';
    private $_expClient;

    public function __construct() {
        $this->_logger   = XLogKit::logger("count");
    }

    public function addExp($uid, $exp, $detail = array()) {
        $env = ($_SERVER["ENV"] == "online") ? $_SERVER["ENV"] : "beta";
        XLogKit::ins('count')->info("[add exp before] uid=$uid&exp=$exp");
        $callerstrace = context::get("callerstrace");
        $callerstrace = $callerstrace ? json_decode($callerstrace,true) : array();
        $this->_expClient = new NumberClient(self::APP_NAME, 'user_exp', $env);
        $data = $this->_expClient->add_value($uid, $exp, $detail, $callerstrace);
        if($data && $data['errno'] == 0) {
            XLogKit::ins('count')->info("[add exp success] uid=$uid&exp=$exp");
            return true;
        }
        XLogKit::ins('count')->error("[add exp fail] ". json_encode($data));
        return false;
    }
}
