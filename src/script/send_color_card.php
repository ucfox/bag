<?php

// 刚上线，产品让给一些等级高的用户免费送一张彩色弹幕卡（三天有效期)
// 等级高的用户（数据来自柏慕海)
$data = file_get_contents("/home/zhaorui/uid.txt");
$data = explode(",", $data);
foreach($data as $k => $uid) {
    $uid = trim($uid);
    $url = "http://beta.ibag.pdtv.io:8360/bag/add";
    $post_data = array(
        "app"      => "pandaren",
        "uid"      => $uid,
        "goods_id" => 3, // 这里写死
        "num"      => 1,
        "_caller"  => "bag",
    );
    // 存背包
    $output = send_post($url, $post_data);
    if (empty($output) || $output["errno"] != 0) {
        echo "fail_color:". $uid."\n";
    } else {
        echo "succ_color:". $uid."\n";
    }
    $url = "http://message.pdtv.io:8360/Message/sendMessageToUser";
    $post_data = array(
        "title" => "尊贵的熊猫用户您好",
        "cat" => "1",
        "to_uid" => $uid,
        "content" => "感谢您一直以来对熊猫直播平台的支持，我们已将彩色弹幕特权卡发送至您的背包，赶紧去畅所欲言体验不一样的弹幕文化吧！",
        "_caller" => "bag",
    );
    // 发送站内信
    $output = send_post($url, $post_data);
    if (empty($output) || $output["errno"] != 0) {
        echo "fail_msg:". $uid."\n";
    } else {
        echo "succ_msg:". $uid."\n";
    }
}

function send_post($url, $post_data) {
    $ch = curl_init();

    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
    // post数据
    curl_setopt($ch, CURLOPT_POST, 1);
    // post的变量
    curl_setopt($ch, CURLOPT_POSTFIELDS, $post_data);

    $output = curl_exec($ch);
    if (curl_errno($ch)) {
        return false;
    }
    curl_close($ch);

    return json_decode($output, true);
}
