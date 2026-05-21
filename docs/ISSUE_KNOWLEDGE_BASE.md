# Macda-Connector Issue 处理知识库

本文件是全栈开发智能体的**持续记忆文档**。
每次处理 GitHub Issue 前必须阅读，每次完成修复后必须更新。

---

## 使用规范

### 处理 Issue 前（必读）
1. 阅读「已知约束与注意事项」（第一节）
2. 搜索「历史记录索引」（第二节），查找相似 Issue
3. 如找到相似记录，跳转到对应历史条目阅读根因和修复思路

### 完成修复后（必写）
在「历史 Issue 处理记录」（第三节）**顶部**追加本次记录，并更新第二节索引。

---

## 一、已知架构约束与注意事项

> 这里记录项目特有的、容易踩坑的约束，每次发现新坑时追加。

### NB67 协议解析
- `connect/codec/NB67.ksy` 是唯一真实来源，`nb67.go` 为自动生成文件，**禁止手工修改**
- 任何字段变更必须走完整链路：`NB67.ksy` → 重新生成 → `nb67_processor.go` → 存储映射 → API 契约 → 文档

### 时间字段
- `RUNTIME=DEV`（当前所有环境默认）时用 `ingest_time`，`RUNTIME=PRD` 时用 `event_time`
- 切换前必须确认现场设备时钟精度，否则历史数据时序会乱

### 前端构建
- 镜像构建前必须先本地 `npm run build` 验证通过，不允许直接 docker build

### 流水线配置
- `nb67-connect.yaml` 已废弃，勿修改；主线是三段式配置
- 扩容 `connect-parser` 不超过 Redpanda partition 数量

---

## 二、历史记录索引

> 快速检索用。格式：`[日期] #Issue编号 - 问题关键词 - 所属模块`

<!-- 新记录追加到此处下方 -->

---

## 三、历史 Issue 处理记录

> 最新记录在最前。每条记录包含：问题描述、根因、修复方法、测试验证、经验总结。

<!-- 新记录追加到此处下方 -->

---

## 四、测试方法速查

### NB67 解析验证
```bash
bash connect/tests/test-nb67-parsing.sh
```

### 前端构建验证
```bash
cd web-nb67-web && npm run build
```

### BFF 接口验证
```bash
cd web-nb67-bff && npm run dev
# 另开终端：curl http://localhost:PORT/api/...
```

### 完整环境启动
```bash
cd baseEnv
docker compose -f docker-compose-mock.yml up -d   # 先建网络+数据源
docker compose -f docker-compose-Dev.yml up -d    # 再拉全栈
```

---

## 五、常见问题 FAQ

> 高频出现的问题，处理完 Issue 后如发现有规律可总结则追加。

<!-- 新条目追加到此处 -->
