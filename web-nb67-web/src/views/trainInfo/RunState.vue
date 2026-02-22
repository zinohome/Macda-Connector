<template>
    <div class="run-state-card">
        <div class="section-header">
            <span class="title-icon"></span>
            <span class="title-text">运行状态信息</span>
        </div>
        <div class="table-container">
            <div class="custom-table">
                <!-- 表头 -->
                <div class="table-header flex-row">
                    <div class="cell label-cell"></div>
                    <div v-for="n in 6" :key="n" class="cell car-cell">
                        <div class="unit-titles">
                            <span v-if="n <= 3">1号机组</span>
                            <span v-else>2号机组</span>
                            
                            <span v-if="n <= 3">2号机组</span>
                            <span v-else>1号机组</span>
                        </div>
                    </div>
                </div>

                <!-- 数据行 -->
                <div v-for="row in rows" :key="row.key" class="table-row flex-row">
                    <div class="cell label-cell">{{ row.label }}</div>
                    <div v-for="n in 6" :key="n" class="cell car-group">
                        <div class="unit-data-pair">
                            <div :class="['unit-val', getValClass(n, 1, row)]">
                                {{ getVal(n, 1, row) }}
                            </div>
                            <div :class="['unit-val', getValClass(n, 2, row)]">
                                {{ getVal(n, 2, row) }}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, watch, onUnmounted } from 'vue'
import { getRunningState } from '@/api/api'

const props = defineProps({
    trainId: { type: String, default: '' }
})

const rawData = ref([])
let refreshTimer = null

const rows = [
    { label: '压缩机1', key: 'comp_u1', isStatus: true },
    { label: '压缩机2', key: 'comp_u2', isStatus: true },
    { label: '通风机', key: 'ef', isStatus: true },
    { label: '冷凝风机', key: 'cf', isStatus: true },
    { label: '新风阀', key: 'fadpos', suffix: '%' },
    { label: '回风阀', key: 'radpos', suffix: '%' }
]

// 对称逻辑：1-3车(1,2), 4-6车(2,1)
const getUnitIdx = (carIdx, pos) => {
    if (carIdx <= 3) return pos // 1, 2
    return pos === 1 ? 2 : 1    // 2, 1
}

const getVal = (carIdx, pos, row) => {
    // 匹配车厢：严格按 ID 匹配，不再使用下标兜底以防数据错位
    const carData = rawData.value.find(item => {
        const cNo = String(item.carriage_id || item.carriage_no || '')
        return cNo == String(carIdx) || 
               cNo == ("0" + carIdx) || 
               cNo.endsWith("0" + carIdx)
    })
    
    if (!carData) return '-'
    const unitIdx = getUnitIdx(carIdx, pos)
    
    // 智能查找 Key：支持多种命名规范 (PascalCase, snake_case, cfbk前缀等)
    const findKeyInObj = (base) => {
        const toPascal = (s) => s.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1).toLowerCase()).join('');
        const pBase = toPascal(base);
        const possible = [
            base,
            base.toLowerCase(),
            pBase,
            `Cfbk${pBase}`,
            `cfbk_${base}`.toLowerCase(),
            `cfbk_${base}`.toUpperCase()
        ];
        for (const k of possible) {
            if (carData[k] !== undefined) return k;
        }
        // 模糊匹配：忽略下划线并转小写进行包含测试
        const flatBase = base.toLowerCase().replace(/_/g, '');
        return Object.keys(carData).find(k => k.toLowerCase().replace(/_/g, '').includes(flatBase));
    };

    if (row.isStatus) {
        // 状态类：构造基础 Key
        let baseKey = row.key.startsWith('comp_u') 
            ? `comp_u${unitIdx}${row.key.slice(-1)}` 
            : `${row.key}_u${unitIdx}`;
        
        const key = findKeyInObj(baseKey);
        const val = carData[key];
        
        if (val === undefined || val === null) return '-'
        // 允许数字 1, 布尔值 true, 字符串 "true" (忽略大小写)
        const isRun = val == 1 || val === true || String(val).toLowerCase() === 'true';
        return isRun ? '运行' : '停机'
    } else {
        // 数值类：如新风阀、回风阀
        const key = findKeyInObj(`${row.key}_u${unitIdx}`);
        const num = carData[key];
        return num !== undefined && num !== null ? `${num}${row.suffix}` : '-'
    }
}

const getValClass = (carIdx, pos, row) => {
    if (!row.isStatus) return ''
    const val = getVal(carIdx, pos, row)
    if (val === '运行') return 'status-run'
    if (val === '停机') return 'status-stop'
    return ''
}

const fetchData = () => {
    // rawData.value = [] // 自动刷新建议不直接置空，避免闪烁
    if (!props.trainId) return
    getRunningState(props.trainId).then(res => {
        rawData.value = res.vw_running_state_info || []
        console.log('[RunState] Data Updated:', rawData.value.length)
    })
}

watch(() => props.trainId, () => {
    rawData.value = [] // ID 变化时可以置空
    fetchData()
})

import { MONITOR_CONFIG } from '@/config/monitorConfig.js'

onMounted(() => {
    fetchData()
    refreshTimer = setInterval(fetchData, MONITOR_CONFIG.refreshInterval)
})

onUnmounted(() => {
    if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<style scoped lang="scss">
.run-state-card {
    padding: 0;
}

.section-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
    padding-left: 10px;

    .title-icon {
        width: 4px;
        height: 16px;
        background: #ffffff;
        border-radius: 2px;
    }

    .title-text {
        color: #ffffff;
        font-size: 15px;
        font-weight: bold;
    }
}

.custom-table {
    display: flex;
    flex-direction: column;
    border: 1px solid #262e45;
    background: #1a2234;

    .flex-row {
        display: grid;
        grid-template-columns: 100px repeat(6, 1fr);
        border-bottom: 1px solid #262e45;
        &:last-child { border-bottom: none; }
    }

    .cell {
        padding: 2px 6px;
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        border-right: 1px solid #262e45;
        &:last-child { border-right: none; }
        min-height: 26px;
    }

    .label-cell {
        background: #141a2e !important; // 与表头背景保持一致
        color: #2186cf;
        font-size: 13px;
        font-weight: bold;
        align-items: center;
    }

    .table-header {
        background: #141a2e;
        .unit-titles {
            width: 100%;
            display: flex;
            justify-content: space-between;
            padding: 0 4px;
            font-size: 11px;
            color: #ffffff; // 机组名称改为白色
            
            span {
                flex: 1;
                text-align: center;
                white-space: nowrap;
                font-weight: 500;
            }
        }
    }

    .unit-data-pair {
        width: 100%;
        display: flex;
        gap: 3px;
        .unit-val {
            flex: 1;
            height: 20px;
            line-height: 20px;
            text-align: center;
            font-size: 11px;
            border-radius: 2px;
            color: rgba(255,255,255,0.8);
            background: rgba(255,255,255,0.05) !important; // 统一所有格子的默认背景

            &.status-run { background: #42ad5d !important; color: #fff; }
            &.status-stop { background: #e65355 !important; color: #fff; }
        }
    }
}
</style>
