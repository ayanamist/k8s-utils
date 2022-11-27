本仓库包含了一些k8s操作的增强小工具

# streamlister

以stream的方式进行list，逐个返回结果，降低全量list时的内存占用

# retrywatcher

在 https://pkg.go.dev/k8s.io/client-go@v0.25.4/tools/watch#RetryWatcher 的基础上，增加了空闲检测，避免虚假连接状态的watch