# banshee
A prometheus push gateway

1.需要给接口 post 如下侧格式的信息 所有 post 的值 需要写在 body 内传输 不要写表单
```
{
  "type":"kv",
  "app":"appname",
  "metric":"api_range",
  "value":"12345",
  "timeout":"60",
  "api":"/",
  "method":"GET",
  "protocol":"HTTP"
}
```


2.必须要有 type、app、metric、value、timeout 字段  除去必须的 kv 外 可以自定义 kv 但不能包含数组或 map
```
{
  "type":"kv",
  "app":"appname",
  "metric":"api_range",
  "value":"12345",
  "timeout":"60",
  "api":"/",
  "method":"GET",
  "protocol":"HTTP"
}
```
其中 type、app、metric、value、timeout 是必须写的 api、method、protocol 是自定义的 也可以写成别的 kv 无名称限制 但只能是 string


3.type 有 kv 和 event 两个选项 目前只支持 kv


4.metric 只允许使用大小写和 _ 下划线 相同 metric 的 json 结构应该相同 如下
```
{
  "type":"kv",
  "app":"l1eng",
  "metric":"api_range",
  "value":"12345",
  "timeout":"60",
  "api":"/",
  "method":"GET",
  "protocol":"HTTP"
}

{
  "type":"kv",
  "app":"l2eng",
  "metric":"api_range",
  "value":"6789",
  "timeout":"120",
  "api":"/s",
  "method":"POST",
  "protocol":"TCP"
}
```

5.所有 kv 的 value 必须是 string 且 value 只能有一个


6.timeout 为采集过期时间（单位为s） 例如 timeoue 值为 60 表示如果 60s 内没有新上报的数据 则会删除条 metric 记录 如果不需要该功能 timeout="" 相同 metric 的 timeout 值应该相同 timeout最小值为 60


样例：
```
curl -XPOST '127.0.0.1:2336/customData/' -d '
{
  "type":"kv",
  "app":"sre",
  "metric":"gatewaytest",
  "value":"12345",
  "timeout":"60",
  "method":"GET",
  "protocol":"HTTP"
}'

curl -XPOST '127.0.0.1:2336/customData/' -d '{
  "type":"kv",
  "app":"sre",
  "metric":"gatewaytest",
  "value":"12345",
  "timeout":"",
  "method":"POST",
  "protocol":"HTTP"
}'
```
