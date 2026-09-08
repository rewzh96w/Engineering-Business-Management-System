# Engineering Business Management System

一套面向工程项目商务工作的轻量级本地管理系统，用于统一维护业主合同、分包合同、业主结算和分包结算，并通过经营总览快速了解合同额、结算额及收付款情况。

系统采用 Vue 3 单页应用作为前端、Go 标准库作为后端，数据持久化到本地 JSON 文件。无需数据库，适合单机使用、项目早期试运行、内部演示及小规模数据管理。

## 主要功能

### 经营总览

- 汇总业主合同额与分包合同额
- 展示合同总数及业主、分包合同数量
- 汇总业主已收款和分包已付款金额
- 统计业主结算与分包结算记录数量
- 展示最近更新的合同记录

### 合同管理

- 分别管理业主合同和分包合同
- 新增、编辑和删除合同
- 按编号、名称、相对方或经办人搜索
- 记录合同金额、已结算金额和已收/付款金额
- 记录签订日期、开始日期、结束日期、履约状态、经办人及备注
- 支持“拟签订、履约中、已完工、已结算、已终止”等合同状态

### 结算管理

- 分别管理业主结算和分包结算
- 新增、编辑和删除结算记录
- 按结算编号、合同或相对方搜索
- 记录结算期间、申报金额、审核金额和已收/付款金额
- 关联合同编号、合同名称、相对方、经办人及备注
- 支持“编制中、已申报、审核中、已审核、已支付、已关闭”等结算状态

## 技术架构

| 层级 | 技术 | 说明 |
| --- | --- | --- |
| 前端 | Vue 3、Vite | 响应式单页管理界面 |
| 后端 | Go | 使用标准库 `net/http` 提供 REST API 和静态文件服务 |
| 存储 | JSON 文件 | 默认保存到 `backend/data/business.json` |
| 部署 | 本地单进程 | 前端构建后由 Go 服务统一托管 |

运行时，浏览器访问 Go 服务；页面通过 `/api` 接口读取和修改数据。后端使用读写锁保护内存数据，并在每次写操作后将完整数据保存到 JSON 文件。

## 项目结构

```text
.
├─ backend/
│  ├─ data/
│  │  └─ business.json       # 示例数据及本地数据文件
│  ├─ go.mod                 # Go 模块配置
│  └─ main.go                # API、业务校验、数据存储和静态服务
├─ frontend/
│  ├─ src/
│  │  ├─ App.vue             # 主界面和业务交互
│  │  ├─ main.js             # 前端入口
│  │  └─ *.css               # 页面样式
│  ├─ index.html
│  ├─ package.json
│  ├─ pnpm-lock.yaml
│  └─ vite.config.js
├─ setup.ps1                 # Windows 环境准备脚本
├─ start.ps1                 # Windows 构建与启动脚本
└─ README.md
```

## 环境要求

- Go 1.22 或更高版本
- Node.js 18 或更高版本
- pnpm 9 或更高版本（也可通过 Corepack 启用）
- Windows PowerShell（使用仓库内脚本时）

## 快速启动

先构建前端：

```powershell
pnpm --dir frontend install
pnpm --dir frontend build
```

再从 `backend` 目录启动后端：

```powershell
Set-Location backend
go run .
```

浏览器打开 `http://localhost:8080`。

> 后端默认使用相对路径查找数据和前端构建结果，因此请从 `backend` 目录启动。

### 当前开发机脚本

项目保留了开发时使用的 `setup.ps1` 和 `start.ps1`。脚本引用原开发环境中的便携运行时；在其他电脑上建议使用上面的通用启动方式。

## 配置项

| 变量 | 默认值 | 用途 |
| --- | --- | --- |
| `PORT` | `8080` | HTTP 服务端口 |
| `DATA_FILE` | `./data/business.json` | JSON 数据文件路径 |
| `FRONTEND_DIST` | `../frontend/dist` | 前端构建产物目录 |

示例：

```powershell
$env:PORT = '8090'
$env:DATA_FILE = './data/my-business.json'
go run .
```

## API 说明

### 健康检查与总览

| 方法 | 路径 | 功能 |
| --- | --- | --- |
| `GET` | `/api/health` | 返回服务健康状态 |
| `GET` | `/api/dashboard` | 返回经营总览汇总数据 |

### 合同接口

| 方法 | 路径 | 功能 |
| --- | --- | --- |
| `GET` | `/api/contracts` | 查询合同列表 |
| `POST` | `/api/contracts` | 新增合同 |
| `GET` | `/api/contracts/{id}` | 查询单个合同 |
| `PUT` | `/api/contracts/{id}` | 更新合同 |
| `DELETE` | `/api/contracts/{id}` | 删除合同 |

合同列表支持 `type=owner`、`type=subcontract` 和 `q=关键词` 查询参数。关键词匹配合同编号、名称、相对方和经办人。

### 结算接口

| 方法 | 路径 | 功能 |
| --- | --- | --- |
| `GET` | `/api/settlements` | 查询结算列表 |
| `POST` | `/api/settlements` | 新增结算 |
| `GET` | `/api/settlements/{id}` | 查询单个结算 |
| `PUT` | `/api/settlements/{id}` | 更新结算 |
| `DELETE` | `/api/settlements/{id}` | 删除结算 |

结算列表同样支持 `type` 和 `q` 查询参数。

## 核心数据字段

合同记录包括：合同类型、合同编号、合同名称、合同相对方、合同金额、已结算金额、已收/付款金额、签订及履约日期、状态、经办人、备注和创建/更新时间。

结算记录包括：结算类型、结算编号、关联合同编号及名称、结算相对方、结算期间、申报金额、审核金额、已收/付款金额、状态、经办人、备注和创建/更新时间。

金额字段使用数字存储，界面按人民币格式展示。后端会校验必填字段、业务类型及非负金额。

## 数据与备份

- 仓库中的 `backend/data/business.json` 仅包含演示数据。
- 实际使用前，可删除示例记录或通过 `DATA_FILE` 指向独立的数据文件。
- 系统暂未接入数据库，建议定期备份 JSON 数据文件。
- 若多人同时使用或数据量显著增加，建议迁移到 PostgreSQL、MySQL 等数据库。

## 安全与适用范围

当前版本定位为本地原型系统，尚未实现用户登录、角色权限、操作审计、HTTPS 和数据库事务。请勿直接暴露到公网，也不要在未增加访问控制的情况下存储敏感生产数据。

## 后续规划

- 用户登录与分级权限
- 项目、合同、变更、索赔、计量和支付的完整业务关联
- 附件上传及文档归档
- Excel 导入导出与报表生成
- 数据库持久化和自动备份
- 审批流、操作日志和到期提醒
- 自动化测试与持续集成

## 开发说明

前端开发模式：

```powershell
pnpm --dir frontend install
pnpm --dir frontend dev
```

后端开发模式：

```powershell
Set-Location backend
go run .
```

生产式本地运行前，应先执行 `pnpm --dir frontend build`。生成的 `frontend/dist` 属于可重复构建的产物，不纳入版本控制。

## License

当前仓库尚未指定开源许可证。在添加许可证前，默认保留全部权利。
