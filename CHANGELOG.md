# Changelog

All notable changes to this project will be documented in this file.


## v2.5.0 (2026-05-13)

### Features
- feat(release): add release.sh with prepare/publish commands
- feat: 预警测试模式+Reset按钮+PHM策略更新+出厂默认值
- feat(event-builder): hot-reload predict thresholds from hvac.warning_config
- feat: add one-click deploy.sh script
- feat(frontend): F13-F17 new pages and navigation
- feat(frontend): F8-F12 medium changes
- feat(frontend): F1-F7 small UI changes
- feat(bff): add B1-B8 API endpoints
- feat(db): add recovery_time to fact_event, create warning_config table
- feat(ui): hide orange/blue flow path lines in AC schematic SVG

### Bug Fixes
- fix(release): guard duplicate CHANGELOG entry and re-tag
- fix(deploy): mkdir 改用 sudo，支持 /data/MACDA2 被彻底清空后重建
- fix(deploy): 修复 01-init.sql 幂等性 + deploy.sh 错误可见性
- fix(pipeline): 兼容坏 ingest_time（.split('%').index(0) 截断 %!(EXTRA) 后缀）
- fix(timezone): now().format → now().ts_format，修复 ingest_time 格式错误
- fix(HVAC_09): 超温前置条件改用制冷系统核心故障，排除传感器故障位
- fix(realtime-warning): strategy 从 DB 读取替代本地硬编码数组
- fix(migration-05): convert literal 
 to real newlines in strategy fields
- fix(event-builder): HVAC_09 回风温度+PHM模式修正 & 全链路时区统一东八区
- fix(deploy): add 03-migration-20260512.sql to Step 3c and Step 6
- fix: SVG no-cache headers + remove .bak files from image
- fix: dark theme for historyWarning detail dialog and warningConfig inputs
- fix: warningConfig dialog complete dark theme overhaul
- fix: dark theme for historyData pagination and warningConfig dialog
- fix: move 预警条件设置 to historyWarning and preserve params on goBack
- fix: restore 预警条件设置 button in trainInfo + fix predict detail timezone
- fix: add global dark theme styles to historyWarning to fix white bg on hard reload
- fix(bff): apply unit filter in JS layer for historyWarning
- fix(bff): unit filter in historyWarning supports HVAC numeric fault codes
- fix: initialize carriageNo from URL trainCoach param in historyAlarm/Warning
- fix: add 全部机组 option as default in unit single-select
- fix: history alarm/warning use single select for carriage and unit
- fix: multi-select tag color in trend analysis matches dark theme
- fix: nav buttons cleanup and trainNo/trainCoach param passing
- fix: fix 5 issues from user feedback
- fix(frontend): remove offline banner, show - naturally when no data
- fix(frontend): severity level font color should be white (#ffffff)
- fix(frontend): offline banner shows last known data instead of clearing display
- fix(bff): fix ROUND() type error - cast to numeric not float
- fix(frontend+bff): fix 2 post-verify issues
- fix(frontend): fix 6 bugs found in code review + browser test
- fix(event-builder): align PHM predict thresholds with doc v02

### Chores
- chore: 从 .gitignore 移除 .release-version（release.sh 需要追踪此文件）
- chore: 从 git 追踪中移除 connect-nb67 二进制（.gitignore 已配置）
- chore: add release script prerequisites
- chore: 将 connect-nb67 二进制加入 .gitignore
- chore: 整合清理，准备一键部署基线 v2.3.5/v2.4.19
- chore: bump nb-parse-connect to v2.3.2 in Dev compose
- chore: bump nb67-web and nb67-bff to v2.4.16 in Dev and Prod compose
- chore: sync image versions to v2.4.11/v2.4.3 in baseEnv compose files
- chore: bump web/bff to v2.4.0 in baseEnv compose files
- chore: sync image versions to v2.3.x in baseEnv compose files
- chore: add docs appendix 2, update package locks, ignore temp files

