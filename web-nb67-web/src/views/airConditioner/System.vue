<template>
    <div class="warp">
        <el-table
            :data="props.tableData"
            stripe
            height="290px"
            style="width: 100%"
        >
            <el-table-column prop="name" :label="props.nameLeft" min-width="50%"></el-table-column>
            <el-table-column prop="value" :label="props.title" align="right" min-width="50%"></el-table-column>
            <template #empty>
                <img src="/src/assets/img/no-data.svg" width="40" height="50" />
                <p style="font-size:12px">当前暂无数据</p>
            </template>
        </el-table>
    </div>
</template>

<script setup>
// import { ref, onMounted } from 'vue'

const props = defineProps({
    title: {
        type: String,
        default: () => 'title'
    },
    nameLeft: {
        type: String,
        default: () => 'nameLeft'
    },
    tableData: {
        type: Array,
    }
})
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
    margin-bottom: 15px;
}
.el-table {
    --el-table-border-color: none;
}
</style>
