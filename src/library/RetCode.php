<?php

Class RetCode {
    const SUCCESS                 = 0;
    const USER_NOT_LOGIN          = 200;
    const SYSTEM_ERROR            = 1001;
    const TOKEN_ERROR             = 1002;
    const GOODS_NOT_EXISTS        = 2001;
    const BAG_DECR_FAIL           = 2002;
    const NUM_NOT_ENOUGH          = 2003;
    const EXPIRE_OUT_OF_DATE      = 2004;
    const GOODS_ADD_FAIL          = 2005;
    const GOODS_NOT_EXISTS_IN_BAG = 3001;
    const GIFT_SELF               = 4001;
    const GUUID_FAIL              = 5001;
    const INVALID_PARAMS          = 7000;
    const PROP_VERSION_UPGRADE    = 8000;
    const PROP_VERSION_FALL_BEHIND= 8001;
    const XY_GOODS                = 108001;
    const ROOM_NOT_SUPPORT        = 108100;
    const PAIMAI_NOT_GIFT         = 108101;
    const FORBID_ACCESS_GOODS     = array(0=>109000, 174=>109174, 285=>109285, 314=>109314);
    const FANS_MEDAL_SUPPORT      = 11000;
    const OCC_LEVEL_FORBIN        = 9000;
    const BOX_LEVEL_FORBIN        = 9001;
    const REMOVE_GRANK_FORBIN     = 9002;
    const REPAIR_GRANK_FORBIN     = 9003;
    const SALVO_FORBIN            = 12001;

    public static $RetMsg = array(
        self::SUCCESS                 => "成功",
        self::USER_NOT_LOGIN          => "账号异常，请重新登录",
        self::SYSTEM_ERROR            => "网络错误，请稍后再试",
        self::TOKEN_ERROR             => "token错误",
        self::GOODS_NOT_EXISTS        => "物品不存在",
        self::BAG_DECR_FAIL           => "物品使用失败，稍后再试",
        self::NUM_NOT_ENOUGH          => "此物品的数量不足",
        self::EXPIRE_OUT_OF_DATE      => "此物品已经过期",
        self::GOODS_ADD_FAIL          => "添加物品失败，稍后再试",
        self::GOODS_NOT_EXISTS_IN_BAG => "此物品已经不在您的背包中",
        self::GIFT_SELF               => "不能给自己送礼物",
        self::GUUID_FAIL              => "获取唯一id错误",
        self::INVALID_PARAMS          => "参数错误",
        self::PROP_VERSION_UPGRADE    => "版本过低，请升级到最新版本尝试",
        self::PROP_VERSION_FALL_BEHIND=> "请更新至最新版本使用此道具",
        self::XY_GOODS                => "该物品只能在星颜房间使用",
        self::ROOM_NOT_SUPPORT        => "移动端还在测试阶段，请移驾去pc端送礼物吧",
        self::PAIMAI_NOT_GIFT         => "主播正在赶来的路上，暂时不能送礼物",
        self::FORBID_ACCESS_GOODS[0]  => "消费姿势错误",
        self::FORBID_ACCESS_GOODS[174]=> "请在弹幕设置中开启",
        self::FORBID_ACCESS_GOODS[285]=> "请在弹幕设置中开启",
        self::FORBID_ACCESS_GOODS[314]=> "请在弹幕设置中开启",
        self::FANS_MEDAL_SUPPORT      => "主播暂未开通粉丝徽章",
        self::OCC_LEVEL_FORBIN        => "当前英雄身份不能使用低阶英雄卡",
        self::BOX_LEVEL_FORBIN        => "贡献卡仅对10级或10级以下宝藏生效",
        self::REMOVE_GRANK_FORBIN     => "您已和该主播分手了，不能再分手啦！",
        self::REPAIR_GRANK_FORBIN     => "您没有和主播分手过，不需要复合哦",
        self::SALVO_FORBIN            => array("请在开启礼炮PK的房间中使用","本场PK我方已使用狂热卡，无法再次使用"),
    );
}
