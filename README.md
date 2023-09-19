# 七牛 qetag

将七牛 kodo etag 算法实现了 Golang 的 `hash.Hash` 接口

* 便于上传/下载文件后的完整性校验
* 实现了 io.Reader 的数据流不用全部读取（下载）就可以计算得到七牛 etag


# 用法

安装
``` bash
go get -u "github.com/br41n10/qetag"
```

简单用法
``` golang
qetag := New()
_, err := qetag.Write([]byte{1, 2, 3, 4, 5, 6, 7})
fmt.Println(qetag.Etag())
```

# 参考
1. https://github.com/qiniu/qetag
2. https://www.jianshu.com/p/3785fc314fc5
