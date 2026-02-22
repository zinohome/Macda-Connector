<template>
    <div class="warp">
        <!-- <dv-full-screen-container> </dv-full-screen-container> -->
        <!-- <el-row
            v-loading="loading"
            element-loading-spinner="el-icon-loading"
            element-loading-text="数据加载中..."
            element-loading-background="rgba(12,17,36, 0.8)"
        > -->
        <el-row :gutter="10">
            <el-col>
                <el-row :gutter="10">
                    <el-col :span="6">
                        <Left :tableData="leftData" />
                    </el-col>
                    <el-col :span="12">
                        <Center :centerData="centerData " />
                    </el-col>
                    <el-col :span="6">
                        <Right :tableDataEa="rightDataEA" :tableDataPa="rightDataPA" />
                    </el-col>
                </el-row>
            </el-col>
            <el-col :span="24">
                        <AirErrorData />
                    </el-col>
        </el-row>
    </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import Left from "./Left.vue";
import Right from "./Right.vue";
import AirErrorData from "./AirErrorData.vue";
import Center from "./Center.vue";
import { getActiveCarApi, getAirSystemApi, getRealtimeAlarm, getRealtimeWarning } from '@/api/api'

let loading = ref(true)
let leftData = ref([])
let rightDataEA = ref([])
let rightDataPA = ref([])
let centerData = ref({})
let num = ref(0)
let count = ref(0)
let normal = ref(0)
let trainNo = ref([])
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
function getApi(){
    // getActiveCarApi().then((res)=>{
    //  centerData.value = res.vw_train_count_info[0]
    // })
    getAirSystemApi().then((res)=>{
       leftData.value = res.vw_train_alarm_count
       centerData.value.totalnumber = res.vw_train_alarm_count.length
       num.value = 0
       count.value = 0
       normal.value = 0
       trainNo.value = []
       res.vw_train_alarm_count.forEach((item)=>{
        console.log(item,(item.warning_count > 0) && (item.warning_count !== null));
        if((item.alarm_count > 0) && (item.alarm_count !== null)){
            num.value++
            trainNo.value.push(item.train_no)
        }
        if((item.warning_count > 0) && (item.warning_count !== null)){
            // console.log(item.warning_count,'333333');
            count.value++
        }
        console.log(((item.warning_count == 0) || (item.warning_count == null)) && ((item.alarm_count == 0) || (item.alarm_count == null)));
        if(((item.warning_count == 0) || (item.warning_count == null)) && ((item.alarm_count == 0) || (item.alarm_count == null))){
            normal.value++
        }
       })
    //    console.log(count.value);
       centerData.value.warning_count = count.value
       centerData.value.alarm = num.value
       centerData.value.normal = normal.value
       rightDataEA.value = []
       getRealtimeAlarm({trainNo: trainNo.value}).then((res)=>{
        res.vw_train_alarm_info.forEach((item)=>{
            Object.keys(item).forEach((value)=>{
                if(item[value]>0 && item[value +'_name']){
                    rightDataEA.value.push({
                        name:item[value +'_name'],
                        code: value,
                        alarm_time: item.alarm_time,
                        carriage_no: item.carriage_no
                    })
                }
            })
        })
        const proposeAdvice = propose.concat(proposeWarn)
        rightDataEA.value.forEach((values)=>{
            values.alarm_time = newDate(values.alarm_time)
            proposeAdvice.forEach((proposeValue)=>{
                if(proposeValue.value == values.code){
                    values.precautions = proposeValue.title
                    values.name = proposeValue.name
                }
            })
        })
    })
    // getRealtimeAlarm().then((res)=>{
    //     rightDataEA.value = []
    //    Object.values(res).forEach((item)=>{
    //     if(item.length !== 0){
    //         item.forEach((items)=>{
    //            propose.forEach((value)=>{
    //             if(value.name == items.name){
    //                 items.precautions = value.title
    //             }
    //            })
    //             rightDataEA.value.push(items)
    //         })
    //     }
    //    })
    //    rightDataEA.value.forEach((value)=>{
    //     value.alarm_time = newDate(value.alarm_time)
    //    })
    // //    console.log(rightDataEA.value);
    // })
    getRealtimeWarning().then((res)=>{
        rightDataPA.value = []
        Object.entries(res).forEach(([key, itemsObj]) => {
           if(Array.isArray(itemsObj) && itemsObj.length !== 0){
                itemsObj.forEach((items)=>{
                    rightDataPA.value.push({...items, code: key})
                })
            } else if (typeof itemsObj === 'object' && itemsObj !== null) {
                 // in case it's grouped via dict
                 Object.values(itemsObj).forEach((itemArr) => {
                     if(Array.isArray(itemArr)) {
                         itemArr.forEach(i => rightDataPA.value.push({...i, code: key}))
                     }
                 })
            }
        });
        
        const proposeAdvice = propose.concat(proposeWarn)
        rightDataPA.value.forEach((value)=>{
            value.warning_time = newDate(value.warning_time)
            proposeAdvice.forEach((proposeValue)=>{
                if(proposeValue.value == value.code){
                    value.precautions = proposeValue.title
                    value.name = proposeValue.name
                }
            })
        })
        console.log("Warnings PA: ", rightDataPA.value);
    })
    })
}
onMounted(() => {
    // initWebSocket()
    getApi()
})
onUnmounted(()=>{
    clearInterval(time)
})
let time = setInterval(() => {
   getApi()
   console.log('请求')
  }, 1000*60*2);
function initWebSocket() { //初始化weosocket
    const uuid = new Date().getTime() + ''
    const wsuri = "ws://" + location.host + "/ws/" + uuid;
    let websock = new WebSocket(wsuri);
    websock.onmessage = websocketonmessage;
    websock.onopen = websocketonopen;
    websock.onerror = websocketonerror;
    websock.onclose = websocketclose;
}

function websocketonopen() { //连接建立之后执行
    loading.value = false
}

function websocketonerror() { //连接建立失败重连
    // initWebSocket();
}

function websocketonmessage(e) { //数据接收
    const data = JSON.parse(e.data);
    leftData.value = data.ActiveCar;
    rightDataEA.value = data.eaList;
    rightDataPA.value = data.paList;
    centerData.value = data.dataNums;
}

function websocketclose(e) {  //关闭
    console.log('断开连接', e);
}
</script>
<style scoped>
/* :deep(thead) {
    border-bottom: 2px solid red;
} */
/* :deep(.el-table__cell) {
    border-bottom: 2px solid transparent;
} */
/* :deep(.el-table .cell) {
    padding: 0 4px;
} */
/* nodate table style */
:deep(.el-table__empty-text) {
   margin-top: 40px;
   line-height: 20px !important;
}
</style>



