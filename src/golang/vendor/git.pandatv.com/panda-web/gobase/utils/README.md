参数解析工具

支持 php 特有的request参数 格式解析 如  form[username]="ssssss"等 相同开头的 参数 会解析到一个跟下
支持 因兼容而产生参数格式问题 请求 ，如 room ，roomid，roomId 等


benchmark utils:
go test git.pandatv.com/panda-web/gobase/utils -bench=. -benchtime=3s

test utils:
go test git.pandatv.com/panda-web/gobase/utils -run TestHash -v
go test git.pandatv.com/panda-web/gobase/utils -run TestGetShard -v
