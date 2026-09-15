# go 案例集合

## 参考源
- https://www.bilibili.com/video/BV1or421J7uU?p=4&vd_source=d459674f69afeae159b25c96f15052e1
- https://www.bilibili.com/video/BV1gf4y1r79E?spm_id_from=333.788.player.switch&vd_source=d459674f69afeae159b25c96f15052e1&p=49


## 环境变量： 查询、修改
```sh
go evn # 查询

# 修改
$Env:CGO_ENABLED=0;$Env:GOARCH="amd64";$Env:GOOS="darwin"
```

## 环境编译与运行
```sh
# 1. 编译当前目录
go build
# 或
go build .

# 2. 编译指定文件或目录
go build ./main.go # 编译main文件，生成可执行文件
go build ./test    # 编译指定目录，不产生可执行文件，仅进行编译检查

# 3. 指定编译结果输出(main包有多个文件的情况下，指定文件编译main包)
go build -o ./out/app ./main.go ./hello.go

# 交叉编译
交叉编译需要修改 GOOS、GOARCH、CGO_ENABLED 三个环境变量
GOOS：目标平台操作系统(darwin、freebsd、linux、windows)
GOARCH：目标平台的体系架构32位、64位(386、amd64、arm)
CGO_ENABLED：是否启用CGO，交叉编译不支持CGO所以要禁用

# 例: 编译一个Linux可执行程序
# windows 编译 Linux Mac 可执行程序
# 设置环境变量
$Env:CGO_ENABLED=0;$Env:GOARCH="amd64";$Env:GOOS="linux"
#  执行编译
go build -o ./out/app .

$Env:CGO_ENABLED=0;$Env:GOARCH="amd64";$Env:GOOS="darwin"
go build -o ./out/app .

#  Mac 编译 Linux windows 可执行程序
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o ./out/app .
CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -o ./out/app.exe .

#  Linux 编译 windows Mac 可执行程序
CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -o ./out/app.exe .
CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -o ./out/app .
```

## 依赖管理
```sh
go mod init ${moduleName} # 初始化项目
# tidy 依赖对齐，添加缺少、删除为使用的依赖
go mod tidy

# go.mod: {
#   require: 当前module(项目)依赖包
#   exculde: 排除的三方包
#   replace: 修改依赖包的路径或版本
# }

# 基础命令
go mod download # 下载模块到本地，需要模块路径、版本号
# 例
go mod download github.com/gin-gonic/gin@v1.9.0

# 添加缺少的依赖, 删除未使用的依赖(依赖对齐)
go mod tidy

# 通过工具或脚本编辑go.mod
go mod edit
# 如:
## 添加依赖项
go mod edit -require="github.com/gin-gonic/gin@v1.9.0"
## 替换路径, old[@version] 替换成 new[@version]
go mod edit -replace="golang.org/x/crypto@v0.0.0=github.com/golang/crypto@latest"
## 排除三方依赖的某个版本
go mod edit -exclude="github.com/gin-gonic/gin@v1.9.0"
## 当前项目作为其他项目的依赖时,添加撤回版本用于排除有问题的版本
go mod edit -retract="v1.0.0"
go mod edit -retract="v1.1.0"
## 删除撤回版本记录
go mod edit -dropretract="v1.0.0" 

# 根据go.mod中的依赖项制作vendor副本
# 有了vendor 副本, 项目不再依赖本地缓存
go mod vendor

# 验证依赖是否正确
go mod verify

# 返回对指定模块的依赖关系最短路径, 解释为什么依赖指定包
go mod why
## 如:
go mod why github.com/gin-gonic/gin@v1.9.0


# 2. go install 安装可执行插件
go install github.com/google/gops@latest

# 3. go get 获取模块信息并更新go.mod文件
# 若本地缓存没有该模块, 则下载模块; 有则直接引用
go get github.com/gin-gonic/gin@1.9.0

# go get -u 更新模块依赖, 并更新go.mod
go get -u github.com/gin-gonic/gin@1.9.1

# 4. go clean 清理临时目录中的文件
go clean -modcache
```