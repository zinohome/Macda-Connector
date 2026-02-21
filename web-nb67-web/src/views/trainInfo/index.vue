<template>
    <div class="warp2">
        <!-- 顶部 -->
        <div class="header">
            <!-- <BackTop /> -->
            <div></div>
            <div class="trainInfo">
                当前列车号：{{ route.query.trainNo }} &nbsp;
                <!-- <Dropdown @handleCommand="handleCommand" /> -->
            </div>
        </div>
        <el-row :gutter="15" class="container">
            <el-col :span="12">
                <ActualWarning :ActualWarningData="ActualWarningData"/>
            </el-col>
            <el-col :span="12">
                <StateWarning :StateWarningData="StateWarningData"/>
            </el-col>
        </el-row>
        <!-- 车厢选择 -->
        <CarCrew :trainId="route.query.trainNo"/>
        <!-- 报警信息 -->
        <AlarmInfo :AlarmInfoData="AlarmInfoData" @callParentMethod="parentMethod"/>
        <!-- 运行状态信息 -->
        <RunState :RunStateData="RunStateData"/>
        <!-- 车厢温度信息 -->
        <CarTemperature :CarTemperatureData="CarTemperatureData"/>
    </div>
</template>

<script setup>

import BackTop from "@/components/BackTop.vue"
import AlarmInfo from "./AlarmInfo.vue"
import ActualWarning from "./ActualWarning.vue"
import StateWarning from "./StateWarning.vue"
import CarCrew from "./CarCrew.vue"
import RunState from "./RunState.vue"
import CarTemperature from "./CarTemperature.vue"
import Dropdown from './Dropdown.vue'
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getRealtimeAlarmDetail, getStatusAlert, getAlarmInformation, getRunningState, getTrainTemperature, getRealtimeAlarm } from '@/api/api'
let route = useRoute()
let router = useRouter()

let trainId = '1'    //列车号
let trainNum = '1'   //车厢号
let trainCrew = '1'  //机组
let AlarmInfoData = ref([])  //报警信息
let ActualWarningData = ref([]) //实时报警
let StateWarningData = ref([]) //状态预警
let RunStateData = ref([]) //运行状态信息
let CarTemperatureData = ref([]) //车厢温度信息
let name = ref(['压缩机','通风机','冷凝风机','新风阀','回风阀'])
let lists = ref([])
let trainNo = ref([])
let queryData = ref({
    query: '',
    variables: {state:"05098021"},
    operationName:"MyQuery"
})
const proposeWarn = [
    {
     value: 'ref_leak_u11',
     name: '机组1系统1冷媒泄漏预警',
     classify: '预警',
     title: `1、检查风阀开启状态
            \n2、检查风机运行电流及工作状态
            \n3、确认制冷系统运行压力
            \n4、检查电子膨胀阀系统`
    },{
     value: 'ref_leak_u12',
     name: '机组1系统2冷媒泄漏预警',
     classify: '预警',
     title: `1、检查风阀开启状态
            \n2、检查风机运行电流及工作状态
            \n3、确认制冷系统运行压力
            \n4、检查电子膨胀阀系统`
    },{
     value: 'ref_leak_u21',
     name: '机组2系统1冷媒泄漏预警',
     classify: '预警',
     title: `1、检查风阀开启状态
            \n2、检查风机运行电流及工作状态
            \n3、确认制冷系统运行压力
            \n4、检查电子膨胀阀系统`
    },{
     value: 'ref_leak_u22',
     name: '机组2系统2冷媒泄漏预警',
     classify: '预警',
     title: `1、检查风阀开启状态
            \n2、检查风机运行电流及工作状态
            \n3、确认制冷系统运行压力
            \n4、检查电子膨胀阀系统`
    },{
     value: 'ref_pump_u1',
     name: '机组1制冷（热泵）系统预警',
     classify: '预警',
     title: `1、检查风机及风阀
            \n2、确认压缩机运行状态。
            \n3、检查确认制冷系统冷媒是否泄漏
            \n4、检查电子膨胀阀系统`
    },{
     value: 'ref_pump_u2',
     name: '机组2制冷（热泵）系统预警',
     classify: '预警',
     title: `1、检查风机及风阀
            \n2、确认压缩机运行状态。
            \n3、检查确认制冷系统冷媒是否泄漏
            \n4、检查电子膨胀阀系统`
    },
    {
     value: 'f_cp_u1',
     name: '机组1制冷系统预警',
     classify: '预警',
     title: `1、检查风机及风阀
            \n2、确认压缩机运行状态。
            \n3、检查确认制冷系统冷媒是否泄漏
            \n4、检查电子膨胀阀系统`
    },{
     value: 'f_cp_u2',
     name: '机组2制冷系统预警',
     classify: '预警',
     title: `1、检查风机及风阀
            \n2、确认压缩机运行状态。
            \n3、检查确认制冷系统冷媒是否泄漏
            \n4、检查电子膨胀阀系统`
    },{
     value: 'fat_sensor',
     name: '新风温度传感器预警',
     classify: '预警',
     title: `1、检查温度传感器接线是否良好。
            \n2、重新上电复位，查看故障是否消除。
            \n3、更换新传感器。`
    },{
     value: 'rat_sensor',
     name: '回风温度传感器预警',
     classify: '预警',
     title: `1、检查温度传感器接线是否良好。
            \n2、重新上电复位，查看故障是否消除。
            \n3、更换新传感器。`
    },
    {
     value: 'f_fas',
     name: '新风温度传感器预警',
     classify: '预警',
     title: `1、检查温度传感器接线是否良好。
            \n2、重新上电复位，查看故障是否消除。
            \n3、更换新传感器。`
    },{
     value: 'f_ras',
     name: '回风温度传感器预警',
     classify: '预警',
     title: `1、检查温度传感器接线是否良好。
            \n2、重新上电复位，查看故障是否消除。
            \n3、更换新传感器。`
    },{
     value: 'bflt_tempover',
     name: '车厢温度超温预警',
     classify: '预警',
     title: `1.检查2个机组的制冷系统是否异常`
    },
    {
     value: 'cabin_overtemp',
     name: '车厢温度超温预警',
     classify: '预警',
     title: `1.检查2个机组的制冷系统是否异常`
    },{
     value: 'f_presdiff_u1',
     name: '机组1滤网脏堵预警',
     classify: '预警',
     title: `1.清洗混合风滤网
            \n2.如果滤网不脏，需要检查微压差传感器是否正常`
    },{
     value: 'f_presdiff_u2',
     name: '机组2滤网脏堵预警',
     classify: '预警',
     title: `1.清洗混合风滤网
            \n2.如果滤网不脏，需要检查微压差传感器是否正常`
    },{
     value: 'f_ef_u11',
     name: '机组1通风机1电流预警',
     classify: '预警',
     title: `1、检查风阀状态
            \n2、检查滤网是否安装
            \n3、检查风机轴承`
    }
    ,{
     value: 'f_ef_u12',
     name: '机组1通风机2电流预警',
     classify: '预警',
     title: `1、检查风阀状态
            \n2、检查滤网是否安装
            \n3、检查风机轴承`
    },{
     value: 'f_ef_u21',
     name: '机组2通风机1电流预警',
     classify: '预警',
     title: `1、检查风阀状态
            \n2、检查滤网是否安装
            \n3、检查风机轴承`
    },{
     value: 'f_ef_u22',
     name: '机组2通风机2电流预警',
     classify: '预警',
     title: `1、检查风阀状态
            \n2、检查滤网是否安装
            \n3、检查风机轴承`
    },{
     value: 'f_cf_u11',
     name: '机组1冷凝风机1电流预警',
     classify: '预警',
     title: `1、检查风阀状态
            \n2、检查滤网是否安装
            \n3、检查风机轴承`
    },{
     value: 'f_cf_u12',
     name: '机组1冷凝风机2电流预警',
     classify: '预警',
     title: `1、检查风阀状态
            \n2、检查滤网是否安装
            \n3、检查风机轴承`
    },{
     value: 'f_cf_u21',
     name: '机组2冷凝风机1电流预警',
     classify: '预警',
     title: `1、检查风阀状态
            \n2、检查滤网是否安装
            \n3、检查风机轴承`
    },{
     value: 'f_cf_u22',
     name: '机组2冷凝风机2电流预警',
     classify: '预警',
     title: `1、检查风阀状态
            \n2、检查滤网是否安装
            \n3、检查风机轴承`
    },{
     value: 'f_exufan',
     name: '废排风机电流预警',
     classify: '预警',
     title: `1、检查风阀状态
            \n2、检查滤网是否安装
            \n3、检查风机轴承`
    },{
     value: 'f_fas_u11',
     name: '机组1压缩机1电流预警',
     classify: '预警',
     title: `1、检查冷凝风机
            \n2、检查冷凝器
            \n3、检查压缩机`
    },{
     value: 'f_fas_u12',
     name: '机组1压缩机2电流预警',
     classify: '预警',
     title: `1、检查冷凝风机
            \n2、检查冷凝器
            \n3、检查压缩机`
    },{
     value: 'f_fas_u21',
     name: '机组2压缩机1电流预警',
     classify: '预警',
     title: `1、检查冷凝风机
            \n2、检查冷凝器
            \n3、检查压缩机`
    },{
     value: 'f_fas_u22',
     name: '机组2压缩机2电流预警',
     classify: '预警',
     title: `1、检查冷凝风机
            \n2、检查冷凝器
            \n3、检查压缩机`
    },{
     value: 'f_aq_u1',
     name: '机组1空气质量预警',
     classify: '预警',
     title: `1、检查新风阀开度是否正常
            \n2、检查车厢内空气环境`
    },{
     value: 'f_aq_u2',
     name: '机组2空气质量预警',
     classify: '预警',
     title: `1、检查新风阀开度是否正常
            \n2、检查车厢内空气环境`
    },
    
    
    
    {
     value: 'bflt_airclean_u1',
     name: '空气净化U1故障',
     classify: '预警',
     title: `1、检查空气净化器运行是否正常
            \n2、检查故障反馈线路是否正常`
    },
    {
     value: 'bflt_dfstemp_u11',
     name: '融霜温度传感器1-1故障',
     classify: '预警',
     title: `1、检查温度传感器线路是否存在短路、虚接情况
            \n2、测量温度传感器电阻是否正常`
    },
    {
     value: 'bflt_dfstemp_u12',
     name: '融霜温度传感器1-2故障',
     classify: '预警',
     title: `1、检查温度传感器线路是否存在短路、虚接情况
            \n2、测量温度传感器电阻是否正常`
    },
    {
     value: 'bflt_airclean_u2',
     name: '空气净化U2故障',
     classify: '预警',
     title: `1、检查空气净化器运行是否正常
            \n2、检查故障反馈线路是否正常`
    },
    {
     value: 'bflt_dfstemp_u21',
     name: '融霜温度传感器2-1故障',
     classify: '预警',
     title: `1、检查温度传感器线路是否存在短路、虚接情况
            \n2、测量温度传感器电阻是否正常`
    },
    {
     value: 'bflt_dfstemp_u22',
     name: '融霜温度传感器2-2故障',
     classify: '预警',
     title: `1、检查温度传感器线路是否存在短路、虚接情况
            \n2、测量温度传感器电阻是否正常`
    },
    {
     value: 'bflt_seattemp',
     name: '车厢温度传感器2故障',
     classify: '预警',
     title: `1、检查温度传感器线路是否存在短路、虚接情况
            \n2、测量温度传感器电阻是否正常`
    },
    {
     value: 'bcomuflt_vfd_u11',
     name: '变频器1-1通讯故障',
     classify: '预警',
     title: `1、检查变频器是否上电
            \n2、检查通讯线路是否存在虚接情况`
    },
    {
     value: 'bcomuflt_vfd_u12',
     name: '变频器1-2通讯故障',
     classify: '预警',
     title: `1、检查变频器是否上电
            \n2、检查通讯线路是否存在虚接情况`
    },
    {
     value: 'bcomuflt_vfd_u21',
     name: '变频器2-1通讯故障',
     classify: '预警',
     title: `1、检查变频器是否上电
            \n2、检查通讯线路是否存在虚接情况`
    },
    {
     value: 'bcomuflt_vfd_u22',
     name: '变频器2-2通讯故障',
     classify: '预警',
     title: `1、检查变频器是否上电
            \n2、检查通讯线路是否存在虚接情况`
    },{
     value: 'bmcbflt_pwr_u1',
     name: '机组1供电故障',
     classify: '预警',
     title: `1、检查主回路电路是否短路、漏电流是否过大
            \n2、 检查断路器是否损坏`
    },{
     value: 'bmcbflt_pwr_u2',
     name: '机组2供电故障',
     classify: '预警',
     title: `1、检查主回路电路是否短路、漏电流是否过大
            \n2、 检查断路器是否损坏`
    }]
    const propose = [
    {
     value: 'bflt_tempover',
     name: '车厢温度超温预警',
     classify: '轻微',
     title: `1.检查2个机组的制冷系统是否异常`,
    },
    {
     value: 'bocflt_ef_u11',
     name: '通风机1-1过流故障',
     classify: '严重',
     title: `1.检查风机是否堵转
             \n2.检查电机主电路绝缘是否正常
             \n3.检查电机保护器是否损坏`
    },{
     value: 'bocflt_ef_u12',
     name: '通风机1-2过流故障',
     classify: '严重',
     title: `1.检查风机是否堵转
             \n2.检查电机主电路绝缘是否正常
             \n3.检查电机保护器是否损坏`
    },{
     value: 'bocflt_cf_u11',
     name: '冷凝风机1-1过流故障',
     classify: '中等',
     title: `1.检查风机是否堵转
             \n2.检查电机主电路绝缘是否正常
             \n3.检查电机保护器是否损坏`
    },{
     value: 'bocflt_cf_u12',
     name: '冷凝风机1-2过流故障',
     classify: '中等',
     title: `1.检查风机是否堵转
             \n2.检查电机主电路绝缘是否正常
             \n3.检查电机保护器是否损坏`
    },{
     value: 'bflt_vfd_u11',
     name: '变频器1-1故障',
     classify: '中等',
     title: `1.检查变频器运行是否正常
             \n2.检查反馈线路是否正常`
    },{
     value: 'bflt_vfd_com_u11',
     name: '变频器通讯故障U1-1',
     classify: '中等',
     title: `1.检查变频器上通讯接口针孔件连接是否牢靠，并重新紧固
             \n2.检查扩展模块上通讯接口插头连接是否牢靠，针孔件是否安装到位，并重新紧固
             \n3.检查变频器和扩展模块之间各接点位置连线是否有虚接情况，重新紧固
             \n4.更换变频器,重新连接测试\n5.更换扩展模块，重新连接测试`
    },{
     value: 'bflt_vfd_u12',
     name: '变频器1-2故障',
     classify: '中等',
     title: `1.检查变频器运行是否正常
            \n2.检查反馈线路是否正常`
    },{
     value: 'bflt_vfd_com_u12',
     name: '变频器通讯故障U1-2',
     classify: '中等',
     title: `1.检查变频器上通讯接口针孔件连接是否牢靠，并重新紧固
            \n2.检查扩展模块上通讯接口插头连接是否牢靠，针孔件是否安装到位，并重新紧固
            \n3.检查变频器和扩展模块之间各接点位置连线是否有虚接情况，重新紧固
            \n4.更换变频器,重新连接测试
            \n5.更换扩展模块，重新连接测试`
    },{
     value: 'blpflt_comp_u11',
     name: '压缩机1-1低压故障',
     classify: '中等',
     title: `1.检查制冷剂是否泄漏
            \n2.检查通风机运转是否正常
            \n3.检查低压开关是否损坏`
    },{
     value: 'bscflt_comp_u11',
     name: '压缩机1-1高压连锁故障',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查高压开关、排气温度保护是否正常
            \n3.检查反馈线路是否正常`
    },{
     value: 'bscflt_vent_u11',
     name: '排气故障U1-1',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.检查空调四周散热是否顺畅`
    },{
     value: 'blpflt_comp_u12',
     name: '压缩机1-2低压故障',
     classify: '中等',
     title: `1.检查制冷剂是否泄漏
            \n2.检查通风机运转是否正常
            \n3.检查低压开关是否损坏`
    },{
     value: 'bscflt_comp_u12',
     name: '压缩机1-2高压连锁故障',
     classify: '中等',
     title: `1.检查冷凝风机运转是否正常
            \n2.检查高压开关、排气温度保护是否正常
            \n3.检查反馈线路是否正常`
    },{
     value: 'bscflt_vent_u12',
     name: '排气故障U1-2',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.检查空调四周散热是否顺畅`
    },{
     value: 'bflt_fad_u1',
     name: '新风阀U1故障',
     classify: '轻微',
     title: `1.检查风阀执行器动作、反馈指示是否正常
            \n2.检查全开反馈线路是否正常`
    },
    


    {
     value: 'bflt_fad_u11',
     name: '新风阀故障U1-1',
     classify: '轻微',
     title: `1.使用PTU控制新风阀全开，接着全关，观察新风阀是否正常动作
            \n2.新风阀动作正常的情况下，手动打开新风阀到全开位置，在本车厢触摸屏或PTU上查看新风阀的反馈信号是否导通
            \n3.如果反馈信号异常，检查反馈回路各接线位置是否虚接`
    },{
     value: 'bflt_fad_u12',
     name: '新风阀故障U1-2',
     classify: '轻微',
     title: `1.使用PTU控制新风阀全开，接着全关，观察新风阀是否正常动作
            \n2.新风阀动作正常的情况下，手动打开新风阀到全开位置，在本车厢触摸屏或PTU上查看新风阀的反馈信号是否导通
            \n3.如果反馈信号异常，检查反馈回路各接线位置是否虚接`
    },{
     value: 'bflt_rad_u1',
     name: '回风阀U1故障',
     classify: '轻微',
     title: `1.检查风阀执行器动作、反馈指示是否正常
            \n2.检查全开反馈线路是否正常`
    },
    
    
    {
     value: 'bflt_rad_u11',
     name: '回风阀故障U1-1',
     classify: '轻微',
     title: `1.使用PTU控制新风阀全开，接着全关，观察新风阀是否正常动作
            \n2.新风阀动作正常的情况下，手动打开新风阀到全开位置，在本车厢触摸屏或PTU上查看新风阀的反馈信号是否导通
            \n3.如果反馈信号异常，检查反馈回路各接线位置是否虚接`
    },{
     value: 'bflt_rad_u12',
     name: '回风阀故障U1-2',
     classify: '轻微',
     title: `1.使用PTU控制新风阀全开，接着全关，观察新风阀是否正常动作
            \n2.新风阀动作正常的情况下，手动打开新风阀到全开位置，在本车厢触摸屏或PTU上查看新风阀的反馈信号是否导通
            \n3.如果反馈信号异常，检查反馈回路各接线位置是否虚接`
    },{
     value: 'bflt_ap_u11',
     name: '空气净化器故障U1',
     classify: '轻微',
     title: `1.使用PTU控制启动空气净化器，10s后查看空气净化器的反馈信号是否导通
            \n2.如果反馈信号断开状态，检查反馈回路各接点是否有虚接情况
            \n3.更换新空气净化器测试`
    },{
     value: 'bflt_expboard_u1',
     name: '扩展模块U1故障',
     classify: '中等',
     title: `1.检查扩展模块供电是否正常
            \n2.检查扩展模块通讯线路是否存在虚接情况`
    },{
     value: 'bflt_frstemp_u1',
     name: '新风温度传感器U1故障',
     classify: '轻微',
     title: `1.检查温度传感器线路是否存在短路、虚接情况
            \n2.测量温度传感器电阻是否正常`
    },{
     value: 'bflt_rnttemp_u1',
     name: '回风温度传感器U1故障',
     classify: '轻微',
     title:`1.检查温度传感器线路是否存在短路、虚接情况
            \n2.测量温度传感器电阻是否正常`
    },{
     value: 'bflt_splytemp_u11',
     name: '送风温度传感器1-1故障',
     classify: '轻微',
     title: `1.检查温度传感器线路是否存在短路、虚接情况
            \n2.测量温度传感器电阻是否正常`
    },
    {
     value: 'bflt_splytemp_u12',
     name: '送风温度传感器1-2故障',
     classify: '轻微',
     title: `1.检查温度传感器线路是否存在短路、虚接情况
            \n2.测量温度传感器电阻是否正常`
    },{
     value: 'bflt_coiltemp_u11',
     name: '融霜温度传感器故障U1-1',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_coiltemp_u12',
     name: '融霜温度传感器故障U1-2',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_insptemp_u11',
     name: '吸气温度传感器故障U1-1',
     classify: '中等',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_insptemp_u12',
     name: '吸气温度传感器故障U1-2',
     classify: '中等',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_lowpres_u11',
     name: '低压压力传感器故障U1-1',
     classify: '中等',
     title: `1.检查吸气压力传感器和扩展模块之间的接线点位是否牢固，重新紧固
            \n2.更换新的压力传感器，然后通过PTU顺序启动制冷，查看压力变化值是否正常`
    },{
     value: 'bflt_lowpres_u12',
     name: '低压压力传感器故障U1-2',
     classify: '中等',
     title: `1.检查吸气压力传感器和扩展模块之间的接线点位是否牢固，重新紧固
            \n2.更换新的压力传感器，然后通过PTU顺序启动制冷，查看压力变化值是否正常`
    },{
     value: 'bflt_highpres_u11',
     name: '高压压力传感器故障U1-1',
     classify: '中等',
     title: `1.检查高压传感器和扩展模块之间的接线点位是否牢固，重新紧固
            \n2.更换新的压力传感器，然后通过PTU顺序启动制冷，查看压力变化值是否正常`
    },{
     value: 'bflt_highpres_u12',
     name: '高压压力传感器故障U1-2',
     classify: '中等',
     title: `1.检查高压传感器和扩展模块之间的接线点位是否牢固，重新紧固
            \n2.更换新的压力传感器，然后通过PTU顺序启动制冷，查看压力变化值是否正常`
    },{
     value: 'bflt_diffpres_u1',
     name: '压差传感器故障U1',
     classify: '轻微',
     title: `1.检查混合风滤网是否脏堵特别严重，清洗后再测试差压传感器
            \n2.检查压差传感器和扩展模块之间的接线点位是否牢固，重新紧固
            \n3.更换新的压差传感器，开启通风机，重新测试`
    },{
     value: 'bocflt_ef_u21',
     name: '通风机2-1过流故障',
     classify: '严重',
     title: `1.检查风机是否堵转；检查电机主电路绝缘是否正常
            \n2.检查电机保护器是否损坏`
    },{
     value: 'bocflt_ef_u22',
     name: '通风机2-2过流故障',
     classify: '严重',
     title: `1.检查风机是否堵转；检查电机主电路绝缘是否正常
            \n2.检查电机保护器是否损坏`
    },{
     value: 'bocflt_cf_u21',
     name: '冷凝风机2-1过流故障',
     classify: '中等',
     title: `1.检查风机是否堵转；检查电机主电路绝缘是否正常
            \n2.检查电机保护器是否损坏`
    },{
     value: 'bocflt_cf_u22',
     name: '冷凝风机过流故障U2-2',
     classify: '中等',
     title: `1.检查风机是否堵转；检查电机主电路绝缘是否正常
            \n2.检查电机保护器是否损坏`
    },{
     value: 'bflt_vfd_u21',
     name: '变频器2-1故障',
     classify: '中等',
     title: `1.检查变频器运行是否正常
            \n2.检查反馈线路是否正常`
    },{
     value: 'bflt_vfd_com_u21',
     name: '变频器通讯故障U2-1',
     classify: '中等',
     title: `1.检查变频器上通讯接口针孔件连接是否牢靠，并重新紧固
            \n2.检查扩展模块上通讯接口插头连接是否牢靠，针孔件是否安装到位，并重新紧固
            \n3.检查变频器和扩展模块之间各接点位置连线是否有虚接情况，重新紧固
            \n4.更换变频器,重新连接测试
            \n5.更换扩展模块，重新连接测试`
    },{
     value: 'bflt_vfd_u22',
     name: '变频器2-2故障',
     classify: '中等',
     title: `1.检查变频器运行是否正常
            \n2.检查反馈线路是否正常`
    },{
     value: 'bflt_vfd_com_u22',
     name: '变频器通讯故障U2-2',
     classify: '中等',
     title: `1.检查变频器上通讯接口针孔件连接是否牢靠，并重新紧固
            \n2.检查扩展模块上通讯接口插头连接是否牢靠，针孔件是否安装到位，并重新紧固
            \n3.检查变频器和扩展模块之间各接点位置连线是否有虚接情况，重新紧固
            \n4.更换变频器,重新连接测试
            \n5.更换扩展模块，重新连接测试`
    },{
     value: 'blpflt_comp_u21',
     name: '压缩机2-1低压故障',
     classify: '中等',
     title: `1.检查制冷剂是否泄漏
            \n2.检查通风机运转是否正常
            \n3.检查低压开关是否损坏`
    },{
     value: 'bscflt_comp_u21',
     name: '压缩机2-1高压连锁故障',
     classify: '中等',
     title: `1.检查冷凝风机运转是否正常
            \n2.检查高压开关、排气温度保护是否正常
            \n3.检查反馈线路是否正常`
    },{
     value: 'bscflt_vent_u21',
     name: '排气故障U2-1',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.检查空调四周散热是否顺畅`
    },{
     value: 'blpflt_comp_u22',
     name: '压缩机2-2低压故障',
     classify: '中等',
     title: `1.检查制冷剂是否泄漏
            \n2.检查通风机运转是否正常
            \n3.检查低压开关是否损坏`
    },{
     value: 'bscflt_comp_u22',
     name: '压缩机2-2高压连锁故障',
     classify: '中等',
     title: `1.检查冷凝风机运转是否正常
            \n2.检查高压开关、排气温度保护是否正常
            \n3.检查反馈线路是否正常`
    },{
     value: 'bscflt_vent_u22',
     name: '排气故障U2-2',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.检查空调四周散热是否顺畅`
    },
    {
     value: 'bflt_fad_u2',
     name: '新风阀U2故障',
     classify: '中等',
     title: `1.检查风阀执行器动作、反馈指示是否正常
            \n2.检查全开反馈线路是否正常`
    },
    
    
    {
     value: 'bflt_fad_u21',
     name: '新风阀故障U2-1',
     classify: '中等',
     title: `1.使用PTU控制新风阀全开，接着全关，观察新风阀是否正常动作
            \n2.新风阀动作正常的情况下，手动打开新风阀到全开位置，在本车厢触摸屏或PTU上查看新风阀的反馈信号是否导通
            \n3.如果反馈信号异常，检查反馈回路各接线位置是否虚接`
    },{
     value: 'bflt_fad_u22',
     name: '新风阀故障U2-2',
     classify: '中等',
     title: `1.使用PTU控制新风阀全开，接着全关，观察新风阀是否正常动作
            \n2.新风阀动作正常的情况下，手动打开新风阀到全开位置，在本车厢触摸屏或PTU上查看新风阀的反馈信号是否导通
            \n3.如果反馈信号异常，检查反馈回路各接线位置是否虚接`
    },{
     value: 'bflt_rad_u2',
     name: '回风阀故障U2-1',
     classify: '轻微',
     title: `1.检查风阀执行器动作、反馈指示是否正常
            \n2.检查全开反馈线路是否正常`
    },
    
    {
     value: 'bflt_rad_u21',
     name: '回风阀故障U2-1',
     classify: '轻微',
     title: `1.使用PTU控制新风阀全开，接着全关，观察新风阀是否正常动作
            \n2.新风阀动作正常的情况下，手动打开新风阀到全开位置，在本车厢触摸屏或PTU上查看新风阀的反馈信号是否导通
            \n3.如果反馈信号异常，检查反馈回路各接线位置是否虚接`
    },{
     value: 'bflt_rad_u22',
     name: '回风阀故障U2-2',
     classify: '轻微',
     title: `1.使用PTU控制新风阀全开，接着全关，观察新风阀是否正常动作
            \n2.新风阀动作正常的情况下，手动打开新风阀到全开位置，在本车厢触摸屏或PTU上查看新风阀的反馈信号是否导通
            \n3.如果反馈信号异常，检查反馈回路各接线位置是否虚接`
    },{
     value: 'bflt_ap_u21',
     name: '空气净化器故障U2',
     classify: '轻微',
     title: `1.使用PTU控制启动空气净化器，10s后查看空气净化器的反馈信号是否导通
            \n2.如果反馈信号断开状态，检查反馈回路各接点是否有虚接情况
            \n3.更换新空气净化器测试`
    },{
     value: 'bflt_expboard_u2',
     name: '扩展模块U2故障',
     classify: '中等',
     title: `1.检查扩展模块供电是否正常
            \n2.检查扩展模块通讯线路是否存在虚接情况`
    },{
     value: 'bflt_frstemp_u2',
     name: '新风温度传感器U2故障',
     classify: '轻微',
     title: `1.检查温度传感器线路是否存在短路、虚接情况
            \n2.测量温度传感器电阻是否正常`
    },{
     value: 'bflt_rnttemp_u2',
     name: '回风温度传感器U2故障',
     classify: '轻微',
     title: `1.检查温度传感器线路是否存在短路、虚接情况
            \n2.测量温度传感器电阻是否正常`
    },{
     value: 'bflt_splytemp_u21',
     name: '送风温度传感器2-1故障',
     classify: '轻微',
     title: `1.检查温度传感器线路是否存在短路、虚接情况
            \n2.测量温度传感器电阻是否正常`
    },{
     value: 'bflt_splytemp_u22',
     name: '送风温度传感器2-2故障',
     classify: '轻微',
     title: `1.检查温度传感器线路是否存在短路、虚接情况
            \n2.测量温度传感器电阻是否正常`
    },{
     value: 'bflt_coiltemp_u21',
     name: '融霜温度传感器故障U2-1',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_coiltemp_u22',
     name: '融霜温度传感器故障U2-2',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_insptemp_u21',
     name: '吸气温度传感器故障U2-1',
     classify: '中等',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_insptemp_u22',
     name: '吸气温度传感器故障U2-2',
     classify: '中等',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_lowpres_u21',
     name: '低压压力传感器故障U2-1',
     classify: '中等',
     title: `1.检查吸气压力传感器和扩展模块之间的接线点位是否牢固，重新紧固
            \n2.更换新的压力传感器，然后通过PTU顺序启动制冷，查看压力变化值是否正常`
    },{
     value: 'bflt_lowpres_u22',
     name: '低压压力传感器故障U2-2',
     classify: '中等',
     title: `1.检查吸气压力传感器和扩展模块之间的接线点位是否牢固，重新紧固
            \n2.更换新的压力传感器，然后通过PTU顺序启动制冷，查看压力变化值是否正常`
    },{
     value: 'bflt_highpres_u21',
     name: '高压压力传感器故障U2-1',
     classify: '中等',
     title: `1.检查高压传感器和扩展模块之间的接线点位是否牢固，重新紧固
            \n2.更换新的压力传感器，然后通过PTU顺序启动制冷，查看压力变化值是否正常`
    },{
     value: 'bflt_highpres_u22',
     name: '高压压力传感器故障U2-2',
     classify: '中等',
     title: `1.检查高压传感器和扩展模块之间的接线点位是否牢固，重新紧固
            \n2.更换新的压力传感器，然后通过PTU顺序启动制冷，查看压力变化值是否正常`
    },{
     value: 'bflt_diffpres_u2',
     name: '压差传感器故障U2',
     classify: '轻微',
     title: `1.检查混合风滤网是否脏堵特别严重，清洗后再测试差压传感器
            \n2.检查压差传感器和扩展模块之间的接线点位是否牢固，重新紧固
            \n3.更换新的压差传感器，开启通风机，重新测试`
    },{
     value: 'bflt_emergivt',
     name: '紧急逆变器故障',
     classify: '严重',
     title: `1.检查紧急逆变器运行是否正常
            \n2.检查故障反馈线路是否正常`
    },{
     value: 'bflt_powersupply_u1',
     name: '机组1供电故障',
     classify: '严重',
     title: `1.合上断路器，通过PTU顺序启动制冷，将压缩机频率调到最大65HZ,测量三相电流，各相电流应该都小于50A
            \n2.空调停机，检查总电源回路是否有短路
            \n3.使用绝缘测试仪，检测三相回路绝缘电阻大于5MΩ
            \n4.检测辅助触点反馈信号回路是否有虚接情况`
    },{
     value: 'bflt_powersupply_u2',
     name: '机组2供电故障',
     classify: '严重',
     title: `1.合上断路器，通过PTU顺序启动制冷，将压缩机频率调到最大65HZ,测量三相电流，各相电流应该都小于50A
            \n2.空调停机，检查总电源回路是否有短路
            \n3.使用绝缘测试仪，检测三相回路绝缘电阻大于5MΩ
            \n4.检测辅助触点反馈信号回路是否有虚接情况`
    },{
     value: 'bflt_exhaustval',
     name: '废排风阀故障',
     classify: '轻微',
     title: `1.使用PTU控制废排风阀全开，接着全关，观察废排风阀是否正常动作
            \n2.动作正常的情况下，手动打开风阀到全开位置，在本车厢触摸屏或PTU上查看风阀的反馈信号是否导通
            \n3.如果反馈信号异常，检查反馈回路各接线位置是否虚接`
    },{
     value: 'bflt_tcms',
     name: 'MVB 故障',
     classify: '严重',
     title: `1.检查通讯电缆接头是否松动，重新紧固
            \n2.检查通讯电缆各接点位置是否虚接
            \n3.更换新的主控制器，重新连接测试`
    },{
     value: 'bflt_exhaustfan',  //不
     name: '废排风机过流故障',
     classify: '中等',
     title: `1.首先检查电机保护器是否跳闸，如果没有跳闸那就是反馈回路断开触发的故障
            \n2.手动开合电机保护器，通过PTU或本车厢触摸屏查看输入信号状态，检查反馈回路是否正常
            \n3.跳闸引起的原因需要检查三相电路是否缺相或有虚接情况
            \n4.检查电机自身三相电阻是否平衡，用绝缘测试仪测试绝缘电阻是否大于5MΩ
            \n5.检查电机保护器主触点之间是否正常导通，电阻小于20Ω。如有异常，更换新物料`
    },{
     value: 'bflt_airmon_u1',
     name: '机组1空气监测终端通讯故障',
     classify: '轻微',
     title: `1.检查空气监测终端供电是否正常，DC24V
            \n2.检查空气监测终端通讯接口插头是否牢固，针孔件是否安装到位，并重新紧固
            \n3.检查扩展模块上通讯接口插头连接是否牢靠，针孔件是否安装到位，并重新紧固
            \n4.检查扩展模块和空气监测终端之间各接点位置连线是否有虚接情况，重新紧固
            \n5.更换空气监测终端，重新连接测试`
    },{
     value: 'bflt_airmon_u2',
     name: '机组2空气监测终端通讯故障',
     classify: '轻微',
     title: `1.检查空气监测终端供电是否正常，DC24V
            \n2.检查空气监测终端通讯接口插头是否牢固，针孔件是否安装到位，并重新紧固
            \n3.检查扩展模块上通讯接口插头连接是否牢靠，针孔件是否安装到位，并重新紧固
            \n4.检查扩展模块和空气监测终端之间各接点位置连线是否有虚接情况，重新紧固
            \n5.更换空气监测终端，重新连接测试`
    },{
     value: 'bflt_currentmon',
     name: '电流监测单元通讯故障',
     classify: '轻微',
     title: `1.检查电流监测单元供电是否正常，DC24V
            \n2.检查电流监测单元通讯接口插头是否牢固，针孔件是否安装到位，并重新紧固
            \n3.检查主控制器上通讯接口插头连接是否牢靠，针孔件是否安装到位，并重新紧固
            \n4.检查主控制器和电流监测单元之间各接点位置连线是否有虚接情况，重新紧固
            \n5.更换电流监测单元，重新连接测试`
    },
    {
     value: 'bFlt_VehTemp',
     name: '车厢温度传感器1故障',
     classify: '轻微',
     title: `1.检查温度传感器线路是否存在短路、虚接情况
             \n2.测量温度传感器电阻是否正常`
    },
    // 旧
    ,{
     value: 'bflt_vehtemp_u2',
     name: '车厢温度传感器2故障',
     classify: '轻微',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'bflt_vehtemp_u1',
     name: '车厢温度传感器1故障',
     classify: '轻微',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },
]
// const propose = [
//     {
//      value: 'dvc_bocflt_ef_u11',
//      name: '通风机1-1过流故障',
//      title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
//     },{
//      value: 'dvc_bocflt_ef_u12',
//      name: '通风机1-2过流故障',
//      title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
//     },{
//      value: 'dvc_bocflt_cf_u11',
//      name: '冷凝风机1-1过流故障',
//      title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
//     },{
//      value: 'dvc_bocflt_cf_u12',
//      name: '冷凝风机1-2过流故障',
//      title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
//     },{
//      value: 'dvc_bflt_vfd_u11',
//      name: '变频器1-1故障',
//      title: '检查变频器运行是否正常；检查反馈线路是否正常'
//     },{
//      value: 'dvc_blpflt_comp_u11',
//      name: '压缩机1-1低压故障',
//      title: '检查制冷剂是否泄漏；检查通风机运转是否正常；检查低压开关是否损坏'
//     },{
//      value: 'dvc_bscflt_comp_u11',
//      name: '压缩机1-1高压连锁故障',
//      title: '检查冷凝风机运转是否正常；检查高压开关、排气温度保护是否正常；检查反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_vfd_u12',
//      name: '变频器1-2故障',
//      title: '检查变频器运行是否正常；检查反馈线路是否正常'
//     },{
//      value: 'dvc_blpflt_comp_u12',
//      name: '压缩机1-2低压故障',
//      title: '检查制冷剂是否泄漏；检查通风机运转是否正常；检查低压开关是否损坏'
//     },{
//      value: 'dvc_bscflt_comp_u12',
//      name: '压缩机1-2高压连锁故障',
//      title: '检查冷凝风机运转是否正常；检查高压开关、排气温度保护是否正常；检查反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_eev_u11',
//      name: '电子膨胀阀1-1故障',
//      title: '检查电子膨胀阀控制器显示参数是否正常；检查故障反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_eev_u12',
//      name: '电子膨胀阀1-2故障',
//      title: '检查电子膨胀阀控制器显示参数是否正常；检查故障反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_fad_u1',
//      name: '新风阀U1故障',
//      title: '检查风阀执行器动作、反馈指示是否正常；检查全开反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_rad_u1',
//      name: '回风阀U1故障',
//      title: '检查风阀执行器动作、反馈指示是否正常；检查全开反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_airclean_u1',
//      name: '空气净化U1故障',
//      title: '检查空气净化器运行是否正常；检查故障反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_expboard_u1',
//      name: '扩展模块U1故障',
//      title: '检查扩展模块供电是否正常；检查扩展模块通讯线路是否存在虚接情况'
//     },{
//      value: 'dvc_bflt_frstemp_u1',
//      name: '新风温度传感器U1故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_splytemp_u11',
//      name: '送风温度传感器1-1故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_splytemp_u12',
//      name: '送风温度传感器1-2故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_rnttemp_u1',
//      name: '回风温度传感器U1故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_dfstemp_u11',
//      name: '融霜温度传感器1-1故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_dfstemp_u12',
//      name: '融霜温度传感器1-2故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bocflt_ef_u21',
//      name: '通风机2-1过流故障',
//      title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
//     },{
//      value: 'dvc_bocflt_ef_u22',
//      name: '通风机2-2过流故障',
//      title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
//     },{
//      value: 'dvc_bocflt_cf_u21',
//      name: '冷凝风机2-1过流故障',
//      title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
//     },{
//      value: 'dvc_bocflt_cf_u22',
//      name: '冷凝风机2-2过流故障',
//      title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
//     },{
//      value: 'dvc_bflt_vfd_u21',
//      name: '变频器2-1故障',
//      title: '检查变频器运行是否正常；检查反馈线路是否正常'
//     },{
//      value: 'dvc_blpflt_comp_u21',
//      name: '压缩机2-1低压故障',
//      title: '检查制冷剂是否泄漏；检查通风机运转是否正常；检查低压开关是否损坏'
//     },{
//      value: 'dvc_bscflt_comp_u21',
//      name: '压缩机2-1高压连锁故障',
//      title: '检查冷凝风机运转是否正常；检查高压开关、排气温度保护是否正常；检查反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_vfd_u22',
//      name: '变频器2-2故障',
//      title: '检查变频器运行是否正常；检查反馈线路是否正常'
//     },{
//      value: 'dvc_blpflt_comp_u22',
//      name: '压缩机2-2低压故障',
//      title: '检查制冷剂是否泄漏；检查通风机运转是否正常；检查低压开关是否损坏'
//     },{
//      value: 'dvc_bscflt_comp_u22',
//      name: '压缩机2-2高压连锁故障',
//      title: '检查冷凝风机运转是否正常；检查高压开关、排气温度保护是否正常；检查反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_eev_u21',
//      name: '电子膨胀阀2-1故障',
//      title: '检查电子膨胀阀控制器显示参数是否正常；检查故障反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_eev_u22',
//      name: '电子膨胀阀2-2故障',
//      title: '检查电子膨胀阀控制器显示参数是否正常；检查故障反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_fad_u2',
//      name: '新风阀U2故障',
//      title: '检查风阀执行器动作、反馈指示是否正常；检查全开反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_rad_u2',
//      name: '回风阀U2故障',
//      title: '检查风阀执行器动作、反馈指示是否正常；检查全开反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_airclean_u2',
//      name: '空气净化U2故障',
//      title: '检查空气净化器运行是否正常；检查故障反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_expboard_u2',
//      name: '扩展模块U2故障',
//      title: '检查扩展模块供电是否正常；检查扩展模块通讯线路是否存在虚接情况'
//     },{
//      value: 'dvc_bflt_frstemp_u2',
//      name: '新风温度传感器U2故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_splytemp_u21',
//      name: '送风温度传感器2-1故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_splytemp_u22',
//      name: '送风温度传感器2-2故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_rnttemp_u2',
//      name: '回风温度传感器U2故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_dfstemp_u21',
//      name: '融霜温度传感器2-1故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_dfstemp_u22',
//      name: '融霜温度传感器2-2故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_vehtemp',
//      name: '车厢温度传感器1故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_seattemp',
//      name: '车厢温度传感器2故障',
//      title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
//     },{
//      value: 'dvc_bflt_emergivt',
//      name: '紧急逆变器故障',
//      title: '检查紧急逆变器运行是否正常；检查故障反馈线路是否正常'
//     },{
//      value: 'dvc_bflt_mvbbus',
//      name: 'MVB 通讯故障',
//      title: '检查控制器供电是否正常；检查MVB连接器是否松动'
//     },{
//      value: 'dvc_bcomuflt_vfd_u11',
//      name: '变频器1-1通讯故障',
//      title: '检查变频器是否上电；检查通讯线路是否存在虚接情况'
//     },{
//      value: 'dvc_bcomuflt_vfd_u12',
//      name: '变频器1-2通讯故障',
//      title: '检查变频器是否上电；检查通讯线路是否存在虚接情况'
//     },{
//      value: 'dvc_bcomuflt_vfd_u21',
//      name: '变频器2-1通讯故障',
//      title: '检查变频器是否上电；检查通讯线路是否存在虚接情况'
//     },{
//      value: 'dvc_bcomuflt_vfd_u22',
//      name: '变频器2-2通讯故障',
//      title: '检查变频器是否上电；检查通讯线路是否存在虚接情况'
//     },{
//      value: 'dvc_bcomuflt_eev_u11',
//      name: '电子膨胀阀1-1通讯故障',
//      title: '检查电子膨胀阀控制器供电是否正常；检查通讯线路是否存在虚接情况'
//     },{
//      value: 'dvc_bcomuflt_eev_u12',
//      name: '电子膨胀阀1-2通讯故障',
//      title: '检查电子膨胀阀控制器供电是否正常；检查通讯线路是否存在虚接情况'
//     },{
//      value: 'dvc_bcomuflt_eev_u21',
//      name: '电子膨胀阀2-1通讯故障',
//      title: '检查电子膨胀阀控制器供电是否正常；检查通讯线路是否存在虚接情况'
//     },{
//      value: 'dvc_bcomuflt_eev_u22',
//      name: '电子膨胀阀2-2通讯故障',
//      title: '检查电子膨胀阀控制器供电是否正常；检查通讯线路是否存在虚接情况'
//     },{
//      value: 'dvc_bmcbflt_pwr_u1',
//      name: '机组1供电故障',
//      title: '检查主回路电路是否短路、漏电流是否过大；检查断路器是否损坏'
//     },{
//      value: 'dvc_bmcbflt_pwr_u2',
//      name: '机组2供电故障',
//      title: '检查主回路电路是否短路、漏电流是否过大；检查断路器是否损坏'
//     }
// ]
function getLocalTime(i,data) { 
    if (typeof i !== 'number') return;
 
   var d = new Date(data); 
//   2023-02-20T13:50:20.954Z
//   2023-02-08T00:00:00+00:00
    //得到1970年一月一日到现在的秒数 
    var len = d.getTime();
    
    //本地时间与GMT时间的时间偏移差(注意：GMT这是UTC的民间名称。GMT=UTC）
    var offset = d.getTimezoneOffset() * 60000;
 
    //得到现在的格林尼治时间
    var utcTime = len + offset;
 
    return new Date(utcTime + 3600000 * i);
}
function newDate(time) {
    let temp = time.split('T')
    const year = temp[0].slice(0,4)
    const month = temp[0].slice(5,7)
     const day = temp[0].slice(8,10)
     const times = temp[1].slice(0,8)
     return `${year}-${month}-${day}-${times}`
    // var date = new Date(time)
    // var y = date.getFullYear()
    // var m = date.getMonth() + 1
    // m = m < 10 ? '0' + m : m
    // var d = date.getDate()
    // d = d < 10 ? '0' + d : d
    // var h = date.getHours()
    // h = h < 10 ? '0' + h : h
    // var minute = date.getMinutes()
    // minute = minute < 10 ? '0' + minute : minute
    // var s = date.getSeconds()
    // s = s < 10 ? '0' + s : s
    // return y + '-' + m + '-' + d + ' ' + h + ':' + minute + ':' + s
}

  function runningState(items){
     return  items ==1?'运行':'停机'
  }
  function runningKeys(name){
    const data = ['ef_u1','ef_u2','cf_u1','cf_u2','fad_u1','fad_u2','comp_u21','comp_u22']
    if(data.indexOf(name)!==-1){

    }
  }
  
  function setName (obj,i){
    

    switch(i){
        case 'ef_u1':
        obj.equipmentName='通风机';
        break;

        case 'cf_u1':
        obj.equipmentName='冷凝风机';
        break;

        case 'fad_u1':
        obj.equipmentName='新风阀';
        break;

         case 'comp_u21':
        obj.equipmentName='回风阀';
        break;
        default:
    }
  }
  function containsNumber(str) {
	    var reg=/\d/;
	    return reg.test(str);
	}
    function hasLetter(str) {
    for (var i in str) {
        var asc = str.charCodeAt(i);
        if ((asc >= 65 && asc <= 90 || asc >= 97 && asc <= 122)) {
            return true;
        }
    }
    return false;
}
function extractNumber(str) {
    // 正则表达式用于匹配数字
    const regex = /(\d+)/;
    const match = str.match(regex);

    // 如果匹配成功，返回匹配到的第一个数字
    if (match) {
        return String(parseInt(match[1], 10));
    }

    // 如果没有匹配到数字，返回 null 或者可以根据需求抛出错误
    return null;
}
function getCurrentDate() {
    const now = new Date();

    // 获取完整的年份 (4位, 1970-?)
    const year = now.getFullYear();

    // 获取当前月份 (0-11, 0代表1月)
    const month = now.getMonth() + 1;

    // 获取当前日 (1-31)
    const day = now.getDate();

    // 格式化月份和日期，确保始终为两位数
    const formattedMonth = month > 9 ? month : '0' + month;
    const formattedDay = day > 9 ? day : '0' + day;

    return `${year}-${formattedMonth}-${formattedDay}`;
}
const startTime = ref(getCurrentDate()+' 00:00:00')
const endTime = ref(getCurrentDate()+' 23:59:59')
function parentMethod(data){
console.log(data,888);
startTime.value = data[0] + ' 00:00:00'
endTime.value = data[1]+' 23:59:59'
listData()
}
async function listData (){
    // AlarmInfoData.value = []
    AlarmInfoData.value = []
    for (let i=1;i<7;i++){
        const params = {
            state: route.query.trainNo+'0'+i+'1',
            startTime: startTime.value,
            endTime: endTime.value
        }
    const logStr = await getAlarmInformation(params)
    const proposeAdvice = propose.concat(proposeWarn)
    if(logStr.alarm_timeline.length !==0){
        logStr.alarm_timeline.forEach((item)=>{
            item.state = item.state.slice(0, -1)
            item.start_time = newDate(item.start_time)
            item.end_time = newDate(item.end_time)
            // console.log(item.propose,extractNumber(item.propose),extractNumber(item.propose).length);
            if(!extractNumber(item.propose)){
                item.unit_no = '-'
            } else if(extractNumber(item.propose) && extractNumber(item.propose).length == 2){
                if(extractNumber(item.propose).slice(-2,-1) == '1'){
                    item.unit_no = '1机组'
                } else {
                    item.unit_no = '2机组'
                }
            } else if(extractNumber(item.propose) && extractNumber(item.propose).length == 1){
                if(extractNumber(item.propose) == '1'){
                    item.unit_no = '1机组'
                } else {
                    item.unit_no = '2机组'
                }
            }
            proposeAdvice.forEach((itemValue)=>{
                if(itemValue.value == item.propose){
                    item.precautions = itemValue.title
                    item.name = itemValue.name
                    item.classify = itemValue.classify
                }
            })
            AlarmInfoData.value.push(item)
        })
    }
    }
    }
function getTrainApi() {
    trainNo.value = []
    trainNo.value.push(route.query.trainNo)
    getRealtimeAlarm({trainNo: trainNo.value}).then((res)=>{
        // res.vw_train_alarm_info[0].bflt_airmon_u1 = 1
        ActualWarningData.value = []
        // AlarmInfoData.value = []
        res.vw_train_alarm_info.forEach((item)=>{
            Object.keys(item).forEach((value)=>{
                console.log(item,value);
                if(item[value]>0 && item[value +'_name']){
                    console.log(item,value);
                    ActualWarningData.value.push({
                        name:item[value +'_name'],
                        code: value 
                    })
                }
            })
            console.log(ActualWarningData.value);
            ActualWarningData.value.forEach((values)=>{
                // 原始
                values.alarm_time = newDate(item.alarm_time)
                values.carriage_no = item.carriage_no
                propose.forEach((proposeValue)=>{
                    if(proposeValue.value == values.code){
                    values.precautions = proposeValue.title
                    values.name = proposeValue.name
                }
                })
            })
            console.log('888',ActualWarningData.value);
        })
    })
    // getRealtimeAlarmDetail(route.query.trainId).then((res)=>{
    //     ActualWarningData.value = []
    //     Object.values(res).forEach((item)=>{
    //     if(item.length !== 0){
    //         item.forEach((items)=>{
    //             ActualWarningData.value.push(items)
    //         })
    //     }
    //    })
    //    ActualWarningData.value.forEach((item)=>{
    //     item.alarm_time = newDate(item.alarm_time)
    //     propose.forEach((value)=>{
    //         if(value.name == item.name){
    //             item.precautions = value.title
    //         }
    //     })
    //    })
    // })
    getStatusAlert(route.query.trainNo).then((res)=>{
        console.log(res);
        StateWarningData.value = []
        Object.entries(res).forEach(([key, item]) => {
        console.log(`Key: ${key}, Value:`, item);
        if(item.length !== 0){
                item.forEach((items)=>{
                    StateWarningData.value.push({...items, code: key})
                })
            }
            });
        // 原始  开始
    //     Object.values(res).forEach((item)=>{
    //         console.log(item);
    //     if(item.length !== 0){
    //         item.forEach((items)=>{
    //             StateWarningData.value.push(items)
    //         })
    //     }
    //    })
        // 原始  结束

    //    StateWarningData.value = res

       console.log(StateWarningData.value.length, '99999999999900');
       const proposeAdvice = propose.concat(proposeWarn)
       StateWarningData.value.forEach((item)=>{
        item.warning_time = newDate(item.warning_time)
        proposeAdvice.forEach((proposeValue)=>{
            if(proposeValue.value == item.code){
                item.precautions = proposeValue.title
                item.name = proposeValue.name
            }
        })
       })
       console.log(StateWarningData.value, 'StateWarningData.value');
        // StateWarningData.value = res.vw_statusalert_list
    })
    const list = []
    
     listData()

    getRunningState(route.query.trainNo).then((res)=>{
        RunStateData.value = []
        let data = {}
        let data5 = {}
        let data1 = {}
        let data2 = {}
        let data3 = {}
        let data4 = {}
        let key = 'comp_u11'
        let keys = 'comp_u12'
        res.vw_running_state_info.forEach((item,index)=>{
         var itemArray =['ef_u1','ef_u2','cf_u1','cf_u2','fad_u1','fad_u2']
            Object.keys(item).forEach((items)=>{
                if(items == 'cfbk_comp_u11'){
                data.equipmentName = '压缩机1'
                data[key+index] = runningState(item.cfbk_comp_u11)
                } else if(items == 'cfbk_comp_u21') {
                data[keys+index] = runningState(item.cfbk_comp_u21)
                }
                if(items == 'cfbk_comp_u12'){
                data5.equipmentName = '压缩机2'
                data5[key+index] = runningState(item.cfbk_comp_u12)
                } else if(items == 'cfbk_comp_u22') {
                data5[keys+index] = runningState(item.cfbk_comp_u22)
                }
                // if(items == 'comp_u11' || items == 'comp_u12'){
                // data.equipmentName = '压缩机'
                // if((item.comp_u11+'/'+item.comp_u12) == '0/0'){
                //     data[key+index] = '停机'
                // } else {
                //     data[key+index] = item.comp_u11+'/'+item.comp_u12
                // }
                // } else if(items == 'comp_u21'||items == 'comp_u22') {
                //     if((item.comp_u21+'/'+item.comp_u22) == '0/0'){
                //         data[keys+index] = '停机'
                //     } else {
                //         data[keys+index] = item.comp_u21+'/'+item.comp_u22
                //     }
                // }
                if(items == 'cfbk_ef_u11'){
                data1.equipmentName = '通风机'
                data1[key+index] = runningState(item.cfbk_ef_u11)
                } else if(items == 'cfbk_ef_u21') {
                data1[keys+index] = runningState(item.cfbk_ef_u21)
                }

                if(items == 'cfbk_cf_u11'){
                data2.equipmentName = '冷凝风机'
                data2[key+index] = runningState(item.cfbk_cf_u11)
                } else if(items == 'cfbk_cf_u21') {
                 data2[keys+index] = runningState(item.cfbk_cf_u21)
                }

                if(items == 'fadpos_u1'){
                data3.equipmentName = '新风阀开度'
                data3[key+index] = item.fadpos_u1 + '%'
                } else if(items == 'fadpos_u2') {
                 data3[keys+index] = item.fadpos_u2 +'%'
                }

                if(items == 'radpos_u1'){
                data4.equipmentName = '回风阀开度'
                data4[key+index] = item.radpos_u1+ '%'
                } else if(items == 'radpos_u2') {
                 data4[keys+index] = item.radpos_u2+ '%'
                }
            })
        })
        console.log(data)
      
        RunStateData.value.push(data,data5,data1,data2,data3,data4)
        console.log(RunStateData.value);
        
    })
    getTrainTemperature(route.query.trainNo).then((res)=>{
        // CarTemperatureData.value = res.vw_traintemperature
        let Tc1 = {}
        let Tc2 = {}
        let Tc3 = {}
        CarTemperatureData.value = []
        res.vw_traintemperature.forEach((item,index)=>{
            if(index == 0){
                Tc1.name = '车厢温度',
                // Tc1.carriageTC1 = (item.ras_sys*0.1).toFixed(2)+'℃'
                Tc1.carriageTC1 = item.ras_sys+'℃'
                Tc2.name = '目标温度',
                // Tc2.carriageTC1 = (item.tic*0.1).toFixed(2)+'℃'
                Tc2.carriageTC1 = item.tic+'℃'
                Tc3.name = '环境温度',
                // Tc3.carriageTC1 = (item.fas_sys*0.1).toFixed(2)+'℃'
                Tc3.carriageTC1 = item.fas_sys+'℃'
            } else if(index == 1) {
                // Tc1.carriageMP1 = (item.ras_sys*0.1).toFixed(2)+'℃'
                // Tc2.carriageMP1 = (item.tic*0.1).toFixed(2)+'℃'
                // Tc3.carriageMP1 = (item.fas_sys*0.1).toFixed(2)+'℃'
                Tc1.carriageMP1 = item.ras_sys+'℃'
                Tc2.carriageMP1 = item.tic+'℃'
                Tc3.carriageMP1 = item.fas_sys+'℃'
            } else if(index == 2){
                // Tc1.carriageM1 = (item.ras_sys*0.1).toFixed(2)+'℃'
                // Tc2.carriageM1 = (item.tic*0.1).toFixed(2)+'℃'
                // Tc3.carriageM1 = (item.fas_sys*0.1).toFixed(2)+'℃'
                Tc1.carriageM1 = item.ras_sys+'℃'
                Tc2.carriageM1 = item.tic+'℃'
                Tc3.carriageM1 = item.fas_sys+'℃'
            } else if(index == 3){
                // Tc1.carriageM2 = (item.ras_sys*0.1).toFixed(2)+'℃'
                // Tc2.carriageM2 = (item.tic*0.1).toFixed(2)+'℃'
                // Tc3.carriageM2 = (item.fas_sys*0.1).toFixed(2)+'℃'
                Tc1.carriageM2 = item.ras_sys+'℃'
                Tc2.carriageM2 = item.tic+'℃'
                Tc3.carriageM2 = item.fas_sys+'℃'
            } else if(index == 4){
                // Tc1.carriageMP2 = (item.ras_sys*0.1).toFixed(2)+'℃'
                // Tc2.carriageMP2 = (item.tic*0.1).toFixed(2)+'℃'
                // Tc3.carriageMP2 = (item.fas_sys*0.1).toFixed(2)+'℃'
                Tc1.carriageMP2 = item.ras_sys+'℃'
                Tc2.carriageMP2 = item.tic+'℃'
                Tc3.carriageMP2 = item.fas_sys+'℃'
            } else if(index == 5){
                // Tc1.carriageTC2 = (item.ras_sys*0.1).toFixed(2)+'℃'
                // Tc2.carriageTC2 = (item.tic*0.1).toFixed(2)+'℃'
                // Tc3.carriageTC2 = (item.fas_sys*0.1).toFixed(2)+'℃'
                Tc1.carriageTC2 = item.ras_sys+'℃'
                Tc2.carriageTC2 = item.tic+'℃'
                Tc3.carriageTC2 = item.fas_sys+'℃'
            }
        })
         CarTemperatureData.value.push(Tc1,Tc2,Tc3)
    })
}
let time = setInterval(() => {
   getTrainApi()
   console.log('请求')
  }, 1000*60*2);
onMounted(() => {
    getTrainApi()
})
onUnmounted(()=>{
    clearInterval(time)
})
</script>

<style scoped lang="scss">
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
    background-color: #181F30 !important;
    box-shadow: 0 0 1px 0 #394153;
    color: #fff !important;
    height: 40px;
    box-sizing: border-box;
    font-size: 110%;
}
.header {
    height: 60px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 5px;
}

.warp {
    box-sizing: border-box;
    width: 100%;
    border-radius: 10px;
    background-color: #181F30 !important;
    border: 1px solid #394153;
    margin-bottom: 15px;
}
.btn_text {
    font-size: 12px;
    color: #2186CF !important;
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
    background: #181F30;
    border-radius: 20px;
    border: 1px solid #3a3b68;
    padding: 5px 20px 20px;
}
:deep(.el-table .el-table__cell) {
    padding: 10px !important;
}
:deep(.el-input__wrapper) {
    background-color: #181F30 !important;
    border: 1px solid #394153;
    box-shadow: 0 0 1px 0 #394153;
    color: #fff !important;
    .el-input__inner {
        border: none !important;
        color: #fff !important;
    }
}
</style>


