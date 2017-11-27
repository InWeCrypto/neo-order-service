---
weight: 10
title: API Reference
---



# 介绍

NEO订单管理微服务，提供以下功能：

1. 用户钱包注册，用于钱包交易推送（无钱包管理功能）；
2. 订单管理，包括：获取订单列表,创建订单，查询订单状态以及状态推送；



# SDK 本地接口

## 创建新钱包

> 创建新的NEO钱包:

```java
package com.inwecrypto.test

public class App {
    public static void main(String args[]) {
        neomobile.Wallet neowallet = neomobile.new_();
    }
}
```
```objc
    
```

#### 请求参数


Parameter | Default | Description
--------- | ------- | -----------



## 通过WIF字符串创建钱包

> 通过WIF字符串创建钱包:

```java
package com.inwecrypto.test

public class App {
    public static void main(String args[]) {
        neomobile.Wallet neowallet = neomobile.fromWIF("xxxxxx");
    }
}
```

### 请求参数


Parameter | Type | Description
--------- | ---- | -----------
wif | string | WIF字符串


## 通过读取web3 keystore字符串创建钱包

> 读取keystore:

```java
package com.inwecrypto.test

public class App {
    public static void main(String args[]) {
        neomobile.Wallet neowallet = neomobile.fromKeyStore("xxxxxx","xxxxx");
    }
}
```
> keystore的格式类似下面的json代码：

```json
{
    "version": 3,
    "id": "1b3cb7fc-306f-4ec3-b753-831cb9e18984",
    "address": "00e773ad3fa1481bc4222277f324d57f35f06b60",
    "crypto": {
        "ciphertext": "423abc4fea2f1f58543b456f9a67f60eb7be076a79471d284d2777c1ce5ee2cd",
        "cipherparams": {
            "iv": "a737c70e4541a9eb053a49e9103d7ccc"
        },
        "cipher": "aes-128-ctr",
        "kdf": "pbkdf2",
        "kdfparams": {
            "dklen": 32,
            "salt": "0de73ba9540afa6424f05d159575a665da3aefa751cd7c56fc5dd87aeac4ea6b",
            "c": 65536,
            "prf": "hmac-sha256"
        },
        "mac": "53b2cf7744e91d2718f9aa586dac8c5fb647b3b912b5c4ba3c1eafd7a99346e3"
    }
}
```

### 请求参数


Parameter | Type | Description
--------- | ---- | -----------
keystore | string | keystore json 字符串
password | string | keystore 秘钥

## 通过助记词创建钱包

> 读取助记词:

```java
package com.inwecrypto.test

public class App {
    public static void main(String args[]) {
        neomobile.Wallet neowallet = neomobile.fromMnemonic("xxxxxx");
    }
}
```

### 请求参数


Parameter | Type | Description
--------- | ---- | -----------
mnemonic | string | 空格分割的助记词字符串

## 转账

> 创建钱包并转账:

```java
package com.inwecrypto.test

public class App {
    public static void main(String args[]) {
        neomobile.Wallet neowallet = neomobile.fromMnemonic("xxxxxx");

        neowallet.createAssertTx("xxxx","from","to","",1,"xxxxxx")
    }
}
```

> unspent 参数示例：

```json
[{
	"txid": "0x07537a82ef57610931cffe3e31ca5f38dbadc9daa2c2c881a49d9a984539ed06",
	"vout": {
		"Address": "AMpupnF6QweQXLfCtF4dR45FDdKbTXkLsr",
		"Asset": "0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b",
		"N": 0,
		"Value": "1"
	},
	"createTime": "2017-11-23T12:58:31Z",
	"spentTime": "",
	"block": 809098,
	"spentBlock": -1
},{
        "txid": "0x07537a82ef57610931cffe3e31ca5f38dbadc9daa2c2c881a49d9a984539ed06",
	"vout": {
		"Address": "AMpupnF6QweQXLfCtF4dR45FDdKbTXkLsr",
		"Asset": "0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b",
		"N": 0,
		"Value": "1"
	},
	"createTime": "2017-11-23T12:58:31Z",
	"spentTime": "",
	"block": 809098,
	"spentBlock": -1
}]
```

### 请求参数


Parameter | Type | Description
--------- | ---- | -----------
assert | string | 资产类型字符串
from | string | 转账源地址
to | string | 转账目标地址
amount | string | 转账金额
unspent | string | 转账源地址所持有的未花费的utxo列表，json格式


### 返回值


Parameter | Type | Description
--------- | ---- | -----------
tx | object | 包含 txid 以及 raw tx string


## 获取Gas

> Claim Gas:

```java
package com.inwecrypto.test

public class App {
    public static void main(String args[]) {
        neomobile.Wallet neowallet = neomobile.fromMnemonic("xxxxxx");

        neowallet.createClaimTx(1,"address","xxxxxx")
    }
}
```

> unspent 参数示例：

```json
[{
    "txid": "0x07537a82ef57610931cffe3e31ca5f38dbadc9daa2c2c881a49d9a984539ed06",
    "vout": {
        "Address": "AMpupnF6QweQXLfCtF4dR45FDdKbTXkLsr",
        "Asset": "0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b",
        "N": 0,
        "Value": "1"
    },
    "createTime": "2017-11-23T12:58:31Z",
    "spentTime": "2017-11-24T06:38:04Z",
    "block": 809098,
    "spentBlock": 811513
}, {
    "txid": "0x07537a82ef57610931cffe3e31ca5f38dbadc9daa2c2c881a49d9a984539ed06",
    "vout": {
        "Address": "AMpupnF6QweQXLfCtF4dR45FDdKbTXkLsr",
        "Asset": "0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b",
        "N": 0,
        "Value": "1"
    },
    "createTime": "2017-11-23T12:58:31Z",
    "spentTime": "2017-11-24T06:38:04Z",
    "block": 809098,
    "spentBlock": 811513
}]
```

### 请求参数


Parameter | Type | Description
--------- | ---- | -----------
amount | string | 转账金额
address | string | 转账目标地址
unspent | string | unclaimed utxo 列表，json格式





### 返回值


Parameter | Type | Description
--------- | ---- | -----------
tx | object | 包含 txid 以及 raw tx string




## 获取地址

> get neo address:

```java
package com.inwecrypto.test

public class App {
    public static void main(String args[]) {
        neomobile.Wallet neowallet = neomobile.fromMnemonic("xxxxxx");

        neowallet.address()
    }
}
```

### 请求参数

### 返回值


Parameter | Type | Description
--------- | ---- | -----------
address | string | neo 地址


## 生成助记词

> generate mnemonic:

```java
package com.inwecrypto.test

public class App {
    public static void main(String args[]) {
        neomobile.Wallet neowallet = neomobile.fromMnemonic("xxxxxx");

        neowallet.Mnemonic()
    }
}
```

### 请求参数

### 返回值


Parameter | Type | Description
--------- | ---- | -----------
Mnemonic | string | neo keystore 助记词