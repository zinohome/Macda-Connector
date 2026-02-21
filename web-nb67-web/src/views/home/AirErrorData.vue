<template>
  <div class="warp">
    <TitleHeader title="空调故障统计" @setEchart="setEchart()"/>
    <div ref="echart" class="echartDom"></div>
  </div>
</template>

<script setup>
// import { getEaDataApi } from "@/api/api.js";
import { ref, onMounted, provide, onUnmounted } from "vue";
import * as echarts from "echarts";
import TitleHeader from "@/components/TitleHeader.vue";
import { getFaultStatistics } from '@/api/api'
const error = [
    {
     value: 'bflt_tempover',
     name: '车厢温度超温预警',
     classify: '轻微',
     title: `1.检查2个机组的制冷系统是否异常`,
    },
    {
     value: 'bocflt_ef_u11',
     name: '通风机过流故障U1-1',
     classify: '严重',
     title: `1.首先检查电机保护器是否跳闸，如果没有跳闸那就是反馈回路断开触发的故障
             \n2.手动开合电机保护器，通过PTU或本车厢触摸屏查看输入信号状态，检查反馈回路是否正常
             \n3.跳闸引起的原因需要检查三相电路是否缺相或有虚接情况
             \n4.检查电机自身三相电阻是否平衡，用绝缘测试仪测试绝缘电阻是否大于5MΩ
             \n5.检查电机保护器主触点之间是否正常导通，电阻小于20Ω。如有异常，更换新物料`
    },{
     value: 'bocflt_ef_u12',
     name: '通风机过流故障U1-2',
     classify: '严重',
     title: `1.首先检查电机保护器是否跳闸，如果没有跳闸那就是反馈回路断开触发的故障
             \n2.手动开合电机保护器，通过PTU或本车厢触摸屏查看输入信号状态，检查反馈回路是否正常
             \n3.跳闸引起的原因需要检查三相电路是否缺相或有虚接情况
             \n4.检查电机自身三相电阻是否平衡，用绝缘测试仪测试绝缘电阻是否大于5MΩ
             \n5.检查电机保护器主触点之间是否正常导通，电阻小于20Ω。如有异常，更换新物料`
    },{
     value: 'bocflt_cf_u11',
     name: '冷凝风机过流故障U1-1',
     classify: '中等',
     title: `1.首先检查电机保护器是否跳闸，如果没有跳闸那就是反馈回路断开触发的故障
             \n2.手动开合电机保护器，通过PTU或本车厢触摸屏查看输入信号状态，检查反馈回路是否正常
             \n3.跳闸引起的原因需要检查三相电路是否缺相或有虚接情况
             \n4.检查电机自身三相电阻是否平衡，用绝缘测试仪测试绝缘电阻是否大于5MΩ
             \n5.检查电机保护器主触点之间是否正常导通，电阻小于20Ω。如有异常，更换新物料`
    },{
     value: 'bocflt_cf_u12',
     name: '冷凝风机过流故障U1-2',
     classify: '中等',
     title: `1.首先检查电机保护器是否跳闸，如果没有跳闸那就是反馈回路断开触发的故障
             \n2.手动开合电机保护器，通过PTU或本车厢触摸屏查看输入信号状态，检查反馈回路是否正常
             \n3.跳闸引起的原因需要检查三相电路是否缺相或有虚接情况
             \n4.检查电机自身三相电阻是否平衡，用绝缘测试仪测试绝缘电阻是否大于5MΩ
             \n5.检查电机保护器主触点之间是否正常导通，电阻小于20Ω。如有异常，更换新物料`
    },{
     value: 'bflt_vfd_u11',
     name: '变频器保护故障U1-1',
     classify: '中等',
     title: `1.检查变频器散热风扇是否正常工作
             \n2.检查变频器和压缩机之间连线是否有虚接或缺相的情况
             \n3.检查变频器的反馈回路各接点接线是否有虚接的情况
             \n4.检查上述问题后，借助PTU顺序启动制冷运行，记录压缩机运行电流，和同车厢未故障的压缩机运行电流对比，差值不可超过20%
             \n5.更换新变频器再测试压缩机运行电流，比对查找异常点`
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
     name: '变频器保护故障U1-2',
     classify: '中等',
     title: `1.检查变频器散热风扇是否正常工作
            \n2.检查变频器和压缩机之间连线是否有虚接或缺相的情况
            \n3.检查变频器的反馈回路各接点接线是否有虚接的情况
            \n4.检查上述问题后，借助PTU顺序启动制冷运行，记录压缩机运行电流，和同车厢未故障的压缩机运行电流对比，差值不可超过20%
            \n5.更换新变频器再测试压缩机运行电流，比对查找异常点`
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
     name: '低压故障U1-1',
     classify: '中等',
     title: `1.使用卤素检漏仪在机组内检查是否有冷媒泄露
            \n2.通过本车厢触摸屏或PTU查看通风机是否正常开启
            \n3.检查风阀是否可以正常打开
            \n4.检查滤网是否脏堵
            \n5.使用PTU查看电子膨胀阀相关数据，吸气温度、吸气压力、过热度是否正常，判断电子膨胀阀是否可以正常打开`
    },{
     value: 'bscflt_comp_u11',
     name: '高压故障U1-1',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.用压力表查看高压侧实际压力是否和PTU上显示的值一致，判断高压传感器是否异常
            \n4.使用PTU顺序启动制冷，测试压缩机运转电流，和同车厢未异常系统的压缩机电流比对，不可超过20%`
    },{
     value: 'bscflt_vent_u11',
     name: '排气故障U1-1',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.检查空调四周散热是否顺畅`
    },{
     value: 'blpflt_comp_u12',
     name: '低压故障U1-2',
     classify: '中等',
     title: `1.使用卤素检漏仪在机组内检查是否有冷媒泄露
            \n2.通过本车厢触摸屏或PTU查看通风机是否正常开启
            \n3.检查风阀是否可以正常打开
            \n4.检查滤网是否脏堵
            \n5.使用PTU查看电子膨胀阀相关数据，吸气温度、吸气压力、过热度是否正常，判断电子膨胀阀是否可以正常打开`
    },{
     value: 'bscflt_comp_u12',
     name: '高压故障U1-2',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.用压力表查看高压侧实际压力是否和PTU上显示的值一致，判断高压传感器是否异常
            \n4.使用PTU顺序启动制冷，测试压缩机运转电流，和同车厢未异常系统的压缩机电流比对，不可超过20%`
    },{
     value: 'bscflt_vent_u12',
     name: '排气故障U1-2',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.检查空调四周散热是否顺畅`
    },{
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
     name: '扩展模块通讯故障U1',
     classify: '中等',
     title: `1.检查扩展模块供电是否正常，正常供电电压DC24V
            \n2.检查通讯电缆接头是否松动，重新紧固
            \n3.检查通讯电缆各接点位置是否虚接
            \n4.更换扩展模块，重新连接测试`
    },{
     value: 'bflt_frstemp_u1',
     name: '新风温度传感器故障U1',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_rnttemp_u1',
     name: '回风温度传感器故障U1',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_splytemp_u11',
     name: '送风温度传感器故障U1-1',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },
    {
     value: 'bflt_splytemp_u12',
     name: '送风温度传感器故障U1-2',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
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
     name: '通风机过流故障U2-1',
     classify: '严重',
     title: `1.首先检查电机保护器是否跳闸，如果没有跳闸那就是反馈回路断开触发的故障
            \n2.手动开合电机保护器，通过PTU或本车厢触摸屏查看输入信号状态，检查反馈回路是否正常
            \n3.跳闸引起的原因需要检查三相电路是否缺相或有虚接情况
            \n4.检查电机自身三相电阻是否平衡，用绝缘测试仪测试绝缘电阻是否大于5MΩ
            \n5.检查电机保护器主触点之间是否正常导通，电阻小于20Ω。如有异常，更换新物料`
    },{
     value: 'bocflt_ef_u22',
     name: '通风机过流故障U2-2',
     classify: '严重',
     title: `1.首先检查电机保护器是否跳闸，如果没有跳闸那就是反馈回路断开触发的故障
            \n2.手动开合电机保护器，通过PTU或本车厢触摸屏查看输入信号状态，检查反馈回路是否正常
            \n3.跳闸引起的原因需要检查三相电路是否缺相或有虚接情况
            \n4.检查电机自身三相电阻是否平衡，用绝缘测试仪测试绝缘电阻是否大于5MΩ
            \n5.检查电机保护器主触点之间是否正常导通，电阻小于20Ω。如有异常，更换新物料`
    },{
     value: 'bocflt_cf_u21',
     name: '冷凝风机过流故障U2-1',
     classify: '中等',
     title: `1.首先检查电机保护器是否跳闸，如果没有跳闸那就是反馈回路断开触发的故障
            \n2.手动开合电机保护器，通过PTU或本车厢触摸屏查看输入信号状态，检查反馈回路是否正常
            \n3.跳闸引起的原因需要检查三相电路是否缺相或有虚接情况
            \n4.检查电机自身三相电阻是否平衡，用绝缘测试仪测试绝缘电阻是否大于5MΩ
            \n5.检查电机保护器主触点之间是否正常导通，电阻小于20Ω。如有异常，更换新物料`
    },{
     value: 'bocflt_cf_u22',
     name: '冷凝风机过流故障U2-2',
     classify: '中等',
     title: `1.首先检查电机保护器是否跳闸，如果没有跳闸那就是反馈回路断开触发的故障
            \n2.手动开合电机保护器，通过PTU或本车厢触摸屏查看输入信号状态，检查反馈回路是否正常
            \n3.跳闸引起的原因需要检查三相电路是否缺相或有虚接情况
            \n4.检查电机自身三相电阻是否平衡，用绝缘测试仪测试绝缘电阻是否大于5MΩ
            \n5.检查电机保护器主触点之间是否正常导通，电阻小于20Ω。如有异常，更换新物料`
    },{
     value: 'bflt_vfd_u21',
     name: '变频器保护故障U2-1',
     classify: '中等',
     title: `1.检查变频器散热风扇是否正常工作
            \n2.检查变频器和压缩机之间连线是否有虚接或缺相的情况
            \n3.检查变频器的反馈回路各接点接线是否有虚接的情况
            \n4.检查上述问题后，借助PTU顺序启动制冷运行，记录压缩机运行电流，和同车厢未故障的压缩机运行电流对比，差值不可超过20%
            \n5.更换新变频器再测试压缩机运行电流，比对查找异常点`
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
     name: '变频器保护故障U2-2',
     classify: '中等',
     title: `1.检查变频器散热风扇是否正常工作
            \n2.检查变频器和压缩机之间连线是否有虚接或缺相的情况
            \n3.检查变频器的反馈回路各接点接线是否有虚接的情况
            \n4.检查上述问题后，借助PTU顺序启动制冷运行，记录压缩机运行电流，和同车厢未故障的压缩机运行电流对比，差值不可超过20%
            \n5.更换新变频器再测试压缩机运行电流，比对查找异常点`
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
     name: '低压故障U2-1',
     classify: '中等',
     title: `1.使用卤素检漏仪在机组内检查是否有冷媒泄露
            \n2.通过本车厢触摸屏或PTU查看通风机是否正常开启
            \n3.检查风阀是否可以正常打开
            \n4.检查滤网是否脏堵
            \n5.使用PTU查看电子膨胀阀相关数据，吸气温度、吸气压力、过热度是否正常，判断电子膨胀阀是否可以正常打开`
    },{
     value: 'bscflt_comp_u21',
     name: '高压故障U2-1',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.用压力表查看高压侧实际压力是否和PTU上显示的值一致，判断高压传感器是否异常
            \n4.使用PTU顺序启动制冷，测试压缩机运转电流，和同车厢未异常系统的压缩机电流比对，不可超过20%`
    },{
     value: 'bscflt_vent_u21',
     name: '排气故障U2-1',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.检查空调四周散热是否顺畅`
    },{
     value: 'blpflt_comp_u22',
     name: '低压故障U2-2',
     classify: '中等',
     title: `1.使用卤素检漏仪在机组内检查是否有冷媒泄露
            \n2.通过本车厢触摸屏或PTU查看通风机是否正常开启
            \n3.检查风阀是否可以正常打开
            \n4.检查滤网是否脏堵
            \n5.使用PTU查看电子膨胀阀相关数据，吸气温度、吸气压力、过热度是否正常，判断电子膨胀阀是否可以正常打开`
    },{
     value: 'bscflt_comp_u22',
     name: '高压故障U2-2',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.用压力表查看高压侧实际压力是否和PTU上显示的值一致，判断高压传感器是否异常
            \n4.使用PTU顺序启动制冷，测试压缩机运转电流，和同车厢未异常系统的压缩机电流比对，不可超过20%`
    },{
     value: 'bscflt_vent_u22',
     name: '排气故障U2-2',
     classify: '中等',
     title: `1.检查冷凝风机是否正常运转，转向是否正确
            \n2.检查冷凝器是否有大面积异物遮挡或脏堵是否很严重
            \n3.检查空调四周散热是否顺畅`
    },{
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
     name: '扩展模块通讯故障U2',
     classify: '中等',
     title: `1.检查扩展模块供电是否正常，正常供电电压DC24V
            \n2.检查通讯电缆接头是否松动，重新紧固
            \n3.检查通讯电缆各接点位置是否虚接
            \n4.更换扩展模块，重新连接测试`
    },{
     value: 'bflt_frstemp_u2',
     name: '新风温度传感器故障U2',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_rnttemp_u2',
     name: '回风温度传感器故障U2',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_splytemp_u21',
     name: '送风温度传感器故障U2-1',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
    },{
     value: 'bflt_splytemp_u22',
     name: '送风温度传感器故障U2-2',
     classify: '轻微',
     title: `1.拔掉传感器插头，直接测量传感器电阻阻值，对照温度-阻值表，计算温度值和环境温度比对
            \n2.阻值无异常的情况下，检查传感器到扩展模块之间的线路是否有短路或虚接的情况`
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
     title: `1.通过模拟三相电源故障，启动紧急通风模式，观察紧急逆变器是否正常运行
            \n2.在紧急通风模式下，通过PTU或本车厢触摸屏上查看正常工作信号是否导通，故障反馈信号是否导通
            \n3.如果是逆变器自身报故障，需要更换逆变器
            \n4.如果逆变器可以正常工作，那就需要检查正常工作反馈信号和故障反馈信号回路是否有虚接或断开的情况存在，重新紧固`
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
    }
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
// const error = [
//  {
//   value: 'dvc_bocflt_ef_u11',
//   title: '通风机1-1过流故障'
//  },
//  {
//   value: 'dvc_bocflt_ef_u12',
//   title: '通风机1-2过流故障'
//  },{
//   value: 'dvc_bocflt_cf_u12',
//   title: '冷凝风机1-2过流故障'
//  },{
//   value: 'dvc_bflt_vfd_u11',
//   title: '变频器1-1故障'
//  },{
//   value: 'dvc_bscflt_comp_u11',
//   title: '压缩机1-1高压连锁故障'
//  },{
//   value: 'dvc_bflt_eev_u11',
//   title: '电子膨胀阀1-1故障'
//  },{
//   value: 'dvc_blpflt_comp_u11',
//   title: '压缩机1-1低压故障'
//  },{
//   value: 'dvc_bscflt_comp_u11',
//   title: '压缩机1-1高压连锁故障'
//  },{
//   value: 'dvc_bflt_vfd_u12',
//   title: '变频器1-2故障'
//  },{
//   value: 'dvc_blpflt_comp_u12',
//   title: '压缩机1-2低压故障'
//  },{
//   value: 'dvc_bflt_eev_u11',
//   title: '电子膨胀阀1-1故障'
//  },{
//   value: 'dvc_bflt_fad_u1',
//   title: '新风阀U1故障'
//  },{
//   value: 'dvc_bflt_rad_u1',
//   title: '回风阀U1故障'
//  },{
//   value: 'dvc_bflt_airclean_u1',
//   title: '空气净化U1故障'
//  },{
//   value: 'dvc_bflt_frstemp_u1',
//   title: '新风温度传感器U1故障'
//  },{
//   value: 'dvc_bflt_splytemp_u11',
//   title: '送风温度传感器1-1故障'
//  },{
//   value: 'dvc_bflt_splytemp_u12',
//   title: '送风温度传感器1-2故障'
//  },{
//   value: 'dvc_bflt_rnttemp_u1',
//   title: '回风温度传感器U1故障'
//  },{
//   value: 'dvc_bflt_dfstemp_u11',
//   title: '融霜温度传感器1-1故障'
//  },{
//   value: 'dvc_bflt_dfstemp_u12',
//   title: '融霜温度传感器1-2故障'
//  },{
//   value: 'dvc_bocflt_ef_u21',
//   title: '通风机2-1过流故障'
//  },{
//   value: 'dvc_bocflt_ef_u22',
//   title: '通风机2-2过流故障'
//  },{
//   value: 'dvc_bocflt_cf_u21',
//   title: '冷凝风机2-1过流故障'
//  },{
//   value: 'dvc_bocflt_cf_u22',
//   title: '冷凝风机2-2过流故障'
//  },{
//   value: 'dvc_bflt_vfd_u21',
//   title: '变频器2-1故障'
//  },{
//   value: 'dvc_blpflt_comp_u21',
//   title: '压缩机2-1低压故障'
//  },{
//   value: 'dvc_bscflt_comp_u21',
//   title: '压缩机2-1高压连锁故障'
//  },{
//   value: 'dvc_bflt_vfd_u22',
//   title: '变频器2-2故障'
//  },{
//   value: 'dvc_blpflt_comp_u22',
//   title: '压缩机2-2低压故障'
//  },{
//   value: 'dvc_bscflt_comp_u22',
//   title: '压缩机2-2高压连锁故障'
//  },{
//   value: 'dvc_bflt_eev_u21',
//   title: '电子膨胀阀2-1故障'
//  },{
//   value: 'dvc_bflt_eev_u22',
//   title: '电子膨胀阀2-2故障'
//  },{
//   value: 'dvc_bflt_fad_u2',
//   title: '新风阀U2故障'
//  },{
//   value: 'dvc_bflt_rad_u2',
//   title: '回风阀U2故障'
//  },{
//   value: 'dvc_bflt_airclean_u2',
//   title: '空气净化U2故障'
//  },{
//   value: 'dvc_bflt_frstemp_u2',
//   title: '新风温度传感器U2故障'
//  },{
//   value: 'dvc_bflt_splytemp_u21',
//   title: '送风温度传感器2-1故障'
//  },{
//   value: 'dvc_bflt_splytemp_u22',
//   title: '送风温度传感器2-2故障'
//  },
//  {
//   value: 'dvc_bflt_rnttemp_u2',
//   title: '回风温度传感器U2故障'
//  },{
//   value: 'dvc_bflt_dfstemp_u21',
//   title: '融霜温度传感器2-1故障'
//  },{
//   value: 'dvc_bflt_dfstemp_u22',
//   title: '融霜温度传感器2-2故障'
//  },{
//   value: 'dvc_bflt_vehtemp',
//   title: '车厢温度传感器故障'
//  },{
//   value: 'dvc_bflt_seattemp',
//   title: '座椅下方温度传感器故障'
//  },{
//   value: 'dvc_bflt_emergivt',
//   title: '紧急逆变器故障'
//  },{
//   value: 'dvc_bflt_mvbbus',
//   title: 'MVB 通讯故障'
//  },{
//   value: 'dvc_bcomuflt_vfd_u11',
//   title: '变频器1-1通讯故障'
//  },{
//   value: 'dvc_bcomuflt_vfd_u12',
//   title: '变频器1-2通讯故障'
//  },{
//   value: 'dvc_bcomuflt_vfd_u21',
//   title: '变频器2-1通讯故障'
//  },{
//   value: 'dvc_bcomuflt_vfd_u22',
//   title: '变频器2-2通讯故障'
//  },{
//   value: 'dvc_bcomuflt_eev_u11',
//   title: '电子膨胀阀1-1通讯故障'
//  },{
//   value: 'dvc_bcomuflt_eev_u12',
//   title: '电子膨胀阀1-2通讯故障'
//  },{
//   value: 'dvc_bcomuflt_eev_u21',
//   title: '电子膨胀阀2-1通讯故障'
//  },{
//   value: 'dvc_bcomuflt_eev_u22',
//   title: '电子膨胀阀2-2通讯故障'
//  },{
//   value: 'dvc_bmcbflt_pwr_u1',
//   title: '机组1供电故障'
//  },{
//   value: 'dvc_bmcbflt_pwr_u2',
//   title: '机组2供电故障'
//  },{
//   value: 'dvc_blplockflt_u11',
//   title: '低压1-1锁死故障'
//  },{
//   value: 'dvc_blplockflt_u12',
//   title: '低压1-2锁死故障'
//  },{
//   value: 'dvc_blplockflt_u21',
//   title: '低压2-1锁死故障'
//  },{
//   value: 'dvc_blplockflt_u22',
//   title: '低压2-2锁死故障'
//  },{
//   value: 'dvc_bsclockflt_u11',
//   title: '高压连锁1-1锁死故障'
//  },{
//   value: 'dvc_bsclockflt_u12',
//   title: '高压连锁1-2锁死故障'
//  },{
//   value: 'dvc_bsclockflt_u21',
//   title: '高压连锁2-1锁死故障'
//  },{
//   value: 'dvc_bsclockflt_u22',
//   title: '高压连锁2-2锁死故障'
//  },{
//   value: 'dvc_bvfdlockflt_u11',
//   title: '变频器1-1锁死故障'
//  },{
//   value: 'dvc_bvfdlockflt_u12',
//   title: '变频器1-2锁死故障'
//  },{
//   value: 'dvc_bvfdlockflt_u21',
//   title: '变频器2-1锁死故障'
//  },{
//   value: 'dvc_bvfdlockflt_u22',
//   title: '变频器2-2锁死故障'
//  },{
//   value: 'dvc_beevlockflt_u11',
//   title: '电子膨胀阀1-1锁死故障'
//  },{
//   value: 'dvc_beevlockflt_u12',
//   title: '电子膨胀阀1-2锁死故障'
//  },{
//   value: 'dvc_beevlockflt_u21',
//   title: '电子膨胀阀2-1锁死故障'
//  },{
//   value: 'dvc_beevlockflt_u22',
//   title: '电子膨胀阀2-2锁死故障'
//  }
// ]
const echart = ref()

const setEchart = (params) => {
  getFaultStatistics(params).then((res) => {
    let count = {};
    let total = 0
for (let i = 0; i < res.vw_alarm_info_all_date.length; i++) {
    for (let key in res.vw_alarm_info_all_date[i]) {
      console.log(res.vw_alarm_info_all_date[i][key]);
        if (res.vw_alarm_info_all_date[i][key] === 1) {
            if (count[key]) {
                count[key]++;
            } else {
                count[key] = 1;
            }
        }
    }
}
    console.log(count);
    total = res.vw_alarm_info_all_date.length
    let keysTitle = []
    let value = []
    let keysName = []
    for (let key in count) {
      count[key] = ((count[key] / total).toFixed(2))*100;
      // keysTitle.push(key);
      // value.push(count[key]);
    }
    for (let key in count) {
      value.push(count[key]);
     error.forEach((val)=>{
      if(key == val.value){
        console.log(key);
        key = val.name
        keysTitle.push(val.name);
        
        // count[key] = count[val.title];
      }
     })
    }
    if(keysTitle.length >8){
       keysTitle= keysTitle.slice(0,8)
       value = value.slice(0,8)
    }
    initEchart(keysTitle, value);
    console.log(count,keysTitle,value);
    // error.forEach((val)=>{
    //   keysTitle.forEach((item)=>{
    //     console.log(item == val.value);
    //     if(item == val.value){
    //       keysName.push(val.title)
    //     }
    //   })
    // })
    // console.log(count,keysTitle,keysName,value);
    // let keysName = []
    // let keysTitle = []
    // let value = []
    // let total = 0
    // Object.values(res).forEach((item)=>{
    //   total = res.train_info.length
    //   if((Object.keys(res) !== 'train_info') && (item.length !== 0)){
    //     if(Object.keys(item[0]).indexOf('train_no') ==-1){
    //     const num = (item.length / total).toFixed(2)
    //     // console.log(num*100);
    //     value.push(num*100)
    //     }
    //   }
    //     if(item.length !== 0){
    //       item.forEach((items)=>{
    //       const [keys, values] = initData(items);
    //       keysName.push(keys[0])
    //       keysName = Array.from(new Set(keysName))
    //       error.forEach((val)=>{
    //         keysName.forEach((values)=>{
    //           if(val.value == values){
    //             keysTitle.push(val.title)
    //             keysTitle = Array.from(new Set(keysTitle))
    //             // console.log(keysTitle)
    //           }
    //         })
    //       })
    //       if(keysTitle.length>8){
    //         // keysTitle= keysTitle.slice(0,1)
    //         // value = value.slice(0,1)
    //         keysTitle= keysTitle.slice(0,8)
    //         value = value.slice(0,8)
    //       }
    //       // console.log(value,keysTitle);
    //       // console.log(keysTitle, value);
    //       // if(keysTitle.length == 0){
    //       //   value = []
    //       // }
    //       initEchart(keysTitle, value);
    //         })
    //     } else {
    //        initEchart(keysTitle, value);
    //     }
    //    })
  });
};
let interval = undefined;
const chooseDay = (title, dateParams) => {
  setEchart(dateParams);
  if (interval != undefined) {
    // 清空定时器
    window.clearTimeout(interval);
  }
  // 开启定时任务
  // interval = setInterval(() => {
  //   setEchart(dateParams);
  // }, 5000);
};
provide("provideChooseDay", chooseDay);

const initEchart = (xAxisData, seriesData) => {
  // console.log(xAxisData,seriesData)
  let option1 = {
    color: ["#FC7B1E"],

    tooltip: {
      trigger: "axis",
      axisPointer: {
        type: "shadow",
      },
      formatter: function (params) {
        let relVal = params[0].name;
        for (let i = 0; i < params.length; i++) {
          // params[i].value = params[i].value.slice(0,-1)
          params[i].value = params[i].value+'%'
          relVal +=
            "<br/>" + params[i].marker + " 故障百分比: " + params[i].value;
        }
        return relVal;
      },
    },
    grid: {
      left: "6%",
      right: "6%",
      bottom: "3%",
      containLabel: true,
    },
    xAxis: {
      type: "category",
      data: xAxisData, // ['蒸发器A', '蒸发器B', '混合风滤网A'],
      axisTick: {
        alignWithLabel: true,
      },
      axisLabel: {
        interval: 0,
        textStyle: {
          color: "#fff",
        },
      },
    },
    yAxis: {
      show: true,
      min:0,
      max:100,
      type: "value",
      name: '单位（%）',
      splitLine: {
        show: true,
        lineStyle: {
          type: "dashed",
        },
      },
      axisLabel: {
        textStyle: {
          color: "#fff",
        },
      },
    },

    series: [
      {
        type: "bar",
        barWidth: "20%",
        data: seriesData,
      },
    ],
  };
  if(echart.value){
  let myChart1 = echarts.init(echart.value);
  myChart1.setOption(option1);
  window.onresize = function () {
    myChart1.resize();
  };
};
}
const initData = (data) => {
  if (typeof data == "object") {
    let keys = Object.keys(data);
    let values = Object.values(data);
    return [keys, values];
  }
};

onMounted(() => {
  // setEchart();
  if (interval != undefined) {
    // 清空定时器
    window.clearTimeout(interval);
  }
  // 开启定时任务
  // interval = setInterval(() => {
  //   setEchart();
  // }, 5000);
});

onUnmounted(() => {
  if (interval != undefined) {
    // 清空定时器
    window.clearTimeout(interval);
  }
});
</script>

<style scoped lang="scss">
$s-primary: #2186CF;
$s-danger: #e65355;
$s-warning: #ffa55c;
$s-info: #bfbfbf;
$s-success: #42ad5d;
$s-yellow: #ffe375;
$s-white: #ffffff;
$s-border: #394153;
:root {
    --el-color-primary: $s-primary;
}
.warp {
  width: 100%;
  border-radius: 10px;
  background-color: #181F30;
  border: 1px solid $s-border;
  height: 500px;
  /* margin-bottom: 10px; */
}
.echartDom {
  width: 100%;
  height: 400px;
  float: left;
}
:deep(.el-date-editor.el-date-editor--daterange.el-input__wrapper.el-range-editor.el-tooltip__trigger.el-tooltip__trigger) {
  background-color: #181F30 !important;
  border: 1px solid $s-border;
  box-shadow: 0 0 1px 0 $s-border;
  color: #fff !important;
}
</style>
