#!/bin/bash
RUN_PATH=$PRJ_ROOT/run/$APP_SYS/

case $1 in
    config)
        echo -n "Exporting Lua ENV def "
        env | grep =  > $RUN_PATH/env.lua
        echo "done"
        exit;
        ;;
    data)
        exit;
        ;;
    start)
        exit;
        ;;
    stop)
        exit;
        ;;
esac
