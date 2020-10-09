
#uuid demo

### 基于snowflake 实现的uuid服务，支持redis协议，通过get请求

###php测试代码
```php

<?php

$redis = new \Redis();
$redis->pconnect('docker.for.mac.host.internal', '9999', 1);
$d = [];
$startTime = time();
echo time() . PHP_EOL;
for ($i = 0; $i < 10000; $i++) {
    $uuid = $redis->get('2');
    $d[] = $uuid;
};
echo time() - $startTime . PHP_EOL;
$uniqList = array_unique($d);
var_dump(count($uniqList));
if (count($uniqList) != 10000) {
    var_dump(array_diff($uniqList, $d));
}
$redis->close();

```

### todo
* 支持mc协议
* 部分代码优化，启动时，初始化好全部uuidServer，然后请求的workid，下发对应的uuidServer