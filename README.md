# Game Trainer Hub (PC 游戏修改器聚合管理客户端)

Game Trainer Hub 是一款基于 Windows 平台的现代 PC 游戏修改器聚合管理客户端。以 FLiNG Trainer 为首发资源站点，通过适配器模式实现多源聚合，具备中英双向智能搜索、HTTP Range 分块断点续传、本地修改器生命周期管理及安全隔离启动能力。

## 核心特性

- **中英双向与拼音搜索引擎**：支持中文查英文、英文查中文、常用简称/别名（如“老头环”、“怪猎世界”、“黑猴”）、全拼与声母拼音索引。
- **本地与在线混合检索**：毫秒级本地 SQLite FTS5 全文索引，无缝融合 FLiNG 原网站实时检索。
- **高可用下载管理器**：支持多任务并发、HTTP Range 范围断点续传、SHA-256 完整性校验、实时测速与异常持久化恢复。
- **本地修改器与版本库**：支持多版本管理、下载历史回溯、一键定位目录及收藏。
- **安全启动机制**：Windows 进程隔离与权限限制，严禁静默提权，启动前提示安全风险与来源哈希。
- **多站点适配器架构**：网站抓取与业务逻辑完全解耦，未来扩展第三方修改器站点只需新增 Adapter。

## 项目结构

```text
GameTrainerHub/
├── frontend/                     # Vue 3 + TypeScript + Element Plus 前端工程
│   ├── src/                      # UI 页面与业务逻辑组件
│   ├── dist/                     # 前端编译生成产物
│   └── package.json
├── internal/                     # Go 后端核心业务层
│   ├── app/                      # Wails IPC 应用服务入口
│   ├── config/                   # 配置持久化与路径解析
│   ├── database/                 # SQLite 连接池、FTS5 全文检索引擎与仓储
│   ├── download/                 # Range 断点续传调度器与 Worker 池
│   ├── game/                     # Steam 游戏扫描与本地库管理
│   ├── launcher/                 # Windows 进程安全启动器
│   ├── logger/                   # 日志轮转与结构化记录
│   ├── model/                    # 全局领域数据实体定义
│   ├── search/                   # 混合搜索与去重排序
│   ├── source/                   # 适配器抽象与 FLiNG 抓取器
│   ├── sync/                     # 定期与增量同步调度
│   ├── trainer/                  # 本地修改器版本管理
│   └── translate/                # 中英文游戏词典与拼音引擎
├── main.go                       # 应用程序启动入口 (Wails + Webview2)
├── wails.json                    # Wails 构建配置
└── GameTrainerHub.exe            # 独立 Windows 可执行程序 (13.7 MB)
```

## 编译与测试

```bash
# 运行全部单元测试与集成测试
go test -v -count=1 ./internal/...

# 前端构建
cd frontend
npm run build
cd ..

# 构建独立客户端
go build -ldflags "-s -w" -o GameTrainerHub.exe main.go
```

