<template>
    <div class="warp">
        <div class="section-header">
            <span class="title-icon"></span>
            <span class="title-text">实时预警</span>
        </div>
        <el-table :data="props.StateWarningData" stripe height="245px" style="width: 100%;">
            <el-table-column prop="train_no" label="车号" min-width="14%" show-overflow-tooltip>
                <template #default="scope">{{ formatTrainNo(scope.row.train_no) }}</template>
            </el-table-column>
            <el-table-column prop="carriage_no" label="车厢" min-width="12%">
                <template #default="scope">{{ getCarriageName(scope.row.carriage_no) }}</template>
            </el-table-column>
            <el-table-column label="机组" min-width="12%">
                <template #default="scope">{{ formatUnit(scope.row.name) }}</template>
            </el-table-column>
            <el-table-column prop="name" label="预警名称" min-width="35%" show-overflow-tooltip></el-table-column>
            <el-table-column prop="warning_time" label="时间" min-width="20%" show-overflow-tooltip></el-table-column>
            <el-table-column label="建议" min-width="10%">
                <template #default="scope">
                    <el-popover popper-class="advice-popover" placement="left" title="指导建议"
                        width="240" trigger="hover" :content="scope.row.precautions || '请联系专业维修人员处理'">
                        <template #reference>
                            <el-link type="primary" underline="never" class="btn_text">建议</el-link>
                        </template>
                    </el-popover>
                </template>
            </el-table-column>
            <template #empty>
                <img src="/img/no-data.svg" width="40" />
                <p style="font-size:12px">当前暂无数据</p>
            </template>
        </el-table>
    </div>
</template>

<script setup>
const props = defineProps({
  StateWarningData: { type: Array, default: () => [] }
})

const CARRIAGE_MAP = { '1':'TC1','2':'MP1','3':'M1','4':'M2','5':'MP2','6':'TC2' }
const getCarriageName = (no) => CARRIAGE_MAP[String(no)] || `${no}车`
const formatTrainNo = (no) => no ? `0${no}` : '-'
const formatUnit = (name) => {
  if (!name) return '-'
  const n = name.toLowerCase()
  if (n.includes('u2') || n.includes('机组2')) return '机组二'
  if (n.includes('u1') || n.includes('机组1')) return '机组一'
  return '-'
}
</script>

<style scoped>
.el-table,
.el-table tr,
.el-table td,
.el-table th {
    background-color: transparent !important;
}

.warp {
    width: 100%;
}

.el-table {
    --el-table-border-color: none;
}
.section-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
    padding: 0;

    .title-icon {
        width: 4px;
        height: 16px;
        background: #ff9800;
        border-radius: 2px;
    }

    .title-text {
        color: #ffffff;
        font-size: 15px;
        font-weight: bold;
    }
}
.btn_text {
    font-size: 14px;
    color: #2186CF !important;
}
</style>
