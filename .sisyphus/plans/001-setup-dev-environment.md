# Plan: 搭建开发环境

## 目标
为 Flux Panel 项目安装所有缺失的开发依赖，使 3 个子项目都能本地开发和构建。

## 当前状态

### ✅ 已安装
| 工具 | 版本 | 用途 |
|------|------|------|
| Node.js | v20.19.2 | vite-frontend |
| npm | 9.2.0 | vite-frontend |
| Go | 1.24.4 | go-gost |
| Docker | 29.1.4 | 容器化部署 |

### ❌ 缺失
| 工具 | 需求版本 | 用途 |
|------|----------|------|
| Java | 21 | springboot-backend |
| Maven | 3.x | 构建后端 |
| Docker Compose | v2 | 容器编排 |

---

## 执行任务

### Task 1: 安装 Java 21
```bash
apt-get update && apt-get install -y openjdk-21-jdk
```
**验证**: `java -version` 应显示 openjdk 21

### Task 2: 安装 Maven
```bash
apt-get install -y maven
```
**验证**: `mvn -v` 应显示 Maven 3.x

### Task 3: 安装 Docker Compose Plugin
```bash
apt-get install -y docker-compose-plugin
```
**验证**: `docker compose version` 应显示版本号

### Task 4: 安装前端依赖
```bash
cd /root/flux-panel/vite-frontend && npm install
```
**验证**: `node_modules/` 目录存在

### Task 5: 验证后端可构建
```bash
cd /root/flux-panel/springboot-backend && mvn clean compile -q
```
**验证**: 编译成功无错误

### Task 6: 验证 Go 模块
```bash
cd /root/flux-panel/go-gost && go mod download
```
**验证**: 依赖下载成功

---

## 完成标准
- [ ] `java -version` → openjdk 21
- [ ] `mvn -v` → Maven 3.x
- [ ] `docker compose version` → v2.x
- [ ] 前端: `npm run dev` 可启动
- [ ] 后端: `mvn compile` 成功
- [ ] Go: `go build .` 成功
