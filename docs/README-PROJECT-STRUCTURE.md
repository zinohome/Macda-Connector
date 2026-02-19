# Macda-Connector 项目结构说明

**最后整理时间**: 2026年2月19日

## 📂 项目结构概览

```
Macda-Connector/
├── README.md              ← 项目总说明（必读）
├── AGENTS.md              ← Agent操作指南
│
├── baseEnv/               ← 基础环境配置
│   ├── docker-compose-*.yml
│   └── redpanda-labs/     ← Redpanda参考实现
│
├── connect/               ← ✨ 核心项目目录
│   ├── cmd/
│   │   └── connect-nb67/  ← Go应用主程序
│   ├── codec/             ← NB67协议解析器
│   ├── config/            ← Connect配置文件
│   └── tests/             ← ✨ 测试脚本目录
│       ├── test-kafka-connection.sh
│       ├── test-nb67-parsing.sh
│       └── test-end-to-end.sh
│
├── docs/                  ← ✨ 所有文档统一存放
│   ├── README.md          ← 文档总索引
│   ├── CLEANUP-SUMMARY.md
│   ├── EXECUTIVE-SUMMARY.md
│   ├── FINAL-OPTIMIZED-CONFIG.md
│   ├── RESOURCE-QUICK-REF.md
│   ├── SOLUTION-COMPARISON.md
│   ├── stage1/            ← Phase 1阶段文档
│   │   ├── README.md
│   │   ├── STAGE1-*.md
│   │   └── ...
│   ├── requirements/      ← 需求文档
│   │   └── binary-spec.md
│   └── 01-10-*.md         ← 其他文档
│
├── oldproj/               ← 已废弃的项目参考
│   ├── MACDA-NB67/
│   └── web-nb67.250513/
│
└── temp/                  ← ✨ 临时文件存放区
    └── .gitkeep
```

## 🎯 关键改动

### ✅ 已完成（Phase 1整理）

1. **文档统一管理** 
   - ✅ 所有文档→docs/目录
   - ✅ CLEANUP-SUMMARY.md→docs/
   - ✅ EXECUTIVE-SUMMARY.md→docs/

2. **测试脚本合理位置**
   - ✅ tests/→moved to connect/tests/
   - ✅ 原因：tests是connect模块的一部分，应该在connect下

3. **清理冗余文件**
   - ✅ 删除 STAGE1-PACKAGE.sh
   - ✅ 删除 STAGE1-SUBMIT.sh
   - ✅ 删除 macda-stage1-*.tar.gz（临时打包文件）

4. **创建临时文件目录**
   - ✅ 创建 temp/ 目录
   - 🔹 用途：存放build artifacts、临时脚本等

5. **根目录清晰化**
   - ✅ 仅保留必要文件（README.md, AGENTS.md）
   - ✅ 删除散乱的脚本和文档
   - ✅ 项目结构清晰易维护

## 📖 快速导航

### 我想...

| 需求 | 位置 | 说明 |
|------|------|------|
| 快速了解项目 | [README.md](./README.md) | ← **从这里开始** |
| Phase 1文档导航 | [stage1/README.md](./stage1/README.md) | 阶段1文档索引 |
| 完成度报告 | [CLEANUP-SUMMARY.md](./CLEANUP-SUMMARY.md) | 整理清单 |
| 技术总结 | [stage1/STAGE1-FINAL-SUMMARY.md](./stage1/STAGE1-FINAL-SUMMARY.md) | 技术细节 |
| NB67规范 | [requirements/binary-spec.md](./requirements/binary-spec.md) | 二进制协议定义 |
| 快速命令 | [stage1/STAGE1-QUICK-REF.md](./stage1/STAGE1-QUICK-REF.md) | Docker命令速查 |
| 验收清单 | [stage1/STAGE1-CHECKLIST.md](./stage1/STAGE1-CHECKLIST.md) | 测试验收 |

### 我想测试...

```bash
# 所有测试脚本都在 connect/tests/

# 1. 验证Kafka连接
bash connect/tests/test-kafka-connection.sh

# 2. 验证NB67解析
bash connect/tests/test-nb67-parsing.sh

# 3. 端到端完整测试
bash connect/tests/test-end-to-end.sh
```

## 🗂️ 各目录职责

### baseEnv/
- Redpanda（Kafka兼容）的Docker Compose配置
- 用于本地开发环境搭建
- 包含connect插件示例

### connect/
- **核心应用代码**
- `cmd/connect-nb67/` - Go应用主程序
- `codec/` - NB67二进制解析器（Kaitai生成的Go代码）
- `config/` - Kafka Connect连接器配置
- `tests/` - 自动化测试脚本

### docs/
- **所有文档统一管理**
- `stage1/` - Phase 1阶段文档（已完成功能说明）
- `requirements/` - 项目需求和规格说明
- 其他 `xx-*.md` - 历史决策文档和技术分析
- CLEANUP-SUMMARY.md - 整理总结
- EXECUTIVE-SUMMARY.md - 执行总结

### oldproj/
- 已废弃项目参考
- 用于参考遗留代码架构
- 包含Python版本和前端参考

### temp/
- 临时文件存放区
- gitkeep确保目录存在
- 用于：
  - build artifacts
  - logs
  - 临时脚本
  - staging区域

## 🚀 入门步骤

### Step 1: 理解项目
```bash
cat README.md              # 项目总体说明
cat docs/README.md         # 文档导航
```

### Step 2: 查看Phase 1成果
```bash
cat docs/stage1/README.md  # Phase 1文档索引
cat docs/CLEANUP-SUMMARY.md # 整理清单
```

### Step 3: 运行测试
```bash
bash connect/tests/test-kafka-connection.sh
bash connect/tests/test-nb67-parsing.sh
bash connect/tests/test-end-to-end.sh
```

### Step 4: 查看技术细节
```bash
# 如需深入了解
cat docs/stage1/STAGE1-FINAL-SUMMARY.md
cat docs/requirements/binary-spec.md
```

## ✨ 整理的益处

1. **清晰的结构** - 用户一眼知道什么在哪
2. **易于维护** - 没有散乱的文件和脚本
3. **专业级组织** - 符合标准的Python/Go项目布局
4. **版本控制友好** - 不会有冗余的临时文件污染git
5. **模块独立** - connect/tests是独立的，不与其他项目混淆

## 📝 项目文件清单

### 根目录文件（必要）
```
README.md       - 项目总说明
AGENTS.md       - Agent操作指南
```

### 文档目录（docs/）- 17份文档
```
stage1/                    - Phase 1完成文档（6份）
requirements/
  binary-spec.md           - NB67二进制规格
01-10-*.md                 - 技术分析和可行性报告
CLEANUP-SUMMARY.md         - 整理清单
EXECUTIVE-SUMMARY.md       - 执行总结
README.md                  - 文档索引
...
```

### 测试脚本（connect/tests/）- 3个
```
test-kafka-connection.sh   - Kafka连接验证
test-nb67-parsing.sh       - NB67解析验证
test-end-to-end.sh         - 端到端测试
```

### 源代码（connect/）
```
cmd/connect-nb67/          - Go main程序 (4个文件)
codec/                     - NB67解析器和规格
config/                    - Connect配置 (2个文件)
```

## 🔄 后续工作结构

后续Phase 2/3的工作也应该遵循这样的结构：

```
connect/                   ← 应用代码所有单一目录
  ├── cmd/
  ├── config/
  ├── tests/               ← 测试脚本
  └── ...

docs/
  ├── phase2/              ← 新阶段文档
  └── ...

temp/                      ← 临时文件
```

**原则**: 
- 代码都在 `connect/` 下
- 文档都在 `docs/` 下
- 临时文件都在 `temp/` 下
- 根目录保持清晰（只有规范必需文件）

---

**最后更新**: 2026-02-19  
**整理完成度**: ✅ 100%  
**下一步**: [执行测试脚本](../connect/tests/test-kafka-connection.sh)
