<template>
    <div class="warp">
        <div class="header">
            <h4>车厢温度信息</h4>
        </div>
        <div>
            <el-table :data="props.CarTemperatureData" stripe style="width: 100%">
                <el-table-column fixed prop="name" label="" min-width="25%"></el-table-column>
                <el-table-column prop="carriageTC1" label="TC1车" min-width="25%"></el-table-column>
                <el-table-column prop="carriageMP1" label="MP1车" min-width="25%"></el-table-column>
                <el-table-column prop="carriageM1" label="M1车" min-width="25%"></el-table-column>
                <el-table-column prop="carriageM2" label="M2车" min-width="25%"></el-table-column>
                <!-- <el-table-column prop="carriageM3" label="M3车" min-width="25%"></el-table-column> -->
                <!-- <el-table-column prop="carriageM4" label="M4车" min-width="25%"></el-table-column> -->
                <el-table-column prop="carriageMP2" label="MP2车" min-width="25%"></el-table-column>
                <el-table-column prop="carriageTC2" label="TC2车" min-width="25%"></el-table-column>
                <template #empty>
                    <img src="/img/no-data.svg" width="40" />
                    <p style="font-size:12px">当前暂无数据</p>
                </template>
            </el-table>
        </div>
    </div>
</template>

<script setup>
const props = defineProps({
  CarTemperatureData: {
    type: Array,
    default: []
  }
})
const CarTemperatureData = [
    {
        name: '车厢内温度',
        carriageTC1: '26℃',
        carriageMP1: '27℃',
        carriageM1: '30℃',
        carriageM2: '28℃',
        carriageMP2: '29℃',
        carriageTC2: '26℃'
    }, {
        name: '环境温度',
        carriageTC1: '28℃',
        carriageMP1: '28℃',
        carriageM1: '28℃',
        carriageM2: '28℃',
        carriageMP2: '28℃',
        carriageTC2: '28℃'
    }, {
        name: '目标温度',
        carriageTC1: '26℃',
        carriageMP1: '27℃',
        carriageM1: '30℃',
        carriageM2: '28℃',
        carriageMP2: '29℃',
        carriageTC2: '26℃'
    }
]

</script>

<style scoped lang="scss">
.el-table,
.el-table tr,
.el-table td,
.el-table th {
    background-color: transparent !important;
}

.mx-1 {
    background-color: rgba(0, 0, 0, 0);
    border-radius: 12px;
}

.el-tag--default {
    height: 20px;
}

// 操作文字
.btn_text {
    font-size: 14px;
    color: #2186CF !important;
}

.warp {
    width: 100%;
    border-radius: 10px;
    background-color: #171c32 !important;
    margin-bottom: 0!important
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
    background: #fff;
    margin-left: -10px;
}
</style>
