<?php
// yaf引导类，初始化app的时候，在启动应用前执行
// 所有以_init开头的函数都会被执行
class Bootstrap extends Yaf_Bootstrap_Abstract
{
    public function _initLoader(Yaf_Dispatcher $dispathcer) {
    }

    /**
     * @brief  初始化日志，基于从Pylon里面抠出来的扩展 pylonlog.so
     */
    public function _initLogger(Yaf_Dispatcher $dispatcher) {
        XLogKit::init($_SERVER['PRJ_NAME'],"",XLogKit::INFO_LEVEL);
    }

    // 默认路由规则，默认就挺好的
    public function _initRoute(Yaf_Dispatcher $dispatcher) {
    }

    // 设置配置文件
    public function _initConfig() {
        Yaf_Dispatcher::getInstance()->disableView();
    }

    // 设置默认路由名，也可以在config.ini中设置
    public function __initDefaultName(Yaf_Dispather $dispather) {
    }

    // 注册插件
    public function _initPlugin(Yaf_Dispatcher $dispatcher) {
        // 注册鉴权插件
		$authPlugin = new AuthPlugin();
		$dispatcher->registerPlugin($authPlugin);
    }
}
