# 开发指南 | Development Guide

## 环境依赖

* Go 1.26+ (`C:\Program Files\Go\bin`)
* Node.js 20+ & npm
* Wails CLI v2.17.1 (`go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.1`)
* NSIS 3.x（编译安装程序时需要，默认路径 `C:\Program Files (x86)\NSIS\makensis.exe`）

## 本地开发运行

```bash
# 安装前端依赖
cd frontend && npm install && cd ..

# 启动 Wails 开发服务器（前端热更新 + Go 后端自动重编译）
wails dev
```

## 测试与静态检查

```bash
# Go 静态检查
go vet ./...

# 运行全量 Go 单元测试
go test -v -count=1 ./internal/...

# 前端类型检查与构建
cd frontend && npm run build
```

## 打包发布产物

在 Windows 命令行运行：

```bat
package.bat
```

脚本将依次执行：前端构建 $\rightarrow$ `wails build -clean -s` $\rightarrow$ 生成便携版 ZIP $\rightarrow$ 编译 `build\installer\installer.nsi` $\rightarrow$ 生成 `bin\checksums.txt`。
