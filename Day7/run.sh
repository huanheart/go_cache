#!/bin/bash
#删除对应内容
trap "rm server;kill 0" EXIT

go build -o server
./server -port=8001 &
./server -port=8002 &
./server -port=8003 -api=1 &

#睡眠两秒，表示服务初始化完毕
sleep 2
#输出到终端，表示现在开启测试
echo ">>> start test"
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &

#等待后台所有进程完成，即保证curl请求全部完成
wait