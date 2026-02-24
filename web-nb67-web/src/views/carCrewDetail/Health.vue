<template>
  <div class="warp">
    <!-- <div class="title">车厢温度信息</div> -->
    <div style="padding: 0 15px">
      <el-table :data="props.tableData" stripe style="width: 100%">
        <!-- <el-table-column prop="name" width="180"></el-table-column> -->
        <el-table-column prop="name" label="设备名称">
          <template #default="scope">
            <span v-if="scope.row.TC1 == 'NaN℃'">N/A</span>
            <span v-else>{{ scope.row.TC1 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="health" label="健康状态">
          <template #default="scope">
            <span v-if="scope.row.MP1 == 'NaN℃'">N/A</span>
            <span v-else>{{ scope.row.MP1 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="usetime" label="累计运行时间">
          <template #default="scope">
            <span v-if="scope.row.M1 == 'NaN℃'">N/A</span>
            <span v-else>{{ scope.row.M1 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="actionnum" label="动作次数">
          <template #default="scope">
            <span v-if="scope.row.M2 == 'NaN℃'">N/A</span>
            <span v-else>{{ scope.row.M2 }}</span>
          </template>
        </el-table-column>
        <template #empty>
          <img src="/img/no-data.svg" width="80" height="50" />
          <p style="font-size: 12px">当前暂无数据</p>
        </template>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";

const props = defineProps({
  tableData: {
    type: Array,
    default: [],
  },
});

onMounted(() => {
});
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
  border: 1px solid #3a3b68;
  margin-bottom: 15px;
  height: 300px;
}
:deep(.el-scrollbar__view) {
  height: 240px;
}
</style>
