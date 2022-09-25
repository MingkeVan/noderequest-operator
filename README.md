# noderequest-operator


``` shell
RELEASE_VERSION=v1.18.1

 curl -LO https://github.com/operator-framework/operator-sdk/releases/download/$\{RELEASE_VERSION\}/operator-sdk_darwin_amd64

chmod +x operator-sdk_darwin_amd64 && sudo mkdir -p /usr/local/bin/ && sudo cp operator-sdk_darwin_amd64 /usr/local/bin/operator-sdk && rm operator-sdk_darwin_amd64

operator-sdk version

operator-sdk init --domain wukong.io --repo=github.com/MingkeVan/noderequest-operator

operator-sdk create api --group=cache --version=v1alpha1 --kind=NodeRequest --resource --controller

修改 *_types.go 文件后，记得要运行以下命令来为该资源类型生成代码：
make generate

一旦使用 spec/status 字段和 CRD 验证标记定义 API 后，可以使用以下命令生成和更新 CRD 清单
make manifests


本地安装crd到k8s
make install

controller逻辑
通过reconcile 监听pod事件
create patch crd，将数据写入crd中。

本地运行controoler
make run


打包运行controller镜像
make docker-build IMG=mikefan2019/noderequest-operator:v0.0.1

make docker-push IMG=mikefan2019/noderequest-operator:v0.0.1

make deploy IMG=mikefan2019/noderequest-operator:v0.0.1

查看crd 确定字段和数据正确
kubectl get noderequest -oyaml

查看controller
kubectl get pod -n noderequest-operator-system

查看日志
kubectl logs -f noderequest-operator-controller-manager-5b4c8fd544-wjmmx -n noderequest-operator-system
```