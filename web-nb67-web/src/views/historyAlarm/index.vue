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
                            :shortcuts="shortcuts"
                            :default-time="[new Date(2000, 1, 1, 0, 0, 0), new Date(2000, 1, 1, 23, 59, 59)]"
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
                    <span class="title-text">历史报警</span>
                </div>
                <el-table :data="tableData" border stripe style="width: 100%" v-loading="loading">
                    <el-table-column prop="train_id" label="列车号" align="center" width="100" />
                    <el-table-column prop="carriage_no" label="车厢号" align="center" width="100">
                        <template #default="scope">{{ scope.row.carriage_no }}车厢</template>
                    </el-table-column>
                    <el-table-column label="机组" align="center" width="100">
                        <template #default="scope">
                            {{ formatUnit(scope.row.fault_code) }}
                        </template>
                    </el-table-column>
                    <el-table-column prop="fault_level" label="严重级别" align="center" width="100">
                        <template #default="scope">
                            <el-tag :type="scope.row.fault_level === '严重' ? 'danger' : (scope.row.fault_level === '轻微' ? 'info' : 'warning')">
                                {{ scope.row.fault_level }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column prop="fault_desc" label="故障详情" header-align="center" min-width="250" show-overflow-tooltip />
                    <el-table-column prop="occurrence_time" label="故障时间" align="center" width="180" />
                </el-table>

                <div class="pagination-container">
                    <el-pagination
                        :current-page="currentPage"
                        :page-size="pageSize"
                        :page-sizes="[10, 20, 50]"
                        layout="total, sizes, prev, pager, next, jumper"
                        :total="total"
                        @size-change="handleSizeChange"
                        @current-change="handleCurrentChange"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getAlarmInformation } from '@/api/api'
import dayjs from 'dayjs'

const router = useRouter()
const route = useRoute()

// ============================================================
// 1. UI 筛选表单 —— 只绑定 UI 组件，不直接参与查询
// ============================================================
const shortcuts = [
  {
    text: '最近1小时',
    value: () => {
      const end = dayjs()
      const start = dayjs().subtract(1, 'hour')
      return [start.format('YYYY-MM-DD HH:mm:ss'), end.format('YYYY-MM-DD HH:mm:ss')]
    },
  },
  {
    text: '最近24小时',
    value: () => {
      const end = dayjs()
      const start = dayjs().subtract(24, 'hour')
      return [start.format('YYYY-MM-DD HH:mm:ss'), end.format('YYYY-MM-DD HH:mm:ss')]
    },
  },
  {
    text: '最近7天',
    value: () => {
      const end = dayjs()
      const start = dayjs().subtract(7, 'day')
      return [start.format('YYYY-MM-DD HH:mm:ss'), end.format('YYYY-MM-DD HH:mm:ss')]
    },
  }
]

// filterForm 仅绑定 UI 组件（el-select / el-date-picker）
// DatePicker blur/关闭时可能将 timeRange 置为 null，但不影响实际查询
const filterForm = reactive({
    trainNo: String(route.query.trainNo || '7001'),
    carriageNo: String(route.query.trainCoach || '1'),
    timeRange: [
        dayjs().subtract(24, 'hour').format('YYYY-MM-DD HH:mm:ss'),
        dayjs().format('YYYY-MM-DD HH:mm:ss')
    ]
})

// ============================================================
// 2. 查询参数锁 —— fetchData 唯一的数据来源
//    与 filterForm 完全解耦，只有点「查询」按钮才会更新
//    分页操作（翻页/改 pageSize）永远不会修改此对象
// ============================================================
const queryParams = reactive({
    trainNo: filterForm.trainNo,
    carriageNo: filterForm.carriageNo,
    startTime: filterForm.timeRange[0],
    endTime: filterForm.timeRange[1]
})

// 监听路由参数变化（只在 trainNo/trainCoach 真正改变时触发）
watch(
    () => [route.query.trainNo, route.query.trainCoach],
    (newVals, oldVals) => {
        const newTrainNo = newVals[0]
        const newCoach = newVals[1]
        const oldTrainNo = oldVals ? oldVals[0] : undefined
        const oldCoach = oldVals ? oldVals[1] : undefined
        if (newTrainNo !== oldTrainNo || newCoach !== oldCoach) {
            if (newTrainNo) filterForm.trainNo = String(newTrainNo)
            if (newCoach) filterForm.carriageNo = String(newCoach)
            handleQuery()
        }
    }
)

// 辅助函数：格式化机组名称
const formatUnit = (faultCode) => {
    if (!faultCode) return '-'
    const code = faultCode.toLowerCase()
    if (code.includes('u1')) return '机组一'
    if (code.includes('u2')) return '机组二'
    return '-'
}

// ============================================================
// 3. 基础数据
// ============================================================
const trainList = ref(Array.from({ length: 40 }, (_, i) => ({
    train_id: (7001 + i).toString()
})))

const carriageList = ref(Array.from({ length: 6 }, (_, i) => ({
    carriage_no: (i + 1).toString(),
    label: (i + 1) + '车厢'
})))

// ============================================================
// 4. 表格状态
// ============================================================
const loading = ref(false)
const tableData = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

/**
 * 核心数据获取函数
 * ★ 只从 queryParams 读时间/列车/车厢参数，永远不读 filterForm
 * ★ 分页相关的 page/limit 在进入函数时立即快照，防止并发修改
 */
const fetchData = async () => {
    loading.value = true
    // 立即快照分页参数
    const page = currentPage.value
    const limit = pageSize.value
    // 从稳定的查询锁读参数
    const fullCarId = `${queryParams.trainNo}0${queryParams.carriageNo}1`
    const params = {
        state: fullCarId,
        startTime: queryParams.startTime,
        endTime: queryParams.endTime,
        page: page,
        limit: limit
    }
    console.log('[HistoryAlarm] fetchData:', JSON.stringify(params))

    try {
        const res = await getAlarmInformation(params)
        console.log('[HistoryAlarm] Response → code:', res?.code, 'total:', res?.data?.total, 'list:', res?.data?.list?.length)

        if (res && res.code === 200 && res.data) {
            tableData.value = res.data.list || []
            total.value = res.data.total || 0
        } else if (res && Array.isArray(res.data)) {
            tableData.value = res.data
            total.value = res.total || 0
        } else if (Array.isArray(res)) {
            tableData.value = res
            total.value = res.length
        } else {
            tableData.value = []
            total.value = 0
        }
    } catch (err) {
        console.error('[HistoryAlarm] 请求失败:', err)
        tableData.value = []
        total.value = 0
    } finally {
        loading.value = false
    }
}

/**
 * 点击「查询」按钮
 * ★ 这是唯一将 filterForm 同步到 queryParams 的入口
 * ★ 时间无效时保留上次有效值（不传空给后端）
 */
const handleQuery = () => {
    // 同步车号/车厢
    queryParams.trainNo = filterForm.trainNo
    queryParams.carriageNo = filterForm.carriageNo
    // 只在时间有效时才更新（DatePicker blur 可能传 null，此时保留上次有效值）
    if (filterForm.timeRange && filterForm.timeRange[0] && filterForm.timeRange[1]) {
        queryParams.startTime = filterForm.timeRange[0]
        queryParams.endTime = filterForm.timeRange[1]
    }
    currentPage.value = 1
    fetchData()
}

/**
 * 改变每页条数
 * 必须重置 currentPage=1，否则 offset=(page-1)*newLimit 超出数据量→空页
 */
const handleSizeChange = (val) => {
    pageSize.value = val
    currentPage.value = 1
    fetchData()
}

/**
 * 切换页码
 */
const handleCurrentChange = (val) => {
    currentPage.value = val
    fetchData()
}

const goBack = () => {
    router.back()
}

onMounted(() => {
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

.monitor-container {
    padding: 15px;
    display: flex;
    flex-direction: column;
    gap: 12px;
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
    display: flex;
    align-items: center;
    justify-content: center;

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

.pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
}

.train-select, .carriage-select {
    width: 110px !important;
}
</style>

<style lang="scss">
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
                border: 1px solid #2186cf !important;
            }

            &--primary {
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

.el-pagination {
    --el-pagination-bg-color: transparent !important;
    --el-pagination-button-bg-color: transparent !important;
    --el-pagination-button-color: #ffffff !important;
    --el-pagination-button-disabled-bg-color: transparent !important;

    .el-select, .el-input {
        --el-fill-color-blank: #0a0f1d !important;
        --el-border-color: #2186cf !important;
        --el-text-color-regular: #ffffff !important;

        .el-input__wrapper {
            background-color: #0a0f1d !important;
            box-shadow: 0 0 0 1px #2186cf inset !important;
            background: #0a0f1d !important;
            border: none !important;

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
            border: none !important;
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
