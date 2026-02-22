<template>
    <div class="trend-card">
        <div class="section-header">
            <div class="title-group">
                <span class="title-icon"></span>
                <span class="title-text">温度趋势分析</span>
            </div>
            <div class="granularity-tabs">
                <el-button-group>
                    <el-button 
                        v-for="tab in tabs" 
                        :key="tab.value"
                        :type="currentTab === tab.value ? 'primary' : 'default'"
                        size="small"
                        @click="handleTabChange(tab.value)"
                    >
                        {{ tab.label }}
                    </el-button>
                </el-button-group>
            </div>
        </div>
        <div class="chart-content">
            <div ref="chartRef" class="main-echart"></div>
            <div v-if="!hasData" class="no-data-overlay">
                <img src="/src/assets/img/no-data.svg" width="60" />
                <p>该时间段暂无历史趋势数据</p>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import { getThDataByDvcAddrApi } from '@/api/api.js'

const props = defineProps({
    carriageId: { type: String, default: '' }
})

const chartRef = ref(null)
let myChart = null
const currentTab = ref('hour')
const hasData = ref(true)

const tabs = [
    { label: '时 (秒级)', value: 'hour' },
    { label: '天 (分级)', value: 'day' },
    { label: '周 (15分)', value: 'week' },
    { label: '月 (30分)', value: 'month' }
]

const handleTabChange = (val) => {
    currentTab.value = val
    fetchData()
}

const formatTime = (timeStr) => {
    const date = new Date(timeStr)
    const h = String(date.getHours()).padStart(2, '0')
    const m = String(date.getMinutes()).padStart(2, '0')
    const s = String(date.getSeconds()).padStart(2, '0')
    
    if (currentTab.value === 'hour') return `${h}:${m}:${s}`
    if (currentTab.value === 'day') return `${h}:${m}`
    return `${date.getMonth() + 1}-${date.getDate()} ${h}:${m}`
}

const getChartOptions = (data) => {
    return {
        backgroundColor: 'transparent',
        tooltip: {
            trigger: 'axis',
            backgroundColor: 'rgba(10, 15, 29, 0.9)',
            borderColor: '#2186cf',
            textStyle: { color: '#fff' },
            formatter: (params) => {
                let res = `<div style="font-weight:bold;margin-bottom:5px">${params[0].name}</div>`
                params.forEach(item => {
                    res += `<div style="display:flex;justify-content:space-between;gap:20px">
                        <span>${item.marker} ${item.seriesName}</span>
                        <span style="font-weight:bold">${item.value} ℃</span>
                    </div>`
                })
                return res
            }
        },
        legend: {
            data: ['客室温度', '新风温度', '目标设定温度'],
            top: 10,
            textStyle: { color: '#676e82' }
        },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: data.time,
            axisLine: { lineStyle: { color: '#262e45' } },
            axisLabel: { color: '#676e82', fontSize: 10 }
        },
        yAxis: {
            type: 'value',
            axisLabel: { color: '#676e82', formatter: '{value} ℃' },
            splitLine: { lineStyle: { color: '#1a2234', type: 'dashed' } }
        },
        series: [
            {
                name: '客室温度',
                type: 'line',
                smooth: true,
                showSymbol: false,
                data: data.cabin,
                lineStyle: { color: '#2186cf', width: 2 },
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        { offset: 0, color: 'rgba(33, 134, 207, 0.3)' },
                        { offset: 1, color: 'rgba(33, 134, 207, 0)' }
                    ])
                }
            },
            {
                name: '新风温度',
                type: 'line',
                smooth: true,
                showSymbol: false,
                data: data.fresh,
                lineStyle: { color: '#42ad5d', width: 2 },
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        { offset: 0, color: 'rgba(66, 173, 93, 0.2)' },
                        { offset: 1, color: 'rgba(66, 173, 93, 0)' }
                    ])
                }
            },
            {
                name: '目标设定温度',
                type: 'line',
                smooth: true,
                showSymbol: false,
                data: data.target,
                lineStyle: { color: '#ffa55c', width: 2, type: 'dashed' }
            }
        ]
    }
}

const fetchData = async () => {
    if (!props.carriageId) return
    
    // 根据状态构建不同的聚合参数
    let params = {
        carriageNo: props.carriageId,
        type: currentTab.value // 假设后端支持此参数
    }
    
    try {
        const res = await getThDataByDvcAddrApi(params)
        // 兼容处理不同维度的 Key
        const list = res[currentTab.value] || res.data || []
        
        if (list.length === 0) {
            hasData.value = false
            return
        }
        hasData.value = true

        const formatted = {
            time: list.map(i => formatTime(i.bucket)),
            cabin: list.map(i => i.ras_sys),
            fresh: list.map(i => i.fas_sys),
            target: list.map(i => i.tic)
        }

        nextTick(() => {
            if (!myChart) myChart = echarts.init(chartRef.value)
            myChart.setOption(getChartOptions(formatted))
        })
    } catch (error) {
        console.error('Trend data error:', error)
        hasData.value = false
    }
}

watch(() => props.carriageId, fetchData)

onMounted(() => {
    fetchData()
    window.addEventListener('resize', () => myChart?.resize())
})
</script>

<style scoped lang="scss">
.trend-card {
    padding: 20px;
}

.section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    .title-group {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .title-icon {
        width: 4px;
        height: 16px;
        background: #ffffff;
        border-radius: 2px;
    }

    .title-text {
        color: #ffffff;
        font-size: 16px;
        font-weight: bold;
    }
}

.chart-content {
    height: 400px;
    position: relative;
    background: #101626;
    border-radius: 8px;
    border: 1px solid #1a2234;
}

.main-echart {
    width: 100%;
    height: 100%;
}

.no-data-overlay {
    position: absolute;
    top: 0; left: 0; width: 100%; height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    background: rgba(10, 15, 29, 0.8);
    color: #676e82;
    z-index: 5;
    p { margin-top: 10px; font-size: 13px; }
}

:deep(.el-button-group) {
    .el-button {
        background: #1a2234;
        border-color: #394153;
        color: #676e82;
        &--primary {
            background: #2186cf;
            border-color: #2186cf;
            color: #fff;
        }
    }
}
</style>
