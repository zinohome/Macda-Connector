<template>
    <!-- <el-dropdown trigger="click" @command="handleCommand">
        <div class="drop-title">
            <img src="/img/snow.png" width="30" height="30" />
            {{ train }}
            <el-icon>
                <arrow-down />
            </el-icon>
        </div>
        <template #dropdown>
            <el-dropdown-menu>
                <el-dropdown-item
                    v-for="(item, index) in trains"
                    :key="index"
                    :command="index"
                >{{ item }}</el-dropdown-item>
            </el-dropdown-menu>
        </template>
    </el-dropdown> -->
    <el-select v-model="value" placeholder="请选择">
        <el-option v-for="(item, index) in trains" :key="index" :command="index">
        </el-option>
    </el-select>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from "vue-router";
import { ArrowDown } from '@element-plus/icons-vue'

let route = useRoute();
let trainNo = route.params.trainNo || route.query.trainNo
const train = ref('TC1车 1号机组')
const trains = ref(['TC1车 1号机组', 'TC1车 2号机组', 'MP1车 1号机组', 'MP1车 2号机组',
 'M1车 1号机组', 'M1车 2号机组', 'M2车 1号机组', 'M2车 2号机组', 'MP2车 1号机组', 
 'MP2车 2号机组', 'TC2车 1号机组', 'TC2车 2号机组'])
const emit = defineEmits(["handleCommand"])


const handleCommand = (index) => {
    train.value = trains.value[index]

    emit('handleCommand', index + 1)
}

onMounted(() => {
  trainNo = route.params.trainNo || route.query.trainNo
  let trainCrew = route.params.trainCrew || route.query.trainCrew
  if(!trainCrew){
    trainCrew = 1
  }
  console.log(trainNo,trainCrew);
  let index = ( trainNo * 2) - (trainCrew % 2) - 1
  train.value = trains.value[index]
})
</script>
<style>
.drop-title {
    padding: 0 10px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 165px;
    height: 50px;
    line-height: 50px;
    font-weight: 700;
    background: #273452;
    color: #fff;
}
</style>


