# 🎯 阶段1快速参考卡

## 件位置速查

| 文件 | 路径 | 内容 |
|------|------|------|
| **主启动器** | `connect/cmd/connect-nb67/main.go` | 处理器注册、Connect启动 |
| **处理器实现** | `connect/cmd/connect-nb67/nb67_processor.go` | Processor接口、180+字段映射 |
| **二进制解析** | `connect/cmd/connect-nb67/nb67.go` | Kaitai生成(1936行) |
| **Go配置** | `connect/cmd/connect-nb67/go.mod` | 依赖声明 |
| **Redpanda配置** | `connect/config/nb67-connect.yaml` | Input/Processor/Output + Bloblang |
| **协议定义** | `connect/codec/NB67.ksy` | 二进制结构 + 新增5字段 |
| **Docker镜像** | `connect/Dockerfile.connect` | 多阶段构建 |
| **编排配置** | `deploy/docker-compose.stage1.yml` | 4 services, 3 profiles |
| **部署指南** | `deploy/README-STAGE1.md` | 详细步骤 |
| **验收清单** | `STAGE1-CHECKLIST.md` | 自查表 |
| **总结文档** | `STAGE1-FINAL-SUMMARY.md` | 完整交付说明 |

---

## 🚀 快速启动（5步）

```bash
# 第1步：上传源代码到192.168.32.17
scp -r Macda-Connector user@192.168.32.17:/opt/

# 第2步：SSH到服务器
ssh user@192.168.32.17

# 第3步：构建镜像
cd /opt/Macda-Connector
docker-compose -f deploy/docker-compose.stage1.yml build connect-nb67

# 第4步：启动服务
docker-compose -f deploy/docker-compose.stage1.yml up -d

# 第5步：验证输出
kafka-console-consumer --bootstrap-server 192.168.32.17:19092 \
  --topic signal-parsed --max-messages 1
```

**预期耗时**: 构建3-5分钟，启动10-30秒

---

## 📊 关键数据验证

### 输出JSON示例
```json
{
  "header_code_01": 44,
  "train_no": 100,
  "route_info": {           ← 【新增字段组】
    "start_station": 291,
    "terminal_station": 129,
    "current_station": 45,
    "next_station": 66,
    "exhaust_damper_position": 78
  },
  "timestamp": "2026-02-19T14:35:42Z",
  "device_id": "HVAC-00100-03",
  "environment": { ... },
  "compressor_u11": { ... },
  "alert_level": "OK"
}
```

### 日志采样
```
[NB67] Processed 100 frames: TrainNo=100 Carriage=3 CurStation=45
[NB67] Processed 200 frames: TrainNo=105 Carriage=3 CurStation=60
```

---

## 🔍 实时监控命令

```bash
# 查看容器状态
docker-compose -f deploy/docker-compose.stage1.yml ps

# 查看实时日志
docker-compose -f deploy/docker-compose.stage1.yml logs -f connect-nb67

# 查看输入topic (signal-in)
kafka-console-consumer --bootstrap-server 192.168.32.17:19092 \
  --topic signal-in --max-messages 1 | hexdump -C

# 查看输出topic (signal-parsed)
kafka-console-consumer --bootstrap-server 192.168.32.17:19092 \
  --topic signal-parsed --max-messages 1 | jq .

# 启动监控工具 (可选)
docker-compose -f deploy/docker-compose.stage1.yml --profile monitor up -d
# 访问：http://192.168.32.17:8080

# 消费者群组进度
kafka-consumer-groups --bootstrap-server 192.168.32.17:19092 \
  --group macda-phase1-parser --describe

# 容器资源使用
docker stats macda-connect-nb67
```

---

## ❌ 故障排查流程

### 问题1：镜像构建失败
```bash
# 查看构建日志
docker-compose -f deploy/docker-compose.stage1.yml build --no-cache \
  connect-nb67 2>&1 | tail -100

# 清理并重试
docker rmi macda-connect-nb67 2>/dev/null
docker-compose -f deploy/docker-compose.stage1.yml build connect-nb67
```

### 问题2：容器无法启动
```bash
# 查看详细错误
docker-compose -f deploy/docker-compose.stage1.yml logs connect-nb67

# 检查Kafka连接
docker exec macda-connect-nb67 \
  nc -zv 192.168.32.17 19092
```

### 问题3：无消息输出
```bash
# 检查input topic是否有数据
kafka-console-consumer --bootstrap-server 192.168.32.17:19092 \
  --topic signal-in --max-messages 1

# 检查消费者进度
kafka-consumer-groups --bootstrap-server 192.168.32.17:19092 \
  --group macda-phase1-parser --describe

# 检查输出topic是否存在
kafka-topics --bootstrap-server 192.168.32.17:19092 \
  --list | grep signal-parsed

# 如果topic不存在，创建它
kafka-topics --bootstrap-server 192.168.32.17:19092 \
  --create --topic signal-parsed \
  --partitions 3 --replication-factor 1
```

---

## 📋 性能指标速查

| 指标 | 方法 | 预期值 |
|------|------|--------|
| **吞吐量** | `grep "\[NB67\]" logs` + 计算fps | >1000 msg/s |
| **延迟** | 时间戳对比 | <100ms (p99) |
| **CPU** | `docker stats` | <200% (2core) |
| **内存** | `docker stats` | <1GB |
| **容器大小** | `docker images` | ~150-200MB |

---

## 🔐 环境变量配置

在`docker-compose.stage1.yml`中修改：

```yaml
environment:
  REDPANDA_BROKERS: "192.168.32.17:19092"      # Kafka地址
  INPUT_TOPIC: "signal-in"                      # 输入topic
  OUTPUT_TOPIC: "signal-parsed"                 # 输出topic
  LOG_LEVEL: "info"                             # debug/info/warn/error
  CONSUMER_GROUP: "macda-phase1-parser"         # 消费者组ID
```

---

## 📬 新增5个字段详解

| 字段 | 偏移 | 类型 | 范围 | 说明 |
|-----|------|------|------|------|
| `dmp_exu_pos` | 452-453 | U16 | 0-100 | 废排风阀开度百分比 |
| `start_station` | 454-455 | U16 | 0-999 | 起始站（枚举值，见docs/requirements/binary-spec.md） |
| `terminal_station` | 456-457 | U16 | 0-999 | 终点站（枚举值） |
| `cur_station` | 458-459 | U16 | 0-999 | 当前站（枚举值） |
| `next_station` | 460-461 | U16 | 0-999 | 下一站（枚举值） |

所有字段都在JSON输出中的 `route_info` 对象中。

---

## 📞 关键联系点

| 主题 | 参考文档 |
|------|---------|
| 12周计划 | `docs/11-macda-refactor-execution-plan.md` |
| 二进制规范 | `docs/requirements/binary-spec.md` |
| Redpanda Connect | https://docs.redpanda.com/redpanda-connect |
| Bloblang语言 | https://docs.redpanda.com/redpanda-connect/bloblang |
| Kaitai结构 | https://kaitai.io/ |
| 任务跟踪 | `AGENTS.md` (使用`bd`命令) |

---

## ✅ 验收清单（部署前）

- [ ] 所有源文件已上传
- [ ] `connect/cmd/connect-nb67/go.mod` 存在
- [ ] `connect/config/nb67-connect.yaml` 包含180+字段
- [ ] `connect/codec/NB67.ksy` 包含新增5字段
- [ ] `Dockerfile.connect` 包含multi-stage构建

## ✅ 验收清单（部署后）

- [ ] `docker-compose build` 执行成功
- [ ] `docker-compose up` 启动无错
- [ ] 日志显示 `[NB67] Processed XXX frames`
- [ ] signal-parsed topic 有JSON输出
- [ ] 输出JSON包含 `route_info` 对象
- [ ] 吞吐量 > 1000 msg/s
- [ ] 容器内存 < 1GB
- [ ] 无错误日志

---

**快速反馈**: 如遇问题，检查：
1. 网络连通性: `nc -zv 192.168.32.17 19092`
2. topic存在性: `kafka-topics --list`
3. 消费进度: `kafka-consumer-groups --describe`
4. 容器日志: `docker logs -f macda-connect-nb67`
