<template>
    <div class="warp">
        <!-- 实时报警 -->
        <div class="header">
            <h4>实时报警</h4>
        </div>
        <el-table :data="props.ActualWarningData" stripe height="245px" style="width: 100%;">
            <el-table-column prop="carriage_no" label="车厢号" min-width="15%"></el-table-column>
            <el-table-column prop="name" label="报警名称" min-width="45%" show-overflow-tooltip></el-table-column>
            <el-table-column prop="alarm_time" label="报警时间" min-width="25%" show-overflow-tooltip></el-table-column>
            <el-table-column label="操作" min-width="15%">
                <template #default="scope">
                    <el-popover
                        class="item" effect="light"
                        placement="bottom"
                        title="指导建议"
                        width="200"
                        trigger="hover"
                        :content="scope.row.precautions">
                        <template #reference>
                         <el-link type="primary" underline="never" :class="['btn_text']">
                            指导建议
                        </el-link>
                        </template>
                    </el-popover>
                    <!-- <el-tooltip class="item" effect="light" :content="scope.row.precautions" placement="left">
                        <el-link type="primary" :underline="false" :class="['btn_text']">
                            指导建议
                        </el-link>
                    </el-tooltip> -->
                </template>
            </el-table-column>
            <template #empty>
                <img src="/src/assets/img/no-data.svg" width="40" />
                <p style="font-size:12px">当前暂无数据</p>
            </template>
        </el-table>
    </div>
</template>

<script setup>
// import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
const props = defineProps({
  ActualWarningData: {
    type: Array,
    default: []
  }
})
const router = useRouter();

// const ActualWarningData = [
//     {
//         carriage_no: 'Tc1',
//         name: '空调蒸发器A报警',
//         alarm_time: '16:12:00',
//         precautions: '建议'
//     }, {
//         carriage_no: 'Mp1',
//         name: '空调蒸发器A报警',
//         alarm_time: '16:12:00',
//         precautions: '建议'
//     }, {
//         carriage_no: 'M1',
//         name: '空调蒸发器A报警',
//         alarm_time: '16:12:00',
//         precautions: '建议'
//     }, {
//         carriage_no: 'M2',
//         name: '空调蒸发器A报警',
//         alarm_time: '16:12:00',
//         precautions: '建议'
//     }, {
//         carriage_no: 'MP2',
//         name: '空调蒸发器A报警',
//         alarm_time: '16:12:00',
//         precautions: '建议'
//     }, {
//         carriage_no: 'Tc2',
//         name: '空调蒸发器A报警',
//         alarm_time: '16:12:00',
//         precautions: '建议'
//     }
// ]
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
    border-radius: 10px;
    background-color: #181F30 !important;
}

.el-table {
    --el-table-border-color: none;
}
/* 标题左边线 */
.header {
    height: 50px;
    box-sizing: border-box;
    top: 0;
    left: 0;
    display: flex;
    align-items: center;
    gap: 30px;
}

.header>h4 {
    font-size: 16px;
    position: relative;
    margin-left: 30px;
    line-height: 1em;
}

.header>h4::before {
    position: absolute;
    content: "";
    width: 4px;
    height: 1em;
    background: #e65355;
    margin-left: -10px;
}
.btn_text {
    font-size: 14px;
    color: #2186CF !important;
}
</style>
