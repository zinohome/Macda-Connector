<template>
    <div class="trend-card">
        <div class="section-header">
            <div class="title-group">
                <span class="title-icon"></span>
                <span class="title-text">参数趋势分析</span>
            </div>
            <div class="controls">
                <!-- 参数多选 -->
                <el-select
                    v-model="selectedParams"
                    multiple
                    collapse-tags
                    collapse-tags-tooltip
                    placeholder="选择参数"
                    size="small"
                    style="width:220px;margin-right:10px"
                    @change="fetchData"
                >
                    <el-option v-for="p in paramOptions" :key="p.key" :label="p.label" :value="p.key" />
                </el-select>
                <!-- 时间粒度 -->
                <el-button-group>
                    <el-button
                        v-for="tab in tabs"
                        :key="tab.value"
                        :type="currentTab === tab.value ? 'primary' : 'default'"
                        size="small"
                        @click="handleTabChange(tab.value)"
                    >{{ tab.label }}</el-button>
                </el-button-group>
            </div>
        </div>
        <div class="chart-content">
            <div ref="chartRef" class="main-echart"></div>
            <div v-if="!hasData" class="no-data-overlay">
                <img src="/img/no-data.svg" width="60" />
                <p>该时间段暂无历史趋势数据</p>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { getThDataByDvcAddrApi, getTrendParamsApi } from '@/api/api.js'

const props = defineProps({
    carriageId: { type: String, default: '' }
})

const chartRef = ref(null)
let myChart = null
const currentTab = ref('hour')
const handleResize = () => myChart?.resize()
const hasData = ref(true)
let refreshTimer = null

const tabs = [
    { label: '近1小时', value: 'hour' },
    { label: '近1天',   value: 'day' },
    { label: '近1周',   value: 'week' },
    { label: '近1月',   value: 'month' },
]

// 默认展示温度三条线
const selectedParams = ref(['ras_u1', 'fas_u1', 'tic'])
const paramOptions = ref([])

// 加载可选参数列表
const loadParamOptions = async () => {
    try {
        const res = await getTrendParamsApi()
        if (res?.code === 200 && res.data) {
            paramOptions.value = res.data
        }
    } catch (e) {
        // BFF 未响应时使用默认三参数
        paramOptions.value = [
            { key: 'ras_u1', label: '回风温度(机组1)', unit: '℃' },
            { key: 'fas_u1', label: '新风温度(机组1)', unit: '℃' },
            { key: 'tic',    label: '客室温度',        unit: '℃' },
        ]
    }
}

const handleTabChange = (val) => {
    currentTab.value = val
    fetchData()
}

const formatBucket = (timeStr) => {
    if (!timeStr) return ''
    const date = new Date(timeStr)
    const h = String(date.getHours()).padStart(2, '0')
    const m = String(date.getMinutes()).padStart(2, '0')
    const s = String(date.getSeconds()).padStart(2, '0')
    if (currentTab.value === 'hour') return `${h}:${m}:${s}`
    if (currentTab.value === 'day')  return `${h}:${m}`
    return `${date.getMonth() + 1}-${date.getDate()} ${h}:${m}`
}

const COLORS = [
    '#2186cf','#42ad5d','#ffa55c','#e65355',
    '#9b59b6','#1abc9c','#f39c12','#3498db',
    '#e74c3c','#27ae60','#8e44ad','#d35400',
]

const buildChartOptions = (timeAxis, seriesList) => ({
    backgroundColor: 'transparent',
    tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(10,15,29,0.9)',
        borderColor: '#2186cf',
        textStyle: { color: '#fff' },
        formatter: (params) => {
            let res = `<div style="font-weight:bold;margin-bottom:5px">${params[0]?.name}</div>`
            params.forEach(item => {
                const val = item.value !== null && item.value !== undefined
                    ? Number(item.value).toFixed(1)
                    : '-'
                res += `<div style="display:flex;justify-content:space-between;gap:20px">
                    <span>${item.marker} ${item.seriesName}</span>
                    <span style="font-weight:bold">${val}</span>
                </div>`
            })
            return res
        }
    },
    legend: {
        data: seriesList.map(s => s.name),
        top: 10,
        textStyle: { color: '#676e82', fontSize: 11 }
    },
    grid: { left: '3%', right: '4%', bottom: '10%', containLabel: true },
    dataZoom: [
        { type: 'inside', xAxisIndex: 0 },
        { type: 'slider', xAxisIndex: 0, height: 20, bottom: 5,
          textStyle: { color: '#676e82' }, borderColor: '#262e45',
          fillerColor: 'rgba(33,134,207,0.15)', handleStyle: { color: '#2186cf' } }
    ],
    xAxis: {
        type: 'category',
        boundaryGap: false,
        data: timeAxis,
        axisLine: { lineStyle: { color: '#262e45' } },
        axisLabel: { color: '#676e82', fontSize: 10 }
    },
    yAxis: {
        type: 'value',
        axisLabel: { color: '#676e82', fontSize: 10 },
        splitLine: { lineStyle: { color: '#1a2234', type: 'dashed' } }
    },
    series: seriesList,
})

const fetchData = async () => {
    if (!props.carriageId) return
    const params = selectedParams.value.length > 0 ? selectedParams.value : ['ras_u1', 'fas_u1', 'tic']
    try {
        const res = await getThDataByDvcAddrApi({
            carriageNo: props.carriageId,
            type: currentTab.value,
            params,
        })
        const list = res[currentTab.value] || res.data || []
        if (list.length === 0) { hasData.value = false; return }
        hasData.value = true

        const timeAxis = list.map(i => formatBucket(i.bucket || i.time))
        const seriesList = params.map((key, idx) => {
            const paramDef = paramOptions.value.find(p => p.key === key)
            return {
                name: paramDef?.label || key,
                type: 'line',
                smooth: true,
                showSymbol: false,
                data: list.map(i => {
                    const v = i[key]
                    return v !== null && v !== undefined ? Number(v) : null
                }),
                lineStyle: { color: COLORS[idx % COLORS.length], width: 2 },
                connectNulls: false,
            }
        })

        nextTick(() => {
            if (!myChart && chartRef.value) myChart = echarts.init(chartRef.value)
            if (myChart) myChart.setOption(buildChartOptions(timeAxis, seriesList), true)
        })
    } catch (error) {
        console.error('Trend data error:', error)
        hasData.value = false
    }
}

watch(() => props.carriageId, () => fetchData())

import { MONITOR_CONFIG } from '@/config/monitorConfig.js'

onMounted(async () => {
    await loadParamOptions()
    fetchData()
    refreshTimer = setInterval(fetchData, MONITOR_CONFIG.refreshInterval)
    window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
    if (refreshTimer) clearInterval(refreshTimer)
    window.removeEventListener('resize', handleResize)
    if (myChart) { myChart.dispose(); myChart = null }
})
</script>

<style scoped lang="scss">
.trend-card { padding: 20px; }

.section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    flex-wrap: wrap;
    gap: 10px;

    .title-group {
        display: flex;
        align-items: center;
        gap: 8px;
    }
    .title-icon {
        width: 4px; height: 16px;
        background: #ffffff; border-radius: 2px;
    }
    .title-text {
        color: #ffffff; font-size: 16px; font-weight: bold;
    }
    .controls {
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 8px;
    }
}

.chart-content {
    height: 400px;
    position: relative;
    background: #101626;
    border-radius: 8px;
    border: 1px solid #1a2234;
}
.main-echart { width: 100%; height: 100%; }

.no-data-overlay {
    position: absolute; top: 0; left: 0; width: 100%; height: 100%;
    display: flex; flex-direction: column;
    justify-content: center; align-items: center;
    background: rgba(10,15,29,0.8);
    color: #676e82; z-index: 5;
    p { margin-top: 10px; font-size: 13px; }
}

:deep(.el-button-group) {
    .el-button {
        background: #1a2234; border-color: #394153; color: #676e82;
        &--primary { background: #2186cf; border-color: #2186cf; color: #fff; }
    }
}

:deep(.el-select) {
    --el-fill-color-blank: #1a2234;
    --el-border-color: #394153;
    --el-text-color-regular: #d1d9e7;
}
</style>
