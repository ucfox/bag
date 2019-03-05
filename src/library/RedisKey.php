<?php


class RedisKey
{

    public static function todayGiftMaxNum($giftid)
    {
        return "b|g|max|{$giftid}";
    }


}
