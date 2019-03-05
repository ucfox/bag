<?php

abstract class SdkBase
{
    const CALLER = 'bag';

    protected $_logger = null;
    protected $remoteApp = null;

    protected function formatSDKOutput($data) {

        if($data && !$data['errno']) {
            return $data['data'];
        }

        $this->_logger->warn("[from ".$this->remoteApp." sdk] [errno:".$data['errno']."] [errmsg:".$data['errmsg']."]");
        return false;
    }
}
