<template>
  <div class="warp2">
    <div class="header">
      <BackTop />
      <div class="trainInfo">
        当前机组：<Dropdown @handleCommand="handleCommand" />
      </div>
    </div>

    <div class="container">
      <Healthy :tableData="healthyData" />
      <el-row :gutter="15" class="mb-15">
        <!-- <el-col :span="8">
            <System title="温度参数" :tableData="system1" nameLeft="系统名称"/>
        </el-col>
        <el-col :span="8">
            <System title="系统1参数" :tableData="system2" nameLeft="设备名称"/>
        </el-col>
        <el-col :span="8">
            <System title="系统2参数" :tableData="system3" nameLeft="设备名称"/>
        </el-col> -->
        <el-col :span="24">
          <TrainUnitMap :crewDetails="crewDetails" :UnitNo="UnitNo" />
        </el-col>
      </el-row>
      <Echart :carID="carID" ref="EchartChild"/>
    </div>
  </div>
</template>

<script setup>
import { getHealthAssessment, gettemperature } from '@/api/api.js'

import BackTop from "@/components/BackTop.vue"
import Healthy from "./Healthy.vue";
import System from "./System.vue";
import Dropdown from './Dropdown.vue'
import Echart from "./Echart.vue";
import TrainUnitMap from "./TrainUnitMap.vue";
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

let route = useRoute();
let router = useRouter();

let carriageID = '1'    //列车号
let trainNum = '1'   //车厢号
let trainCrew = '1'  //机组
let carID = ref('1')
let healthyData = ref([])  //健康评估信息
let system1 = ref([])
let system2 = ref([])
let system3 = ref([])
let UnitNo = ref('1号机组')
const EchartChild = ref(null);
const crewDetails = ref({})
const Unit1 = [
  {
    title:'通风机',
    value: '更换风机轴承'
  },
  {
    title:'冷凝风机',
    value: '更换风机轴承'
  },
  {
    title:'压缩机1',
    value: '更换压缩机'
  },
  {
    title:'压缩机2',
    value: '更换压缩机'
  },
  {
    title:'新风阀',
    value: '更换风阀执行器'
  },
  {
    title:'回风阀',
    value: '更换风阀执行器'
  },
]
const Unit2 = [
  {
    title:'通风机',
    value: '更换风机轴承'
  },
  {
    title:'冷凝风机',
    value: '更换风机轴承'
  },
  {
    title:'压缩机1',
    value: '更换压缩机'
  },
  {
    title:'压缩机2',
    value: '更换压缩机'
  },
  {
    title:'新风阀',
    value: '更换风阀执行器'
  },
  {
    title:'回风阀',
    value: '更换风阀执行器'
  },
]
const setSystem1 = (obj) => {
  let arr = Object.keys(obj);
  if (arr.length == 0) {
    return [
      { name: '新风温度', value: `N/A` },
      { name: '回风温度', value: `N/A` },
      { name: '送风温度1', value: `N/A` },
      { name: '送风温度2', value: `N/A` }
      //   { name: '车厢温度1', value: `N/A` },
      //   { name: '车厢温度2', value: `N/A` }
    ]
  }
  // return [
  //     { name: '新风温度', value: `${(obj[0].i_outer_temp*0.1).toFixed(2)}℃` },
  //     { name: '回风温度', value: `${UnitNo.value == '1号机组'?(obj[0].i_rat_u1*0.1).toFixed(2): (obj[0].i_rat_u2*0.1).toFixed(2)}℃` },
  //     { name: '送风温度1', value: `${UnitNo.value == '1号机组'?(obj[0].i_sat_u11*0.1).toFixed(2):(obj[0].i_sat_u21*0.1).toFixed(2)}℃` },
  //     { name: '送风温度2', value: `${UnitNo.value == '1号机组'?(obj[0].i_sat_u12*0.1).toFixed(2):(obj[0].i_sat_u22*0.1).toFixed(2)}℃` }
  //     // { name: '车厢温度1', value: `${obj['last_iSeatTemp'] / 10}℃` },
  //     // { name: '车厢温度2', value: `${obj['last_iVehTemp'] / 10}℃` }
  // ]
  return [
    { name: '新风温度', value: `${obj[0].i_outer_temp}℃` },
    { name: '回风温度', value: `${UnitNo.value == '1号机组'?obj[0].i_rat_u1: obj[0].i_rat_u2}℃` },
    { name: '送风温度1', value: `${UnitNo.value == '1号机组'?obj[0].i_sat_u11:obj[0].i_sat_u21}℃` },
    { name: '送风温度2', value: `${UnitNo.value == '1号机组'?obj[0].i_sat_u12:obj[0].i_sat_u22}℃` }
    // { name: '车厢温度1', value: `${obj['last_iSeatTemp'] / 10}℃` },
    // { name: '车厢温度2', value: `${obj['last_iVehTemp'] / 10}℃` }
  ]
}

const setSystem2 = (obj, num) => {
  let arr = Object.keys(obj);
  if (arr.length == 0) {
    return [
      { name: '变频器频率', value: `N/A` },
      { name: '变频器电流', value: `N/A` },
      { name: '吸气温度', value: `N/A` },
      { name: '吸气压力', value: `N/A` },
      { name: '过热度', value: `N/A` },
      //   { name: '膨胀阀开度', value: `N/A` }
    ]
  }
  // return [
  //     { name: '变频器频率', value: `${UnitNo.value == '1号机组'?(obj[0].w_crnt_u11*0.01).toFixed(2):(obj[0].w_crnt_u21*0.01).toFixed(2)}Hz` },
  //     { name: '变频器电流', value: `${UnitNo.value == '1号机组'?(obj[0].w_freq_u11*0.1).toFixed(2):(obj[0].w_freq_u21*0.1).toFixed(2)}A` },
  //     { name: '吸气温度', value: `${UnitNo.value == '1号机组'?(obj[0].i_suck_temp_u11*0.1).toFixed(2):(obj[0].i_suck_temp_u21*0.1).toFixed(2)}℃` },
  //     { name: '吸气压力', value: `${UnitNo.value == '1号机组'?(obj[0].i_suck_temp_u11*0.1).toFixed(2):(obj[0].i_suck_temp_u21*0.1).toFixed(2)}bar` },
  //     { name: '过热度', value: `${UnitNo.value == '1号机组'?(obj[0].i_sup_heat_u11*0.1).toFixed(2):(obj[0].i_sup_heat_u21*0.1).toFixed(2)}k` },
  //     // { name: '膨胀阀开度', value: `${obj.frequency}%` }
  // ]
  return [
    { name: '变频器频率', value: `${UnitNo.value == '1号机组'?(obj[0].w_freq_u11).toFixed(2):(obj[0].w_freq_u21).toFixed(2)}Hz` },
    { name: '变频器电流', value: `${UnitNo.value == '1号机组'?(obj[0].w_crnt_u11).toFixed(2):(obj[0].w_crnt_u21).toFixed(2)}A` },
    { name: '吸气温度', value: `${UnitNo.value == '1号机组'?obj[0].i_suck_temp_u11:obj[0].i_suck_temp_u21}℃` },
    { name: '吸气压力', value: `${UnitNo.value == '1号机组'?(obj[0].i_suck_pres_u11).toFixed(2):(obj[0].i_suck_pres_u21).toFixed(2)}bar` },
    { name: '过热度', value: `${UnitNo.value == '1号机组'?(obj[0].i_sup_heat_u11).toFixed(2):(obj[0].i_sup_heat_u21).toFixed(2)}k` },
    // { name: '膨胀阀开度', value: `${obj.frequency}%` }
  ]
}
const setSystem3 = (obj, num) => {
  let arr = Object.keys(obj);
  if (arr.length == 0) {
    return [
      { name: '变频器频率', value: `N/A` },
      { name: '变频器电流', value: `N/A` },
      { name: '吸气温度', value: `N/A` },
      { name: '吸气压力', value: `N/A` },
      { name: '过热度', value: `N/A` },
      //   { name: '膨胀阀开度', value: `N/A` }
    ]
  }
  // return [
  //     { name: '变频器频率', value: `${UnitNo.value == '1号机组'?(obj[0].w_crnt_u12*0.01).toFixed(2):(obj[0].w_crnt_u22*0.01).toFixed(2)}Hz` },
  //     { name: '变频器电流', value: `${UnitNo.value == '1号机组'?(obj[0].w_freq_u12*0.1).toFixed(2):(obj[0].w_freq_u22*0.1).toFixed(2)}A` },
  //     { name: '吸气温度', value: `${UnitNo.value == '1号机组'?(obj[0].i_suck_temp_u12*0.1).toFixed(2):(obj[0].i_suck_temp_u22*0.1).toFixed(2)}℃` },
  //     { name: '吸气压力', value: `${UnitNo.value == '1号机组'?(obj[0].i_suck_temp_u21*0.1).toFixed(2):(obj[0].i_suck_temp_u22*0.1).toFixed(2)}bar` },
  //     { name: '过热度', value: `${UnitNo.value == '1号机组'?(obj[0].i_sup_heat_u12*0.1).toFixed(2):(obj[0].i_sup_heat_u22*0.1).toFixed(2)}k` },
  //     // { name: '膨胀阀开度', value: `${obj.frequency}%` }
  // ]
  return [
    { name: '变频器频率', value: `${UnitNo.value == '1号机组'?(obj[0].w_freq_u12).toFixed(2):(obj[0].w_freq_u22).toFixed(2)}Hz` },
    { name: '变频器电流', value: `${UnitNo.value == '1号机组'?(obj[0].w_crnt_u12).toFixed(2):(obj[0].w_crnt_u22).toFixed(2)}A` },
    { name: '吸气温度', value: `${UnitNo.value == '1号机组'?obj[0].i_suck_temp_u12:obj[0].i_suck_temp_u22}℃` },
    { name: '吸气压力', value: `${UnitNo.value == '1号机组'?(obj[0].i_suck_pres_u21).toFixed(2):(obj[0].i_suck_pres_u22).toFixed(2)}bar` },
    { name: '过热度', value: `${UnitNo.value == '1号机组'?(obj[0].i_sup_heat_u12).toFixed(2):(obj[0].i_sup_heat_u22).toFixed(2)}k` },
    // { name: '膨胀阀开度', value: `${obj.frequency}%` }
  ]
}


const getData = (id) => {
  getHealthAssessment(id).then(res => {
    // healthyData.value = res.vw_health_assessment
    let data1 = {}
    let data2 = {}
    let data3 = {}
    let data4 = {}
    let data5 = {}
    let data6 = {}
    let data7 = {}
    let data8 = {}
    const ventilator70 = 25000*0.7*3600
    const ventilator85 = 25000*0.85*3600
    const million70 = 1000000*0.7
    const million85 = 1000000*0.85
    const engine70 = 50000*0.7*3600
    const engine85 = 50000*0.85*3600
    // console.log(ventilator75);
    if(UnitNo.value == '1号机组'){
      healthyData.value = []
      res.vw_health_assessment.forEach(item => {
        data1.equipment_name = '通风机'
        data1.cumulative_running_time = item.dwef_op_tm_u11
        data1.cumulative_running_time = (data1.cumulative_running_time/3600).toFixed(2)
        if(item.dwef_op_tm_u11 <ventilator70){
          data1.health_status = '健康'
        } else if(ventilator70<item.dwef_op_tm_u11<ventilator85){
          data1.health_status = '亚健康'
          data1.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data1.health_status = '非健康'
          data1.precautions = '更换风机轴承'
        }
        data1.number_actions = item.dwef_op_cnt_u11
        data2.equipment_name = '冷凝机'
        data2.cumulative_running_time = item.dwcf_op_tm_u11
        data2.cumulative_running_time = (data2.cumulative_running_time/3600).toFixed(2)
        if(item.dwcf_op_tm_u11 <ventilator70){
          data2.health_status = '健康'
        } else if(ventilator70<item.dwcf_op_tm_u11<ventilator85){
          data2.health_status = '亚健康'
          data2.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data2.health_status = '非健康'
          data2.precautions = '更换风机轴承'
        }
        data2.number_actions = item.dwcf_op_cnt_u11
        data3.equipment_name = '新风阀'
        data3.cumulative_running_time = '-'
        data3.number_actions = item.dwfad_op_cnt_u1
        if(item.dwfad_op_cnt_u1 <million70){
          data3.health_status = '健康'
        } else if(million70<item.dwfad_op_cnt_u1<million85){
          data3.health_status = '亚健康'
          data3.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data3.health_status = '非健康'
          data3.precautions = '更换风阀执行器'
        }
        data4.equipment_name = '回风阀'
        data4.cumulative_running_time = '-'
        data4.number_actions = item.dwrad_op_cnt_u1
        if(item.dwrad_op_cnt_u1 <million70){
          data4.health_status = '健康'
        } else if(million70<item.dwrad_op_cnt_u1<million85){
          data4.health_status = '亚健康'
          data4.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data4.health_status = '非健康'
          data4.precautions = '更换风阀执行器'
        }
        data5.equipment_name = '压缩机1'
        data5.cumulative_running_time = item.dwcp_op_tm_u11
        data5.cumulative_running_time = (data5.cumulative_running_time/3600).toFixed(2)
        data5.number_actions = item.dwcp_op_cnt_u11
        if(item.dwcp_op_tm_u11 <engine70){
          data5.health_status = '健康'
        } else if(engine70<item.dwcp_op_tm_u11<engine85){
          data5.health_status = '亚健康'
          data5.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data5.health_status = '非健康'
          data5.precautions = '更换压缩机'
        }
        data6.equipment_name = '压缩机2'
        data6.cumulative_running_time = item.dwcp_op_tm_u12
        data6.cumulative_running_time = (data6.cumulative_running_time/3600).toFixed(2)
        data6.number_actions = item.dwcp_op_cnt_u12
        if(item.dwcp_op_tm_u12 <engine70){
          data6.health_status = '健康'
        } else if(engine70<item.dwcp_op_tm_u12<engine85){
          data6.health_status = '亚健康'
          data6.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data6.health_status = '非健康'
          data6.precautions = '更换压缩机'
        }

        data7.equipment_name = '废排风机'
        data7.cumulative_running_time = item.dwexufan_op_tm
        data7.cumulative_running_time = (data7.cumulative_running_time/3600).toFixed(2)
        if(item.dwexufan_op_tm <ventilator70){
          data7.health_status = '健康'
        } else if(ventilator70<item.dwexufan_op_tm<ventilator85){
          data7.health_status = '亚健康'
          data7.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data7.health_status = '非健康'
          data7.precautions = '更换风机轴承'
        }
        data7.number_actions = item.dwexufan_op_cnt

        data8.equipment_name = '废排风阀'
        data8.cumulative_running_time = '-'
        data8.number_actions = item.dwdmpexu_op_cnt
        if(item.dwdmpexu_op_cnt <million70){
          data8.health_status = '健康'
        } else if(million70<item.dwdmpexu_op_cnt<million85){
          data8.health_status = '亚健康'
          data8.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data8.health_status = '非健康'
          data8.precautions = '更换风阀执行器'
        }
      })
    } else {
      healthyData.value = []
      res.vw_health_assessment.forEach(item => {
        data1.equipment_name = '通风机'
        data1.cumulative_running_time = item.dwef_op_tm_u21
        data1.cumulative_running_time = (data1.cumulative_running_time/3600).toFixed(2)
        if(item.dwef_op_tm_u21 <ventilator70){
          data1.health_status = '健康'
        } else if(ventilator70<item.dwef_op_tm_u21<ventilator85){
          data1.health_status = '亚健康'
          data1.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data1.health_status = '非健康'
          data1.precautions = '更换风机轴承'
        }
        data1.number_actions = item.dwef_op_cnt_u21
        data2.equipment_name = '冷凝机'
        data2.cumulative_running_time = item.dwcf_op_tm_u21
        data2.cumulative_running_time = (data2.cumulative_running_time/3600).toFixed(2)
        if(item.dwcf_op_tm_u21 <ventilator70){
          data2.health_status = '健康'
        } else if(ventilator70<item.dwcf_op_tm_u21<ventilator85){
          data2.health_status = '亚健康'
          data2.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data2.health_status = '非健康'
          data2.precautions = '更换风机轴承'
        }
        data2.number_actions = item.dwcf_op_cnt_u11
        data3.equipment_name = '新风阀'
        data3.cumulative_running_time = '-'
        data3.number_actions = item.dwfad_op_cnt_u2
        if(item.dwfad_op_cnt_u2 <million70){
          data3.health_status = '健康'
        } else if(million70<item.dwfad_op_cnt_u2<million85){
          data3.health_status = '亚健康'
          data3.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data3.health_status = '非健康'
          data3.precautions = '更换风阀执行器'
        }
        data4.equipment_name = '回风阀'
        data4.cumulative_running_time = '-'
        data4.number_actions = item.dwrad_op_cnt_u2
        if(item.dwrad_op_cnt_u2 <million70){
          data4.health_status = '健康'
        } else if(million70<item.dwrad_op_cnt_u2<million85){
          data4.health_status = '亚健康'
          data4.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data4.health_status = '非健康'
          data4.precautions = '更换风阀执行器'
        }
        data5.equipment_name = '压缩机1'
        data5.cumulative_running_time = item.dwcp_op_tm_u21
        data5.cumulative_running_time = (data5.cumulative_running_time/3600).toFixed(2)
        data5.number_actions = item.dwcp_op_cnt_u21
        if(item.dwcp_op_tm_u21 <engine70){
          data5.health_status = '健康'
        } else if(engine70<item.dwcp_op_tm_u21<engine85){
          data5.health_status = '亚健康'
          data5.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data5.health_status = '非健康'
          data5.precautions = '更换压缩机'
        }
        data6.equipment_name = '压缩机2'
        data6.cumulative_running_time = item.dwcp_op_tm_u22
        data6.cumulative_running_time = (data6.cumulative_running_time/3600).toFixed(2)
        data6.number_actions = item.dwcp_op_cnt_u22
        if(item.dwcp_op_tm_u22 <engine70){
          data6.health_status = '健康'
        } else if(engine70<item.dwcp_op_tm_u22<engine85){
          data6.health_status = '亚健康'
          data6.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data6.health_status = '非健康'
          data6.precautions = '更换压缩机'
        }
        data7.equipment_name = '废排风机'
        data7.cumulative_running_time = item.dwexufan_op_tm
        data7.cumulative_running_time = (data7.cumulative_running_time/3600).toFixed(2)
        if(item.dwexufan_op_tm <ventilator70){
          data7.health_status = '健康'
        } else if(ventilator70<item.dwexufan_op_tm<ventilator85){
          data7.health_status = '亚健康'
          data7.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data7.health_status = '非健康'
          data7.precautions = '更换风机轴承'
        }
        data7.number_actions = item.dwexufan_op_cnt

        data8.equipment_name = '废排风阀'
        data8.cumulative_running_time = '-'
        data8.number_actions = item.dwdmpexu_op_cnt
        if(item.dwdmpexu_op_cnt <million70){
          data8.health_status = '健康'
        } else if(million70<item.dwdmpexu_op_cnt<million85){
          data8.health_status = '亚健康'
          data8.precautions = '设备处于亚健康状态，即将进入检修期'
        } else {
          data8.health_status = '非健康'
          data8.precautions = '更换风阀执行器'
        }
      })
    }
    // console.log(data1);
    healthyData.value.push(data1,data2,data3,data4,data5,data6,data7,data8)
    // console.log(healthyData.value);
  })

  gettemperature(id).then(res => {
    console.log(res);
    // system1.value = setSystem1(res.vw_system_info)
    // system2.value = setSystem2(res.vw_system_info)
    // system3.value = setSystem3(res.vw_system_info)
    crewDetails.value = res.vw_system_info[0]
    console.log(crewDetails.value);
  })

}

const handleCommand = (index, value) => {
  console.log(index,value)
  value.forEach((item)=>{
    if(item.name == index){
      carID.value = item.trainNo
      console.log(item);
      router.push({
          name: 'airConditioner',
          query: {
          trainNo:route.query.trainNo,
          trainCoach:item.trainNo
          }
      })
      // if(item.trainNo == undefined){
      //     alert('222222')
      // } else {
      //     carID.value = item.trainNo
      // }
    }
  })
  UnitNo.value = index.split(" ")[1]
  console.log(UnitNo.value);
  // trainNum = Math.ceil(index / 2)       // 车厢号
  // trainCrew = (index % 2) ? 1 : 2       // 机组号

  // carID.value = route.query.carriageID

  // router.push({
  //     name: 'airConditioner',
  //     params: {
  //         trainNum: trainNum,
  //         trainCrew: trainCrew
  //     },
  //     replace: true
  // })
  console.log(carID.value);
  getData(carID.value)
}
let time = setInterval(() => {
  carID.value = route.params.trainCoach || route.query.trainCoach
  getData(carID.value)
  console.log('请求')
}, 1000*60*2);
onMounted(() => {
  // let carriageID = '1'    //列车号
  // let trainNum = '1'   //车厢号
  // let trainCrew = '1'  //机组
  carriageID = route.params.trainCoach || route.query.trainCoach
  trainNum = route.params.trainNo || route.query.trainNo
  trainCrew = route.params.trainCrew || route.query.trainCrew
  //alert( carriageID)
  if(!trainCrew){
    trainCrew = 1
  }
  console.log( carriageID, trainNum, trainCrew)
  // carID.value = carriageID + '-' + trainNum
  carID.value = carriageID
  getData(carID.value)
  EchartChild.value.initEchart(carID.value,'res.hourly')
})
onUnmounted(()=>{
  clearInterval(time)
})

</script>

<style scoped lang="scss">
$s-primary: #2186CF;
$s-danger: #e65355;
$s-warning: #ffa55c;
$s-info: #bfbfbf;
$s-success: #42ad5d;
$s-yellow: #ffe375;
$s-white: #ffffff;
$s-card-bg: #181F30;
$s-line: #394153;

:deep(table) {
  border-color: transparent!important;
}
:deep(thead) {
  background: transparent;
}
:deep(.title) {
  width: 100%;
  height: 40px;
  line-height: 40px;
  padding-left: 20px;
  box-sizing: border-box;
  /* border-bottom: 2px solid #1f273c; */
  margin-bottom: 10px;
  font-size: 13px;
  font-weight: bold;
}
/* nodate table style */
:deep(.el-table__empty-text) {
  margin: 30px 0;
  line-height: 20px !important;
}
:deep(.el-table th) {
  border-color: transparent!important;
}
:deep(.trainInfo .el-input__inner, .trainInfo .el-input__inner .el-input__inner:hover, .trainInfo .el-input__inner .el-input__inner:focus) {
  background-color: $s-card-bg!important;
  box-shadow: 0 0 1px 0 $s-line;
  color: #fff !important;
  height: 40px;
  box-sizing: border-box;
  font-size: 110%;
}
:deep(.el-table .el-table__cell) {
  padding: 10px !important;
}

:deep(.el-input__wrapper) {
  background-color: $s-card-bg !important;
  border: 1px solid $s-line;
  box-shadow: 0 0 1px 0 $s-line;
  color: #fff !important;

  .el-input__inner {
    border: none !important;
    color: #fff !important;
  }
}
.header {
  height: 60px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 5px;
}
.ul {
  width: 100%;
  box-sizing: border-box;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 15px;
}
/* .ul > li {
    width: 32.5%;
} */
.warp {
  box-sizing: border-box;
  width: 100%;
  border-radius: 10px;
  background-color: $s-card-bg !important;
  border: 1px solid $s-line;
  margin-bottom: 15px;
}
.btn_text {
  font-size: 12px;
  color: #2186CF;
}
.el-table tr th {
  color: #fff !important;
  font-size: 13px;
}
.title {
  width: 100%;
  height: 40px;
  line-height: 40px;
  padding-left: 20px;
  box-sizing: border-box;
  border-bottom: 2px solid #1f273c;
  margin-bottom: 10px;
}

.warp2 {
  background: $s-card-bg;
  border-radius: 20px;
  border: 1px solid #3a3b68;
  padding: 5px 20px 20px;
}
</style>


