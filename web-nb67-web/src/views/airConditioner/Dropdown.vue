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
    <el-select v-model="valueName" placeholder="请选择" @change="changeSelect">
        <el-option v-for="(item, index) in trains" :key="index" :command="index" :value="item.name">
        </el-option>
    </el-select>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute,useRouter } from "vue-router"
import { ArrowDown } from '@element-plus/icons-vue'
import { getTrainSelection } from '@/api/api'
let route = useRoute();
let trainNo = route.query.trainNo;
const valueName = ref('')
const train = ref('TC1车 1号机组')
const carriage_name = ref(['TC1','MP1','M1','M2','MP2','TC2'])
// const trains = ref(['TC1车 1号机组', 'TC1车 2号机组', 'MP1车 1号机组', 'MP1车 2号机组',
//  'M1车 1号机组', 'M1车 2号机组', 'M2车 1号机组', 'M2车 2号机组', 'MP2车 1号机组', 
//  'MP2车 2号机组', 'TC2车 1号机组', 'TC2车 2号机组'])
const trains = ref([
    {
    name: 'TC1车 1号机组',
    value: 'TC1',
    index: '01',
    },{
    name: 'TC1车 2号机组',
    value: 'TC1',
    index: '01',
    },{
    name: 'MP1车 1号机组',
    value: 'MP1',
    index: '02',
    },{
    name: 'MP1车 2号机组',
    value: 'MP1',
    index: '02',
    },{
    name: 'M1车 1号机组',
    value: 'M1',
    index: '03',
    },{
    name: 'M1车 2号机组',
    value: 'M1',
    index: '03',
    },{
    name: 'M2车 1号机组',
    value: 'M2',
    index: '04',
    },{
    name: 'M2车 2号机组',
    value: 'M2',
    index: '04',
    },{
    name: 'MP2车 1号机组',
    value: 'MP2',
    index: '05',
    },{
    name: 'MP2车 2号机组',
    value: 'MP2',
    index: '05',
    },{
    name: 'TC2车 1号机组',
    value: 'TC2',
    index: '06',
    },{
    name: 'TC2车 2号机组',
    value: 'TC2',
    index: '06',
    },
])
// const trains = ref(['1机组', '2机组'])
const emit = defineEmits(["handleCommand"])
console.log(trainNo);
getTrainSelection(trainNo.slice(0,5)).then((res)=>{
    console.log('执行');
    const data = []
    res.vw_carriage_status.forEach((element,indexs) => {
        // console.log(res.vw_carriage_status[indexs],carriage_name.value[indexs]);
        // carriage_name.value.forEach((item,index)=>{
            console.log(element.carriage_no.slice(-1),res.vw_carriage_status,carriage_name.value,indexs);
            let num = element.carriage_no.slice(-1) -1
            console.log(num);
            data.push({
                carriage_no: res.vw_carriage_status[indexs].carriage_no,
                carriage_name: carriage_name.value[num]
            })
            
        // })
    });
    console.log(data, "getTrainSelection.data");
    data.forEach((value)=>{
        trains.value.forEach((item)=>{
            if(value.carriage_name == item.value){
                item.trainNo = value.carriage_no
            }
        })
    })
    console.log(trains.value);
    const current = []
    trains.value.forEach((item)=>{
    if(item.trainNo == route.query.trainCoach){
        console.log(item, '9999999');
        current.push(item.name)
        console.log(current);
        valueName.value = current[0]
        console.log(valueName.value);
    }
})
})
// trains.value.forEach((item)=>{
    
//     if(item.trainNo == route.query.trainCoach){
//         valueName.value = item.name
//     }
// })
const handleCommand = (index) => {
    train.value = trains.value[index]

    emit('handleCommand', index + 1)
}
const changeSelect = (value) => {
console.log(value)
emit('handleCommand', value, trains.value)
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


