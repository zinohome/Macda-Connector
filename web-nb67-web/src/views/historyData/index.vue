<template>
    <div class="warp2">
        <!-- 顶部筛选栏 -->
        <div class="header-monitor">
            <div class="header-left">
                <el-button type="primary" class="nav-btn" icon="ArrowLeft" @click="goBack">返回</el-button>
            </div>

            <div class="header-right">
                <el-form :inline="true" :model="filterForm" class="filter-form">
                    <el-form-item label="车号">
                        <el-select v-model="filterForm.trainNo" placeholder="请选择车号" class="train-select">
                            <el-option v-for="item in trainList" :key="item.train_id" :label="item.train_id" :value="item.train_id" />
                        </el-select>
                    </el-form-item>
                    <el-form-item label="车厢号">
                        <el-select v-model="filterForm.carriageNo" placeholder="请选择车厢号" class="carriage-select">
                            <el-option v-for="item in carriageList" :key="item.carriage_no" :label="item.label" :value="item.carriage_no" />
                        </el-select>
                    </el-form-item>
                    <el-form-item label="时间范围">
                        <el-date-picker
                            v-model="filterForm.timeRange"
                            type="datetimerange"
                            range-separator="至"
                            start-placeholder="开始时间"
                            end-placeholder="结束时间"
                            format="YYYY-MM-DD HH:mm"
                            value-format="YYYY-MM-DD HH:mm:ss"
                        />
                    </el-form-item>
                    <el-form-item>
                        <el-button type="primary" class="nav-btn" icon="Search" @click="handleQuery">查询</el-button>
                        <el-button type="success" class="export-btn" icon="Download" @click="handleExportAll">导出数据</el-button>
                    </el-form-item>
                </el-form>
            </div>
        </div>

        <div class="monitor-container">
            <div class="section-box">
                <div class="section-header">
                    <span class="title-icon"></span>
                    <span class="title-text">历史数据 (日聚合)</span>
                </div>
                <el-table :data="summaryData" border stripe style="width: 100%" v-loading="loading">
                    <el-table-column label="车号" prop="trainNo" align="center" width="120" />
                    <el-table-column label="车厢号" prop="carriageNo" align="center" width="120">
                        <template #default="scope">{{ scope.row.carriageNo }}车厢</template>
                    </el-table-column>
                    <el-table-column label="日期" prop="date" align="center" />
                    <el-table-column label="操作" align="center" width="250">
                        <template #default="scope">
                            <el-button link type="primary" icon="Download" @click="handleDownloadRow(scope.row)">下载</el-button>
                            <el-button link type="primary" icon="View" @click="handleViewDetail(scope.row)">查看</el-button>
                        </template>
                    </el-table-column>
                </el-table>
            </div>
        </div>

        <!-- 详情对话框 (非模态) -->
        <el-dialog
            v-model="detailVisible"
            title="历史数据详情"
            width="90%"
            top="5vh"
            :modal="false"
            draggable
            custom-class="history-detail-dialog"
        >
            <div class="detail-filter">
                <span class="detail-info">当前选择：{{ selectedDate }} | {{ filterForm.trainNo }}号车 | {{ filterForm.carriageNo }}车厢</span>
            </div>
            
            <el-table :data="detailList" border stripe height="60vh" v-loading="detailLoading">
                <el-table-column type="index" label="序号" width="60" fixed align="center" />
                <el-table-column prop="train_id" label="列车号" width="100" />
                <el-table-column prop="carriage_no" label="车厢号" width="100">
                    <template #default="scope">{{ scope.row.carriage_no }}车厢</template>
                </el-table-column>
                <el-table-column label="机组" width="100" align="center">
                    <template #default="scope">
                        {{  scope.row.unit_no || formatUnit(scope.row.device_id) }}
                    </template>
                </el-table-column>
                <el-table-column prop="i_seat_temp" label="室内温度" width="100" />
                <el-table-column prop="i_outer_temp" label="新风温度" width="100" />
                <el-table-column prop="i_set_temp" label="设定温度" width="100" />
                <el-table-column prop="i_hum" label="湿度" width="80" />
                <el-table-column prop="i_co2" label="CO2浓度" width="100" />
                <el-table-column prop="work_status" label="工作状态" width="100" />
                <el-table-column prop="dwef_op_tm" label="通风机时间" width="120" />
                <el-table-column prop="w_crnt_cf1" label="冷凝流1" width="100" />
                <el-table-column prop="w_crnt_cf2" label="冷凝流2" width="100" />
                <el-table-column prop="w_crnt_cp1" label="压机流1" width="100" />
                <el-table-column prop="w_crnt_cp2" label="压机流2" width="100" />
                <el-table-column prop="w_freq_cp1" label="压机频1" width="100" />
                <el-table-column prop="w_freq_cp2" label="压机频2" width="100" />
                <el-table-column prop="w_crnt_ef1" label="送风流1" width="100" />
                <el-table-column prop="w_crnt_ef2" label="送风流2" width="100" />
                <el-table-column prop="i_high_pres1" label="高压1" width="90" />
                <el-table-column prop="i_low_pres1" label="低压1" width="90" />
                <el-table-column prop="i_high_pres2" label="高压2" width="90" />
                <el-table-column prop="i_low_pres2" label="低压2" width="90" />
                <el-table-column prop="i_sat_u1" label="送风温1" width="100" />
                <el-table-column prop="i_sat_u2" label="送风温2" width="100" />
                <el-table-column prop="dwcf_op_tm1" label="冷凝时1" width="120" />
                <el-table-column prop="dwcf_op_tm2" label="冷凝时2" width="120" />
                <el-table-column prop="dwcp_op_tm1" label="压机时1" width="120" />
                <el-table-column prop="dwcp_op_tm2" label="压机时2" width="120" />
                <el-table-column prop="dwexufan_op_tm" label="废排时" width="120" />
                <el-table-column prop="dwfad_op_cnt" label="新风开度" width="100" />
                <el-table-column prop="dwrad_op_cnt" label="回风开度" width="100" />
                <el-table-column prop="event_time" label="采集时间" width="180" fixed="right" />
            </el-table>

            <div class="detail-pagination">
                <el-pagination
                    v-model:current-page="detailPage"
                    v-model:page-size="detailPageSize"
                    :page-sizes="[10, 20, 50]"
                    layout="total, sizes, prev, pager, next, jumper"
                    :total="detailTotal"
                    @size-change="handleDetailSizeChange"
                    @current-change="fetchDetailData"
                />
            </div>
        </el-dialog>
    </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'

import { useRouter, useRoute } from 'vue-router'
import { getTrainDataApi } from '@/api/api'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()

// 1. 筛选状态
const filterForm = reactive({
    trainNo: String(route.query.trainNo || '7001'),
    carriageNo: String(route.query.trainCoach || '1'),
    timeRange: [
        dayjs().subtract(24, 'hour').format('YYYY-MM-DD HH:mm:ss'),
        dayjs().format('YYYY-MM-DD HH:mm:ss')
    ]
})

// 监听路由参数变化
watch(() => route.query, (newQuery) => {
    if (newQuery.trainNo) filterForm.trainNo = String(newQuery.trainNo)
    if (newQuery.trainCoach) filterForm.carriageNo = String(newQuery.trainCoach)
    handleQuery()
}, { deep: true })


// 2. 基础选择数据
const trainList = ref(Array.from({ length: 40 }, (_, i) => ({
    train_id: (7001 + i).toString()
})))

const carriageList = ref(Array.from({ length: 6 }, (_, i) => ({
    carriage_no: (i + 1).toString(),
    label: (i + 1) + '车厢'
})))

// 3. 主表格数据 (按天聚合)
const loading = ref(false)
const summaryData = ref([])

// 辅助函数：格式化机组名称
const formatUnit = (id) => {
    if (!id) return '-'
    const code = id.toLowerCase()
    if (code.includes('u1')) return '机组一'
    if (code.includes('u2')) return '机组二'
    return '-'
}


const handleQuery = () => {
    loading.value = true
    summaryData.value = []
    
    if (!filterForm.timeRange || filterForm.timeRange.length < 2) {
        loading.value = false
        return
    }

    const start = dayjs(filterForm.timeRange[0])
    const end = dayjs(filterForm.timeRange[1])
    const daysDiff = end.diff(start, 'day')
    
    // 生成日期行
    for (let i = 0; i <= daysDiff; i++) {
        const currentDate = start.add(i, 'day').format('YYYY-MM-DD')
        summaryData.value.push({
            trainNo: filterForm.trainNo,
            carriageNo: filterForm.carriageNo,
            date: currentDate
        })
    }
    
    loading.value = false
}

// 4. 详情展示逻辑
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailList = ref([])
const selectedDate = ref('')
const detailPage = ref(1)
const detailPageSize = ref(10)
const detailTotal = ref(0)

const handleDetailSizeChange = (val) => {
    detailPageSize.value = val
    detailPage.value = 1
    fetchDetailData()
}

const handleViewDetail = (row) => {
    selectedDate.value = row.date
    detailVisible.value = true
    detailPage.value = 1
    fetchDetailData()
}

const fetchDetailData = async () => {
    detailLoading.value = true
    try {
        const fullCarId = `${filterForm.trainNo}0${filterForm.carriageNo}`
        const params = {
            state: fullCarId,
            startTime: selectedDate.value + ' 00:00:00',
            endTime: selectedDate.value + ' 23:59:59',
            page: detailPage.value,
            limit: detailPageSize.value
        }
        
        // 假设接口 getTrainDataApi 返回明细
        const res = await getTrainDataApi(params)
        if (res && res.data) {
            detailList.value = res.data.list || []
            detailTotal.value = res.data.total || 0
        } else {
            detailList.value = []
            detailTotal.value = 0
        }
    } catch (err) {
        console.error('获取明细数据失败:', err)
        detailList.value = []
        detailTotal.value = 0
    } finally {
        detailLoading.value = false
    }
}

// 5. 导出逻辑
const handleExportAll = () => {
    ElMessage.success('正在准备全量数据导出，请稍候...')
}

const handleDownloadRow = (row) => {
    ElMessage.success(`正在下载 ${row.date} 的数据包...`)
}

const goBack = () => {
    router.back()
}

const initializeTime = () => {
    // 默认当前至24小时前
    const now = dayjs()
    const yesterday = now.subtract(24, 'hour')
    filterForm.timeRange = [yesterday.format('YYYY-MM-DD HH:mm:ss'), now.format('YYYY-MM-DD HH:mm:ss')]
}

onMounted(() => {
    initializeTime()
    handleQuery()
})
</script>

<style scoped lang="scss">
.warp2 {
    background: #0a0f1d;
    min-height: 100vh;
    padding: 0;
    color: #fff;
}

.header-monitor {
    height: 50px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 20px;
    background: linear-gradient(180deg, #181f30 0%, #0a0f1d 100%);
    position: sticky;
    top: 0;
    z-index: 1000;
    border-bottom: 1px solid rgba(33, 134, 207, 0.2);

    .header-left {
        display: flex;
        align-items: center;
        gap: 15px;
        .page-title {
            color: #2186cf;
            font-size: 16px;
            font-weight: bold;
            text-shadow: 0 0 10px rgba(33, 134, 207, 0.3);
        }
    }

    .header-right {
        .filter-form {
            display: flex;
            align-items: center;
            :deep(.el-form-item) {
                margin-bottom: 0 !important;
                margin-right: 12px !important;
                display: flex;
                align-items: center;
            }
            :deep(.el-form-item__label) {
                color: #2186cf;
                font-weight: bold;
                padding-right: 6px;
                white-space: nowrap;
                font-size: 13px;
            }
        }
    }
}

.section-box {
    background: #141a2e;
    border-radius: 8px;
    border: 1px solid #262e45;
    padding: 15px;
    box-shadow: 0 6px 15px rgba(0,0,0,0.3);
}

.section-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 15px;
    padding: 0;

    .title-icon {
        width: 4px;
        height: 16px;
        background: #2186cf;
        border-radius: 2px;
    }

    .title-text {
        color: #ffffff;
        font-size: 15px;
        font-weight: bold;
    }
}

.nav-btn, .export-btn {
    background: #0a0f1d !important;
    border: 1px solid #2186cf !important;
    color: #ffffff !important;
    font-size: 13px !important;
    font-weight: 500 !important;
    height: 32px !important;
    padding: 0 15px !important;
    border-radius: 4px !important;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
    display: flex;
    align-items: center;
    justify-content: center;
    
    &:hover, &:focus {
        background: rgba(33, 134, 207, 0.2) !important;
        border-color: #409eff !important;
        color: #ffffff !important;
        box-shadow: 0 0 10px rgba(33, 134, 207, 0.3) !important;
    }

    /* 图标样式微调 */
    :deep(.el-icon) {
        font-size: 14px !important;
        margin-right: 5px !important;
        color: #2186cf !important;
    }
}

:deep(.el-table) {
    background: transparent !important;
    color: #d1d9e7;
    --el-table-border-color: #262e45;
    --el-table-header-bg-color: #1a2234;
    --el-table-tr-bg-color: transparent;
    --el-table-row-hover-bg-color: rgba(33, 134, 207, 0.1);
    
    th.el-table__cell {
        background: #1a2234 !important;
        color: #2186cf;
        font-weight: bold;
        border-bottom: 2px solid #2186cf;
    }
    
    td.el-table__cell {
        border-bottom: 1px solid #262e45;
    }

    .el-table__inner-wrapper::before {
        display: none;
    }
}

/* 详情对话框样式 */
:deep(.history-detail-dialog) {
    background: #0a1124 !important;
    border: 1px solid #2186cf !important;
    box-shadow: 0 10px 30px rgba(0,0,0,0.8);
    border-radius: 8px;
    
    .el-dialog__header {
        margin: 0;
        padding: 15px 20px;
        border-bottom: 1px solid rgba(33, 134, 207, 0.2);
        .el-dialog__title { 
            color: #2186cf; 
            font-weight: bold;
            font-size: 16px;
        }
    }
    
    .el-dialog__body { 
        padding: 20px; 
        background: #0a0f1d;
    }
    
    .el-dialog__headerbtn .el-dialog__close {
        color: #2186cf;
        &:hover { color: #fff; }
    }
}

.detail-filter {
    margin-bottom: 15px;
    .detail-info {
        color: #2186CF;
        font-weight: bold;
        font-size: 14px;
    }
}

.detail-pagination {
    margin-top: 20px;
    display: flex;
    justify-content: center;
}

/* 统一 Select 样式 */
.train-select, .carriage-select {
    width: 110px !important;
}
</style>

<style lang="scss">
/* 全局样式覆盖，确保暗色主题和蓝色边框生效 */
.filter-form {
    --el-fill-color-blank: #0a0f1d !important;
    --el-border-color: #2186cf !important;
    --el-text-color-regular: #ffffff !important;
    --el-border-color-hover: #409eff !important;

    .el-input__wrapper {
        background-color: #0a0f1d !important;
        box-shadow: 0 0 0 1px #2186cf inset !important;
        border-radius: 4px !important;
        height: 32px !important;
        
        &:hover {
            box-shadow: 0 0 0 1px #409eff inset, 0 0 10px rgba(33, 134, 207, 0.4) !important;
        }
        
        &.is-focus {
            box-shadow: 0 0 0 1px #409eff inset, 0 0 15px rgba(33, 134, 207, 0.6) !important;
        }
        
        .el-input__inner {
            color: #ffffff !important;
            font-size: 13px !important;
        }
    }

    /* 时间选择器特定覆盖 */
    .el-range-editor.el-input__wrapper {
        width: 350px !important;
    }
    .el-range-input {
        color: #fff !important;
    }
    .el-range-separator {
        color: #2186cf !important;
    }
}

/* 下拉菜单与时间弹出层样式对齐 */
.el-select__popper.el-popper, 
.el-picker__popper.el-popper {
    background-color: #141a2e !important;
    border: 1px solid #2186cf !important;
    box-shadow: 0 4px 12px rgba(0,0,0,0.5) !important;
    color: #fff !important;

    .el-select-dropdown__item {
        color: #d1d9e7 !important;
        &:hover, &.hover {
            background-color: rgba(33, 134, 207, 0.2) !important;
            color: #ffffff !important;
        }
        &.selected {
            color: #2186cf !important;
            font-weight: bold;
            background-color: rgba(33, 134, 207, 0.1) !important;
        }
    }

    /* 时间选择器内部样式深度覆盖 */
    .el-picker-panel {
        background-color: #141a2e !important;
        color: #fff !important;
        border: none !important;
    }

    .el-date-range-picker__time-header,
    .el-date-range-picker__header {
        border-bottom: 1px solid #262e45 !important;
        .el-input__wrapper {
            background-color: #0a0f1d !important;
            box-shadow: 0 0 0 1px #2186cf inset !important;
            .el-input__inner { color: #fff !important; }
        }
    }

    .el-date-range-picker__content.is-left {
        border-right: 1px solid #262e45 !important;
    }
    .el-date-table th {
        color: #2186cf !important;
        border-bottom: 1px solid #262e45 !important;
    }
    .el-date-table td.next-month, .el-date-table td.prev-month {
        color: #4e5969 !important;
    }
    .el-date-table td.available:hover {
        color: #2186cf !important;
    }
    .el-date-table td.in-range .el-date-table-cell {
        background-color: rgba(33, 134, 207, 0.1) !important;
    }
    .el-time-panel {
        background-color: #141a2e !important;
        border: 1px solid #2186cf !important;
    }
    .el-time-spinner__item {
        color: #d1d9e7 !important;
        &:hover { background: rgba(33, 134, 207, 0.2) !important; }
        &.is-active { color: #2186cf !important; font-weight: bold; }
    }
    .el-picker-panel__footer {
        background-color: #141a2e !important;
        border-top: 1px solid #262e45 !important;
        padding: 10px !important;
        
        .el-button {
            background: #0a0f1d !important;
            border: 1px solid #2186cf !important;
            color: #ffffff !important;
            height: 28px !important;
            padding: 0 12px !important;
            font-size: 12px !important;
            
            &.is-text {
                 /* 清空按钮 */
                border: 1px solid #2186cf !important;
            }
            
            &--primary {
                /* 确定按钮 */
                background: #2186cf !important;
                border-color: #2186cf !important;
                &:hover {
                    background: #409eff !important;
                    border-color: #409eff !important;
                }
            }

            &:hover {
                background: rgba(33, 134, 207, 0.2) !important;
                border-color: #409eff !important;
                color: #ffffff !important;
            }
        }
    }
}

/* 分页组件样式深度对齐 */
.el-pagination {
    --el-pagination-bg-color: transparent !important;
    --el-pagination-button-bg-color: transparent !important;
    --el-pagination-button-color: #ffffff !important;
    --el-pagination-button-disabled-bg-color: transparent !important;
    
    /* 强力覆盖分页内部的 select 和 input */
    .el-select, .el-input {
        --el-fill-color-blank: #0a0f1d !important;
        --el-border-color: #2186cf !important;
        --el-text-color-regular: #ffffff !important;
        
        .el-input__wrapper {
            background-color: #0a0f1d !important;
            box-shadow: 0 0 0 1px #2186cf inset !important;
            background: #0a0f1d !important;
            border: none !important; /* 移除外层可能的边框线 */
            
            &:hover {
                box-shadow: 0 0 0 1px #409eff inset !important;
            }
            
            &.is-focus {
                box-shadow: 0 0 0 1px #409eff inset !important;
            }
        }
        
        .el-input__inner {
            color: #ffffff !important;
            background-color: transparent !important;
            border: none !important; /* 移除内层灰色边框线 */
            box-shadow: none !important;
            text-align: center;
        }
    }
    
    .el-pagination__total, .el-pagination__jump {
        color: #d1d9e7 !important;
    }
    
    .el-pager li {
        background: transparent !important;
        color: #d1d9e7 !important;
        &.is-active {
            color: #2186cf !important;
            font-weight: bold;
        }
    }
}
</style>
