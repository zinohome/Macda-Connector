<template>
  <div class="warp">
    <div style="padding: 0 15px">
      <el-table :data="tableData" stripe style="width: 100%">
        <el-table-column
          prop="name"
          label="设备名称"
          min-width="100%"
        ></el-table-column>
        <el-table-column prop="equ1_1" label="1号机组">
          <template #default="scope">
            <span v-if="scope.row.equ1_1" style="color: #67c32a">{{
              stateName(scope.row.name, scope.row.equ1_1)
            }}</span>
            <span v-else style="color: #8e8586">
                <span v-if="scope.row.name=='新风阀开度' || scope.row.name=='回风阀开度' ">0%</span>
                <span v-else>停机</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="equ1_2" label="2号机组">
          <template #default="scope">
            <span v-if="scope.row.equ1_2" style="color: #67c32a">{{
              stateName(scope.row.name, scope.row.equ1_2)
            }}</span>
             <span v-else style="color: #8e8586">
                <span v-if="scope.row.name=='新风阀开度' || scope.row.name=='回风阀开度' ">0%</span>
                <span v-else>停机</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="equ2_1" label="3号机组">
          <template #default="scope">
            <span v-if="scope.row.equ2_1" style="color: #67c32a">{{
              stateName(scope.row.name, scope.row.equ2_1)
            }}</span>
             <span v-else style="color: #8e8586">
                <span v-if="scope.row.name=='新风阀开度' || scope.row.name=='回风阀开度' ">0%</span>
                <span v-else>停机</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="equ2_2" label="4号机组">
          <template #default="scope">
            <span v-if="scope.row.equ2_2" style="color: #67c32a">{{
              stateName(scope.row.name, scope.row.equ2_2)
            }}</span>
             <span v-else style="color: #8e8586">
                <span v-if="scope.row.name=='新风阀开度' || scope.row.name=='回风阀开度' ">0%</span>
                <span v-else>停机</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="equ3_1" label="5号机组">
          <template #default="scope">
            <span v-if="scope.row.equ3_1" style="color: #67c32a">{{
              stateName(scope.row.name, scope.row.equ3_1)
            }}</span>
             <span v-else style="color: #8e8586">
                <span v-if="scope.row.name=='新风阀开度' || scope.row.name=='回风阀开度' ">0%</span>
                <span v-else>停机</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="equ3_2" label="6号机组">
          <template #default="scope">
            <span v-if="scope.row.equ3_2" style="color: #67c32a">{{
              stateName(scope.row.name, scope.row.equ3_2)
            }}</span>
             <span v-else style="color: #8e8586">
                <span v-if="scope.row.name=='新风阀开度' || scope.row.name=='回风阀开度' ">0%</span>
                <span v-else>停机</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="equ4_1" label="7号机组">
          <template #default="scope">
            <span v-if="scope.row.equ4_1" style="color: #67c32a">{{
              stateName(scope.row.name, scope.row.equ4_1)
            }}</span>
             <span v-else style="color: #8e8586">
                <span v-if="scope.row.name=='新风阀开度' || scope.row.name=='回风阀开度' ">0%</span>
                <span v-else>停机</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="equ4_2" label="8号机组">
          <template #default="scope">
            <span v-if="scope.row.equ4_2" style="color: #67c32a">{{
              stateName(scope.row.name, scope.row.equ4_2)
            }}</span>
            <span v-else style="color: #8e8586">
                <span v-if="scope.row.name=='新风阀开度' || scope.row.name=='回风阀开度' ">0%</span>
                <span v-else>停机</span>
            </span>
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

const stateName = (name, num) => {
  switch (name) {
    case "新风阀开度":
    case "回风阀开度":
      return num + "%";
    default:
      return "运行";
  }
};
const tableData = [
  {
    name: '新风机A',
  },
  {
    name: '新风机B',
  },
  {
    name: '新风机C',
  },
  {
    name: '新风机D',
  },
  {
    name: '新风机E',
  },
  {
    name: '新风机F',
  },
  {
    name: '新风机G',
  },
]

onMounted(() => {});
</script>

<style scoped>
.el-table,
.el-table tr,
.el-table td,
.el-table th {
  background-color: transparent !important;
}
.warp{
  background-color: #181F30 !important;
  border: 1px solid #3a3b68;
  height: 300px;
}
:deep(.el-scrollbar__wrap--hidden-default) {
  height: 240px;
}
</style>
