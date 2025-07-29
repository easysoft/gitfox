###  📌 GitFox 简介
**[GitFox](https://www.gitfox.net/)** 是一款专注于企业研发协同的Git源代码管理平台 ，提供从代码托管、流水线构建，到质量扫描、制品管理的一站式能力，旨在帮助企业高效CI&CD。

### 🔗 与 Gitness 的关系
**GitFox** 基于 [Gitness](https://github.com/harness/harness) 深度定制开发，聚焦源代码管理、CI/CD 流程、质量治理与协同效率提升。GitFox，面向开源社区持续演进。

GitFox 由 [禅道 DevOps](https://www.zentao.net/solution-devops.html) 团队主导开发，遵循 Apache 2.0 开源协议。GitFox 保留核心架构基础，同时进行了功能扩展与界面优化，旨在满足国内企业的研发管理实践需求。
###  🔍  Gitness 是什么
Gitness 是 [Harness](https://www.harness.io) 开源的现代化 Git 服务器与 DevOps 平台，具备代码托管、权限管理、流水线构建等功能，支持 Web UI 与 Git CLI 操作，适用于构建现代团队的协作开发平台。

其目标是提供一个轻量、高效、可集成的 DevOps 栈核心组件，具备高度可扩展性，受到大量开源社区用户的关注与认可。

### 🌟  4 GitFox 对上游的贡献与改动说明
GitFox 始终坚持与上游 Gitness 保持**开放协作与共创关系**，在遵循原始许可证（Apache 2.0）的前提下，持续向上游反馈改进建议和优化实践，具体包括：
- **功能增强**：扩展了分支策略模板、代码评审流程、质量门禁等企业级关键能力；
- **本地化适配**：优化了界面语言、日期格式、权限模型等内容，以提升国内团队使用体验；
- **插件与集成能力拓展**：增加与禅道项目管理、CI 工具、制品仓库的联动插件，构建端到端 DevOps 流程；
- **安全与合规性改进**：引入更完善的审计日志机制与权限分级控制，以满足政企用户需求；
- **上游回馈**：针对核心性能、权限逻辑、UI交互等问题向 Gitness 社区提交多项 Issue 与 Pull Request。 
---
### 📄 授权协议说明

本项目基于 Apache 2.0 协议，更多第三方扩展详见 `LICENSE` 文件说明。

---
### 🚀 快速开始
```bash
# 获取代码
git clone https://github.com/easysoft/gitfox
cd gitfox

# 启动服务（Docker Compose 示例）
docker compose up -d
```
