Go 版派聪明（PaiSmart-Go）是一个企业级的 AI 知识库管理系统，采用 RAG 技术提供智能文档处理和检索能力。核心技术栈包括：

- Go 1.23+、模块化目录：`cmd/` `internal/` `pkg/`；分层：`handler/service/repository`
- 配置/日志/关停：Viper、Zap（结构化日志）、Gin + Context 优雅停机
- Gin（路由分组/中间件）、Gorilla WebSocket（双向通信、增量写出、停止指令）
- JWT（access/refresh）、基于 `org_tag` 的层级聚合，检索期过滤（should + minimum_should_match）
- MySQL 8 + GORM（文件/分片/向量等元数据持久化）、Redis 7（分片进度与重试计数）
- MinIO：分片对象存储；单分片 Copy、多分片 Compose；合并后后台清理分片对象
- Kafka（segmentio/kafka-go）：生产/消费、失败阈值重试、手动提交 offset
- 任务解耦：`TaskProcessor` 接口承载解析/向量化/索引流水线
- Apache Tika（HTTP 服务）：PDF/DOCX/PPT/XLS 等文本抽取
- 分块策略：固定窗口 + 重叠切分（提升语义覆盖）
- Elasticsearch 8：KNN 语义召回 + BM25 rescore + 短语兜底 should；索引含 `userId/orgTag/isPublic`
- Embedding：OpenAI 兼容协议，已适配 DashScope（维度 2048 可配）
- LLM：DeepSeek Chat 流式；可按同协议切换本地 Ollama
- Docker 容器化：一键拉起 MySQL/Redis/ES/Kafka/MinIO/Tika
- 集中管理 LLM/Embedding/ES 等参数

它的目标是帮助企业和个人更高效地管理和利用知识库中的信息，支持多租户架构，允许用户通过自然语言查询知识库，并获得基于自身文档的 AI 生成响应。

![派聪明的前后端](https://cdn.tobebetterjavaer.com/stutymore/README-20251027092633.png)

系统允许用户：

- 上传和管理各种类型的文档
- 自动处理和索引文档内容
- 使用自然语言查询知识库
- 接收基于自身文档的 AI 生成响应



## 后端启动

可 Docker 容器化一键部署前置环境，教程见：[派聪明环境部署教程](https://paicoding.com/column/10/29)。

![Docker 拉取前置环境](https://cdn.tobebetterjavaer.com/stutymore/README-20251027093622.png)

然后找到 main.go 文件，点击运行应用程序的按钮即可启动后端服务，默认监听 8081 端口。

![后端启动](https://cdn.tobebetterjavaer.com/stutymore/README-20251027093531.png)


## 前端启动

```bash
# 进入前端项目目录
cd frontend

# 安装依赖
pnpm install

# 启动项目
pnpm run dev
```

聊天助手的访问效果如下图所示：

![Go 版派聪明的运行后效果](https://cdn.tobebetterjavaer.com/paicoding/754665f76be3ff5b0a65b684377a4d1e.png)