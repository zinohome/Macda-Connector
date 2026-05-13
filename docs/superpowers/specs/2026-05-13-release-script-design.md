# Release Script Design

**Date:** 2026-05-13  
**Status:** Approved  
**Scope:** 统一版本发布脚本，收尾当前开发阶段成果

---

## 背景与目标

当前项目版本号散落在 5 个文件共 8 处，`dist/` 部署包与实际镜像版本严重脱节（dist 写的 v2.1.2，实际镜像已到 v2.4.19）。每次发布需手动同步，容易遗漏。

**目标：** 一个脚本（`release.sh`），两个子命令，把"版本号同步 + 镜像构建推送 + git tag"串成一个可重复执行的流程。

---

## 架构设计

### 脚本入口

```
release.sh prepare <version>   # Phase 1：更新文件 + 生成 CHANGELOG + commit
release.sh publish             # Phase 2：构建推送镜像 + 打 tag + push
```

两个阶段之间有人工 review 节点，publish 前可检查 prepare 的变更内容。

---

## Phase 1：prepare

### 触发方式
```bash
./release.sh prepare v2.5.0
```

### 执行步骤

1. **前置校验**
   - 验证版本号格式（`vX.Y.Z` semver）
   - 检查 git 工作区干净（无未提交变更）
   - 检查当前分支为 `main`

2. **更新版本号**（`sed` 精确替换，不做全局替换）

   | 文件 | 替换目标 |
   |------|---------|
   | `baseEnv/docker-compose-Dev.yml` | `nb67-web:<旧版>` → `nb67-web:<新版>` |
   | `baseEnv/docker-compose-Dev.yml` | `nb67-bff:<旧版>` → `nb67-bff:<新版>` |
   | `baseEnv/docker-compose-Dev.yml` | `nb-parse-connect:<旧版>` → `nb-parse-connect:<新版>` |
   | `baseEnv/docker-compose-Prod.yml` | `nb67-web:<旧版>` → `nb67-web:<新版>` |
   | `baseEnv/docker-compose-Prod.yml` | `nb67-bff:<旧版>` → `nb67-bff:<新版>` |
   | `dist/docker-compose-Web.yml` | `nb67-web:<旧版>` → `nb67-web:<新版>` |
   | `dist/docker-compose-Web.yml` | `nb67-bff:<旧版>` → `nb67-bff:<新版>` |
   | `dist/docker-compose-Data.yml` | `nb-parse-connect:<旧版>` → `nb-parse-connect:<新版>` |
   | `dist/README.md` | 版本头（`**版本**：vX.Y.Z`） |

   旧版本号从当前文件中读取（`grep` 提取），不硬编码。

3. **生成 CHANGELOG**

   从上一个 git tag 到 HEAD，按 conventional commit 前缀分组：
   ```
   ## v2.5.0 (2026-05-13)

   ### Features
   - feat: ...

   ### Bug Fixes
   - fix: ...

   ### Chores
   - chore: ...
   ```
   追加写入 `CHANGELOG.md`（新版本内容插入文件头部）。

4. **保存版本号**
   将目标版本写入 `.release-version`（`publish` 阶段读取）。

5. **自动 commit**
   ```
   chore(release): prepare v2.5.0
   ```
   包含所有修改文件 + `CHANGELOG.md` + `.release-version`。

---

## Phase 2：publish

### 触发方式
```bash
./release.sh publish
```

### 执行步骤

1. **读取版本号**
   从 `.release-version` 读取（由 prepare 写入）。

2. **构建并推送 web + bff**
   复用 `build-and-push.sh` 中的镜像名和构建逻辑：
   - `harbor.naivehero.top:8443/macda2/nb67-web:<版本>`
   - `harbor.naivehero.top:8443/macda2/nb67-bff:<版本>`
   - 平台：`linux/amd64`

3. **构建并推送 connect**
   ```bash
   docker build --platform linux/amd64 \
     -f connect/Dockerfile.connect \
     -t harbor.naivehero.top:8443/macda2/nb-parse-connect:<版本> \
     connect
   docker push harbor.naivehero.top:8443/macda2/nb-parse-connect:<版本>
   ```

4. **打 git tag 并推送**
   ```bash
   git tag <版本>
   git push origin main
   git push origin <版本>
   ```

5. **打印发布摘要**（镜像地址、tag、耗时）

---

## 文件变更清单

| 文件 | 变更类型 |
|------|---------|
| `release.sh`（新建） | 主脚本 |
| `CHANGELOG.md`（新建） | 版本历史 |
| `.release-version`（临时，加入 .gitignore） | prepare → publish 传值 |
| `baseEnv/docker-compose-Dev.yml` | 版本号更新 |
| `baseEnv/docker-compose-Prod.yml` | 版本号更新 |
| `dist/docker-compose-Web.yml` | 版本号更新 |
| `dist/docker-compose-Data.yml` | 版本号更新 |
| `dist/README.md` | 版本号更新 |

---

## 约束

- 脚本不修改 `connect/Dockerfile.connect` 或任何源码
- 不替换 Redpanda、TimescaleDB 等基础设施镜像版本（这些独立维护）
- `publish` 必须在 `prepare` 之后执行（检查 `.release-version` 存在）
- 发布完成后自动删除 `.release-version`
