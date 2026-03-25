# GoDaddy DNS Sync API (MySQL)

这是一个按你最新要求整理的 Go 项目：

- 主域名 **不支持手动添加**
- 主域名全部来自 **GoDaddy 账号下的域名列表**
- 服务会 **定时同步 GoDaddy 主域名到本地 MySQL**
- “禁用主域名” 只是在 **本地数据库打标**，不会删除 GoDaddy 域名
- 拉取可用主域名时，只返回本地 `active + is_available=1` 的记录
- 所有业务接口都要求 `api_token`
- 增加 GoDaddy 请求节流
- GoDaddy 本地节流命中或 GoDaddy 返回 `429` 时，接口 **直接报错返回**

## 依赖

- Go 1.22+
- MySQL 8+
- Go MySQL Driver：`github.com/go-sql-driver/mysql`

## 说明 

这个包已经整理成完整项目结构，但因为当前离线环境无法下载第三方依赖，所以我没法在这里真实编译 `github.com/go-sql-driver/mysql`。你在本地联网环境执行 `go mod tidy && go build ./cmd/server` 即可。

GoDaddy 官方说明：
- 生产环境基地址：`https://api.godaddy.com`
- 测试环境基地址：`https://api.ote-godaddy.com`
- 每个 endpoint 默认限制 60 requests/minute。 citeturn0search0

## 初始化

1. 创建数据库：

```sql
CREATE DATABASE adnx_dns CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 导入表结构：

```bash
mysql -uroot -p adnx_dns < schema.sql
```

3. 复制环境变量：

```bash
cp .env.example .env
```

4. 修改 `.env`：

```env
APP_ENV=dev
HTTP_ADDR=:8080
API_TOKEN=replace_with_your_api_token
MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/adnx_dns?parseTime=true&charset=utf8mb4&loc=Local
GODADDY_BASE_URL=https://api.godaddy.com
GODADDY_API_KEY=your_godaddy_api_key
GODADDY_API_SECRET=your_godaddy_api_secret
DOMAIN_SYNC_INTERVAL_SECONDS=300
GODADDY_REQUEST_TIMEOUT_SECONDS=15
GODADDY_RATE_LIMIT_PER_MINUTE=60
RANDOM_SUBDOMAIN_LENGTH=8
```

5. 启动：

```bash
go mod tidy
go run ./cmd/server
```

## 接口

所有请求都需要带：

```http
X-API-Token: your_api_token
```

也支持 query 参数：

```text
?api_token=your_api_token
```

### 1. 拉取所有可用主域名

```http
GET /api/v1/domains
```

返回的是本地库里：
- 已从 GoDaddy 同步到本地
- 未被本地禁用
- 当前可用

示例：

```bash
curl -H "X-API-Token: replace_with_your_api_token" \
  http://127.0.0.1:8080/api/v1/domains
```

### 2. 手动触发同步 GoDaddy 主域名

```http
POST /api/v1/domains/sync
```

这个接口会：
- 调用 GoDaddy `GET /v1/domains`
- 把域名同步到本地 `domains`
- 对已被本地禁用的域名，保持 `disabled`
- 对 GoDaddy 已不存在的域名，标记为 `missing`

### 3. IP 绑定域名

```http
POST /api/v1/bind
Content-Type: application/json
```

请求体：

```json
{
  "ipv4": "1.2.3.4",
  "subdomain": "hkabc",
  "domain": "example.com"
}
```

说明：
- `ipv4` 必填
- `subdomain` 可选，不传则自动生成随机小写字母
- `domain` 可选，不传则自动从本地可用主域名里选一个
- `subdomain` 只允许小写字母

规则：
- 如果同一个 `fqdn` 已绑定相同 IP，直接返回 `already bound`
- 如果同一个 `fqdn` 已存在但 IP 不同，则切换到新 IP
- 如果该 IP 已绑定别的子域名，则先解绑旧记录，再绑定新记录
- 如果没有可用主域名，返回错误
- 如果 GoDaddy 节流或返回 429，直接返回错误

示例：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/bind \
  -H "X-API-Token: replace_with_your_api_token" \
  -H "Content-Type: application/json" \
  -d '{"ipv4":"1.2.3.4","subdomain":"hkabc","domain":"example.com"}'
```

### 4. 禁用主域名（仅本地）

```http
POST /api/v1/domains/{id}/disable
```

这个动作只更新本地表：
- `sync_status = disabled`
- `is_available = 0`

不会操作 GoDaddy。

### 5. 解除子域名 IP 绑定

```http
POST /api/v1/unbind
Content-Type: application/json
```

两种传参任选其一：

按 IP：

```json
{
  "ipv4": "1.2.3.4"
}
```

按子域名：

```json
{
  "subdomain": "hkabc",
  "domain": "example.com"
}
```

服务会：
- 先在本地找到 active 绑定
- 将本地状态改为 released
- 调用 GoDaddy 删除 A 记录

## 设计说明

### domains 表

- `source`: 当前固定 `godaddy`
- `sync_status`: `active / disabled / missing`
- `is_available`: 是否参与“可用主域名”筛选

### ip_bindings 表

保存当前子域名与 IP 的关系：
- 一个 `fqdn` 只能对应一个 active 记录
- 一个 `ipv4` 也只能对应一个 active 记录

## GoDaddy 相关接口

项目里实际使用的是：

- `GET /v1/domains`
- `PUT /v1/domains/{domain}/records/A/{name}`
- `DELETE /v1/domains/{domain}/records/A/{name}`

这些都属于 GoDaddy Domains / DNS 管理接口。 citeturn0search0
