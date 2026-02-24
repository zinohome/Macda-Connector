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
                    <el-table-column label="操作" align="center" width="200">
                        <template #default="scope">
                            <el-button link type="primary" icon="Download" :loading="scope.row.downloading" @click="handleDownloadRow(scope.row)">下载 ZIP</el-button>
                            <el-button link type="primary" icon="View" @click="handleViewDetail(scope.row)">查看</el-button>
                        </template>
                    </el-table-column>
                </el-table>
            </div>
        </div>

        <!-- 详情对话框：暗色主题 -->
        <el-dialog
            v-model="detailVisible"
            title="历史数据详情"
            width="92%"
            top="4vh"
            draggable
            :append-to-body="true"
            class="dark-dialog"
        >
            <template #header>
                <div class="dialog-header">
                    <span class="dialog-title">历史数据详情</span>
                    <span class="dialog-sub">{{ selectedDate }} | {{ filterForm.trainNo }}号车 | {{ filterForm.carriageNo }}车厢</span>
                </div>
            </template>

            <el-table
                :data="detailList"
                border
                stripe
                style="width: 100%"
                height="58vh"
                v-loading="detailLoading"
                element-loading-background="rgba(14,22,40,0.92)"
                :scrollbar-always-on="true"
            >
                <!-- 前 4 列固定左侧 -->
                <el-table-column type="index"       label="序号"         width="60"  fixed="left" align="center" />
                <el-table-column prop="train_id"    label="列车号"         width="90"  fixed="left" show-overflow-tooltip />
                <el-table-column prop="carriage_id" label="车厢号"         width="80"  fixed="left" show-overflow-tooltip>
                    <template #default="scope">第{{ scope.row.carriage_id }}车厢</template>
                </el-table-column>
                <el-table-column prop="unit_name"   label="机组"           width="90"  fixed="left" show-overflow-tooltip />
                <!-- 中间可滚动字段 -->
                <el-table-column prop="i_inner_temp"  label="室内温度"       width="100" show-overflow-tooltip />
                <el-table-column prop="i_outer_temp"  label="新风温度"       width="100" show-overflow-tooltip />
                <el-table-column prop="i_set_temp"    label="设定温度"       width="100" show-overflow-tooltip />
                <el-table-column prop="i_hum"         label="湿度"           width="80"  show-overflow-tooltip />
                <el-table-column prop="i_co2"         label="二氧化碗浓度"     width="120" show-overflow-tooltip />
                <el-table-column prop="work_status"   label="工作状态"       width="100" show-overflow-tooltip />
                <el-table-column prop="dwef_op_tm"    label="通风机运行时间"   width="130" show-overflow-tooltip />
                <el-table-column prop="w_crnt_cf1"    label="冷凝风机电流1"     width="120" show-overflow-tooltip />
                <el-table-column prop="w_crnt_cf2"    label="冷凝风机电流2"     width="120" show-overflow-tooltip />
                <el-table-column prop="w_crnt_cp1"    label="压缩机电流1"       width="120" show-overflow-tooltip />
                <el-table-column prop="w_crnt_cp2"    label="压缩机电流2"       width="120" show-overflow-tooltip />
                <el-table-column prop="w_freq_cp1"    label="压缩机频率1"       width="120" show-overflow-tooltip />
                <el-table-column prop="w_freq_cp2"    label="压缩机频率2"       width="120" show-overflow-tooltip />
                <el-table-column prop="w_crnt_ef1"    label="送风机电流1"       width="120" show-overflow-tooltip />
                <el-table-column prop="w_crnt_ef2"    label="送风机电流2"       width="120" show-overflow-tooltip />
                <el-table-column prop="i_high_pres1"  label="高压压力1"         width="110" show-overflow-tooltip />
                <el-table-column prop="i_low_pres1"   label="低压压力1"         width="110" show-overflow-tooltip />
                <el-table-column prop="i_high_pres2"  label="高压压力2"         width="110" show-overflow-tooltip />
                <el-table-column prop="i_low_pres2"   label="低压压力2"         width="110" show-overflow-tooltip />
                <el-table-column prop="i_sat1"        label="送风温度1"         width="110" show-overflow-tooltip />
                <el-table-column prop="i_sat2"        label="送风温度2"         width="110" show-overflow-tooltip />
                <el-table-column prop="dwcf_op_tm1"   label="冷凝风机运行时间1" width="150" show-overflow-tooltip />
                <el-table-column prop="dwcf_op_tm2"   label="冷凝风机运行时间2" width="150" show-overflow-tooltip />
                <el-table-column prop="dwcp_op_tm1"   label="压缩机运行时间1"   width="140" show-overflow-tooltip />
                <el-table-column prop="dwcp_op_tm2"   label="压缩机运行时间2"   width="140" show-overflow-tooltip />
                <el-table-column prop="dwexufan_op_tm" label="废排风机运行时间" width="140" show-overflow-tooltip />
                <el-table-column prop="dwfad_op_cnt"  label="新风阀开度"       width="110" show-overflow-tooltip />
                <el-table-column prop="dwrad_op_cnt"  label="回风阀开度"       width="110" show-overflow-tooltip />
                <!-- 最后 1 列固定右侧 -->
                <el-table-column prop="ingest_time"   label="数据采集时间"     width="185" fixed="right" show-overflow-tooltip />
            </el-table>

            <div class="detail-pagination">
                <el-pagination
                    v-model:current-page="detailPage"
                    v-model:page-size="detailPageSize"
                    :page-sizes="[20, 50, 100]"
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
import { getTrainDataApi, getTrainDataDatesApi } from '@/api/api'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import JSZip from 'jszip'
import { saveAs } from 'file-saver'

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

const handleQuery = async () => {
    loading.value = true
    summaryData.value = []

    if (!filterForm.timeRange || filterForm.timeRange.length < 2) {
        loading.value = false
        return
    }

    try {
        const carriageStr = String(filterForm.carriageNo).replace(/^0+/, '').padStart(2, '0')
        const fullCarId = `${filterForm.trainNo}${carriageStr}`
        const res = await getTrainDataDatesApi({
            state: fullCarId,
            startTime: filterForm.timeRange[0],
            endTime: filterForm.timeRange[1]
        })
        
        if (res && res.code === 200 && Array.isArray(res.data)) {
            // 后端直接返回有数据的日期列表
            res.data.forEach(dateStr => {
                summaryData.value.push({
                    trainNo: filterForm.trainNo,
                    carriageNo: filterForm.carriageNo,
                    date: dateStr,
                    downloading: false
                })
            })
        }
    } catch (err) {
        console.error("Failed to fetch dates", err)
    } finally {
        loading.value = false
    }
}

// 4. 详情展示逻辑
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailList = ref([])
const selectedDate = ref('')
const detailPage = ref(1)
const detailPageSize = ref(20)
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
        const carriageStr = String(filterForm.carriageNo).replace(/^0+/, '').padStart(2, '0')
        const fullCarId = `${filterForm.trainNo}${carriageStr}`
        const params = {
            state: fullCarId,
            startTime: selectedDate.value + ' 00:00:00',
            endTime: selectedDate.value + ' 23:59:59',
            page: detailPage.value,
            limit: detailPageSize.value
        }

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

// 5. 按天下载 ZIP（拉取全天数据，打包成 CSV）
// 与表格一致的 CSV 标题
const CSV_HEADERS = [
    '序号', '列车号', '车厢号', '机组',
    '室内温度', '新风温度', '设定温度', '湿度', '二氧化碳浓度',
    '工作状态', '通风机运行时间',
    '冷凝风机电流1', '冷凝风机电流2',
    '压缩机电流1', '压缩机电流2',
    '压缩机频率1', '压缩机频率2',
    '送风机电流1', '送风机电流2',
    '高压压力1', '低压压力1', '高压压力2', '低压压力2',
    '送风温度1', '送风温度2',
    '冷凝风机运行时间1', '冷凝风机运行时间2',
    '压缩机运行时间1', '压缩机运行时间2',
    '废排风机运行时间', '新风阀开度', '回风阀开度',
    '数据采集时间', '数据设备时间'
]

const rowToCsv = (row, index) => {
    const fields = [
        index + 1, row.train_id, `第${row.carriage_id}车厢`, row.unit_name,
        row.i_inner_temp, row.i_outer_temp, row.i_set_temp, row.i_hum, row.i_co2,
        row.work_status, row.dwef_op_tm,
        row.w_crnt_cf1, row.w_crnt_cf2,
        row.w_crnt_cp1, row.w_crnt_cp2,
        row.w_freq_cp1, row.w_freq_cp2,
        row.w_crnt_ef1, row.w_crnt_ef2,
        row.i_high_pres1, row.i_low_pres1, row.i_high_pres2, row.i_low_pres2,
        row.i_sat1, row.i_sat2,
        row.dwcf_op_tm1, row.dwcf_op_tm2,
        row.dwcp_op_tm1, row.dwcp_op_tm2,
        row.dwexufan_op_tm, row.dwfad_op_cnt, row.dwrad_op_cnt,
        row.ingest_time, row.event_time
    ]
    return fields.map(v => (v === undefined || v === null) ? '' : String(v)).join(',')
}

const handleDownloadRow = async (row) => {
    row.downloading = true
    try {
        const carriageStr = String(filterForm.carriageNo).replace(/^0+/, '').padStart(2, '0')
        const fullCarId = `${filterForm.trainNo}${carriageStr}`
        // 分批拉取全天数据（每批 500 条，最多 10 批 = 5000 条）
        let page = 1
        const pageLimit = 500
        let allRows = []

        while (true) {
            const res = await getTrainDataApi({
                state: fullCarId,
                startTime: row.date + ' 00:00:00',
                endTime:   row.date + ' 23:59:59',
                page,
                limit: pageLimit
            })
            const batch = res?.data?.list || []
            allRows = allRows.concat(batch)
            const total = res?.data?.total || 0
            if (allRows.length >= total || batch.length < pageLimit || page >= 10) break
            page++
        }

        if (allRows.length === 0) {
            ElMessage.warning(`${row.date} 暂无数据`)
            return
        }

        // 生成 CSV（UTF-8 BOM 确保 Excel 正常打开）
        const bom = '\uFEFF'
        const csvContent = bom + CSV_HEADERS.join(',') + '\n' + allRows.map(rowToCsv).join('\n')

        // 打包 ZIP
        const zip = new JSZip()
        const fileName = `历史数据_${filterForm.trainNo}号车_${filterForm.carriageNo}车厢_${row.date}.csv`
        zip.file(fileName, csvContent)
        const blob = await zip.generateAsync({ type: 'blob', compression: 'DEFLATE', compressionOptions: { level: 6 } })
        saveAs(blob, fileName.replace('.csv', '.zip'))

        ElMessage.success(`${row.date} 数据已下载，共 ${allRows.length} 条`)
    } catch (err) {
        console.error('下载失败:', err)
        ElMessage.error('下载失败，请检查网络或联系管理员')
    } finally {
        row.downloading = false
    }
}

const goBack = () => {
    router.push({
        name: 'trainInfo',
        query: {
            trainNo:    filterForm.trainNo,
            trainCoach: filterForm.carriageNo
        }
    })
}

const initializeTime = () => {
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

.monitor-container {
    padding: 20px;
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

.nav-btn {
    background: #0a0f1d !important;
    border: 1px solid #2186cf !important;
    color: #ffffff !important;
    font-size: 13px !important;
    font-weight: 500 !important;
    height: 32px !important;
    padding: 0 15px !important;
    border-radius: 4px !important;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;

    &:hover, &:focus {
        background: rgba(33, 134, 207, 0.2) !important;
        border-color: #409eff !important;
        color: #ffffff !important;
        box-shadow: 0 0 10px rgba(33, 134, 207, 0.3) !important;
    }

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

.detail-pagination {
    margin-top: 20px;
    display: flex;
    justify-content: center;
}

.train-select, .carriage-select {
    width: 110px !important;
}

/* 对话框头部自定义布局 */
.dialog-header {
    display: flex;
    align-items: baseline;
    gap: 16px;

    .dialog-title {
        color: #2186cf;
        font-size: 16px;
        font-weight: bold;
    }

    .dialog-sub {
        color: #8899bb;
        font-size: 13px;
    }
}
</style>

<style lang="scss">
/* 暗色对话框：全局覆盖（el-dialog appended to body，必须用非 scoped） */
.dark-dialog.el-dialog {
    background: #0e1628 !important;
    border: 1px solid #2186cf !important;
    box-shadow: 0 0 40px rgba(0, 0, 0, 0.9), 0 0 0 1px rgba(33, 134, 207, 0.3);
    border-radius: 10px;

    .el-dialog__header {
        margin: 0;
        padding: 14px 20px;
        border-bottom: 1px solid rgba(33, 134, 207, 0.25);
        background: #141b2e;
        border-radius: 10px 10px 0 0;
    }

    .el-dialog__headerbtn {
        top: 14px;
        .el-dialog__close {
            color: #2186cf;
            font-size: 18px;
            &:hover { color: #fff; }
        }
    }

    .el-dialog__body {
        padding: 20px;
        background: #0e1628;
    }

    /* 内部表格继承暗色 */
    .el-table {
        background: transparent !important;
        color: #d1d9e7 !important;
        --el-table-border-color: #262e45;
        --el-table-header-bg-color: #1a2234;
        --el-table-tr-bg-color: transparent;
        --el-table-row-hover-bg-color: rgba(33, 134, 207, 0.1);
        --el-fill-color-lighter: #1e2840;

        th.el-table__cell {
            background: #1a2234 !important;
            color: #2186cf !important;
            font-weight: bold;
            border-bottom: 2px solid #2186cf;
        }

        td.el-table__cell {
            border-bottom: 1px solid #262e45;
            background: transparent !important;
        }

        .el-table__body-wrapper {
            background: transparent !important;
        }

        .el-table__inner-wrapper::before { display: none; }
    }

    /* 内部分页 */
    .el-pagination {
        --el-pagination-bg-color: transparent;
        --el-pagination-button-bg-color: transparent;
        --el-pagination-button-color: #ffffff;
        color: #d1d9e7;

        .el-pager li {
            background: transparent !important;
            color: #d1d9e7 !important;
            &.is-active { color: #2186cf !important; font-weight: bold; }
        }

        .el-input__wrapper {
            background-color: #0a0f1d !important;
            box-shadow: 0 0 0 1px #2186cf inset !important;
            .el-input__inner { color: #fff !important; }
        }

        .el-pagination__total, .el-pagination__jump { color: #d1d9e7 !important; }
    }
}

/* 全局筛选框暗色 */
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

        &:hover { box-shadow: 0 0 0 1px #409eff inset, 0 0 10px rgba(33, 134, 207, 0.4) !important; }
        &.is-focus { box-shadow: 0 0 0 1px #409eff inset, 0 0 15px rgba(33, 134, 207, 0.6) !important; }

        .el-input__inner { color: #ffffff !important; font-size: 13px !important; }
    }

    .el-range-editor.el-input__wrapper { width: 350px !important; }
    .el-range-input { color: #fff !important; }
    .el-range-separator { color: #2186cf !important; }
}

/* 下拉菜单 & 日期弹窗 */
.el-select__popper.el-popper,
.el-picker__popper.el-popper {
    background-color: #141a2e !important;
    border: 1px solid #2186cf !important;
    box-shadow: 0 4px 12px rgba(0,0,0,0.5) !important;
    color: #fff !important;

    .el-select-dropdown__item {
        color: #d1d9e7 !important;
        &:hover, &.hover { background-color: rgba(33, 134, 207, 0.2) !important; color: #ffffff !important; }
        &.selected { color: #2186cf !important; font-weight: bold; background-color: rgba(33, 134, 207, 0.1) !important; }
    }

    .el-picker-panel { background-color: #141a2e !important; color: #fff !important; border: none !important; }

    .el-date-range-picker__time-header,
    .el-date-range-picker__header {
        border-bottom: 1px solid #262e45 !important;
        .el-input__wrapper { background-color: #0a0f1d !important; box-shadow: 0 0 0 1px #2186cf inset !important; .el-input__inner { color: #fff !important; } }
    }

    .el-date-range-picker__content.is-left { border-right: 1px solid #262e45 !important; }
    .el-date-table th { color: #2186cf !important; border-bottom: 1px solid #262e45 !important; }
    .el-date-table td.next-month, .el-date-table td.prev-month { color: #4e5969 !important; }
    .el-date-table td.available:hover { color: #2186cf !important; }
    .el-date-table td.in-range .el-date-table-cell { background-color: rgba(33, 134, 207, 0.1) !important; }
    .el-time-panel { background-color: #141a2e !important; border: 1px solid #2186cf !important; }
    .el-time-spinner__item { color: #d1d9e7 !important; &:hover { background: rgba(33, 134, 207, 0.2) !important; } &.is-active { color: #2186cf !important; font-weight: bold; } }
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
            &--primary { background: #2186cf !important; border-color: #2186cf !important; &:hover { background: #409eff !important; } }
            &:hover { background: rgba(33, 134, 207, 0.2) !important; border-color: #409eff !important; }
        }
    }
}
</style>
