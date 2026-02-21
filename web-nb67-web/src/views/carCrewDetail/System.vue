<template>
  <div class="warp">
    <!-- <div class="title">车厢温度信息</div> -->
    <div style="padding: 0 15px">
      <el-table :data="props.tableData" stripe style="width: 100%">
        <!-- <el-table-column prop="name" width="180"></el-table-column> -->
        <el-table-column prop="name" label="系统名称"></el-table-column>
        <el-table-column align="right" prop="actionnum" label="系统1参数">
          <template #default="scope">
            <span v-if="scope.row.M2 == 'NaN℃'">N/A</span>
            <span v-else>{{ scope.row.M2 }}</span>
          </template>
        </el-table-column>
        <template #empty>
          <img src="/src/assets/img/no-data.svg" width="80" height="50" />
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
