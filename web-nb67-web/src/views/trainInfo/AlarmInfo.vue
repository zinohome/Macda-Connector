<template>
    <div class="warp">
        <div class="header">
            <h4>报警信息</h4>
            <el-date-picker
                v-model="dateValue"
                type="daterange"
                range-separator="~"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                :clearable="false"
                @change="dateChange"
            />
        </div>
        <div class="content">
            <el-table :data="props.AlarmInfoData" stripe style="width: 100%" height="300px">
                <el-table-column prop="state" label="车厢号" min-width="20%"></el-table-column>
                <el-table-column prop="unit_no" label="机组号" min-width="20%"></el-table-column>
                <el-table-column prop="classify" label="故障类型" min-width="15%">
                    <template #default="scope">
                        <el-tag class="s-tag-yellow" effect="plain" round v-if="scope.row.classify.indexOf('轻微') !==-1">轻微</el-tag>
                        <el-tag effect="plain" round v-if="scope.row.classify.indexOf('正常') !==-1">正常</el-tag>
                        <el-tag type="warning" effect="plain" round v-if="scope.row.classify.indexOf('一般') !==-1">一般</el-tag>
                        <el-tag type="danger" effect="plain" round v-if="scope.row.classify.indexOf('严重') !==-1">严重</el-tag>
                        <el-tag class="s-tag-yellow" effect="plain" round v-if="scope.row.classify.indexOf('预警') !==-1">预警</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="start_time" label="开始时间" min-width="25%"></el-table-column>
                <el-table-column prop="end_time" label="结束时间" min-width="25%"></el-table-column>
                <el-table-column prop="name" label="故障/预警信息" min-width="35%" show-overflow-tooltip></el-table-column>
                <el-table-column label="操作" min-width="15%">
                    <template #default="scope">
                        <el-popover
                        class="item" effect="light"
                        placement="bottom"
                        title="指导建议"
                        width="200"
                        trigger="hover"
                        :content="scope.row.precautions">
                        <template #reference>
                         <el-link type="primary" underline="never" :class="['btn_text']">
                            指导建议
                        </el-link>
                        </template>
                    </el-popover>
                        <!-- <el-tooltip class="item" effect="light" :content="scope.row.propose" placement="left">
                            <el-link type="primary" underline="never" :class="['btn_text']">
                                指导建议
                            </el-link>
                        </el-tooltip> -->
                    </template>
                </el-table-column>
                <template #empty>
                    <img src="/img/no-data.svg" width="40" />
                    <p style="font-size:12px">当前暂无数据</p>
                </template>
            </el-table>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { defineEmits } from 'vue';
const emit = defineEmits(['callParentMethod']);
const router = useRouter();
const props = defineProps({
  AlarmInfoData: {
    type: Array,
    default: []
  }
})
console.log(props.AlarmInfoData);
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
const dateValue = ref([getCurrentDate(),getCurrentDate()])
const tagType = (state) => {
    switch (state) {
        case "一般":
            return ["warning", "一般"];
        case "轻微":
            return ["info", "轻微"];
        case "严重":
            return ["danger", "严重"];
        default:
            return ["info", state];
    }
}
const AlarmInfoData = [
    {
        state: 'TC1',
        unit_no: '1号机组',
        classify: '轻微',
        start_time: '2022-09-20 16:12:00',
        end_time: '2022-09-20 16:12:00',
        name: '通风机1-1过流故障',
        propose: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏',
    }, {
        state: 'TC2',
        unit_no: '2号机组',
        classify: '正常',
        start_time: '-',
        end_time: '-',
        name: '-',
        propose: '无需处理'
    }, {
        state: 'TC3',
        unit_no: '1号机组',
        classify: '一般',
        start_time: '2022-09-20 16:12:00',
        end_time: '2022-09-20 16:12:00',
        name: '新风阀U1故障',
        propose: '检查风阀执行器动作、反馈指示是否正常；检查全开反馈线路是否正常'
    }, {
        state: 'TC4',
        unit_no: '1号机组',
        classify: '严重',
        start_time: '2022-09-20 16:12:00',
        end_time: '2022-09-20 16:12:00',
        name: '冷凝风机1-2过流故障',
        propose: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏',
    }, {
        state: 'TC2',
        unit_no: '2号机组',
        classify: '轻微',
        start_time: '2022-09-20 16:12:00',
        end_time: '2022-09-20 16:12:00',
        name: '压缩机1-2高压连锁故障',
        propose: '检查冷凝风机运转是否正常；检查高压开关、排气温度保护是否正常；检查反馈线路是否正常'
    }, 
]
const propose = [
    {
     value: 'dvc_bocflt_ef_u11',
     name: '通风机1-1过流故障',
     title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
    },{
     value: 'dvc_bocflt_ef_u12',
     name: '通风机1-2过流故障',
     title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
    },{
     value: 'dvc_bocflt_cf_u11',
     name: '冷凝风机1-1过流故障',
     title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
    },{
     value: 'dvc_bocflt_cf_u12',
     name: '冷凝风机1-2过流故障',
     title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
    },{
     value: 'dvc_bflt_vfd_u11',
     name: '变频器1-1故障',
     title: '检查变频器运行是否正常；检查反馈线路是否正常'
    },{
     value: 'dvc_blpflt_comp_u11',
     name: '压缩机1-1低压故障',
     title: '检查制冷剂是否泄漏；检查通风机运转是否正常；检查低压开关是否损坏'
    },{
     value: 'dvc_bscflt_comp_u11',
     name: '压缩机1-1高压连锁故障',
     title: '检查冷凝风机运转是否正常；检查高压开关、排气温度保护是否正常；检查反馈线路是否正常'
    },{
     value: 'dvc_bflt_vfd_u12',
     name: '变频器1-2故障',
     title: '检查变频器运行是否正常；检查反馈线路是否正常'
    },{
     value: 'dvc_blpflt_comp_u12',
     name: '压缩机1-2低压故障',
     title: '检查制冷剂是否泄漏；检查通风机运转是否正常；检查低压开关是否损坏'
    },{
     value: 'dvc_bscflt_comp_u12',
     name: '压缩机1-2高压连锁故障',
     title: '检查冷凝风机运转是否正常；检查高压开关、排气温度保护是否正常；检查反馈线路是否正常'
    },{
     value: 'dvc_bflt_eev_u11',
     name: '电子膨胀阀1-1故障',
     title: '检查电子膨胀阀控制器显示参数是否正常；检查故障反馈线路是否正常'
    },{
     value: 'dvc_bflt_eev_u12',
     name: '电子膨胀阀1-2故障',
     title: '检查电子膨胀阀控制器显示参数是否正常；检查故障反馈线路是否正常'
    },{
     value: 'dvc_bflt_fad_u1',
     name: '新风阀U1故障',
     title: '检查风阀执行器动作、反馈指示是否正常；检查全开反馈线路是否正常'
    },{
     value: 'dvc_bflt_rad_u1',
     name: '回风阀U1故障',
     title: '检查风阀执行器动作、反馈指示是否正常；检查全开反馈线路是否正常'
    },{
     value: 'dvc_bflt_airclean_u1',
     name: '空气净化U1故障',
     title: '检查空气净化器运行是否正常；检查故障反馈线路是否正常'
    },{
     value: 'dvc_bflt_expboard_u1',
     name: '扩展模块U1故障',
     title: '检查扩展模块供电是否正常；检查扩展模块通讯线路是否存在虚接情况'
    },{
     value: 'dvc_bflt_frstemp_u1',
     name: '新风温度传感器U1故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_splytemp_u11',
     name: '送风温度传感器1-1故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_splytemp_u12',
     name: '送风温度传感器1-2故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_rnttemp_u1',
     name: '回风温度传感器U1故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_dfstemp_u11',
     name: '融霜温度传感器1-1故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_dfstemp_u12',
     name: '融霜温度传感器1-2故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bocflt_ef_u21',
     name: '通风机2-1过流故障',
     title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
    },{
     value: 'dvc_bocflt_ef_u22',
     name: '通风机2-2过流故障',
     title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
    },{
     value: 'dvc_bocflt_cf_u21',
     name: '冷凝风机2-1过流故障',
     title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
    },{
     value: 'dvc_bocflt_cf_u22',
     name: '冷凝风机2-2过流故障',
     title: '检查风机是否堵转；检查电机主电路绝缘是否正常；检查电机保护器是否损坏'
    },{
     value: 'dvc_bflt_vfd_u21',
     name: '变频器2-1故障',
     title: '检查变频器运行是否正常；检查反馈线路是否正常'
    },{
     value: 'dvc_blpflt_comp_u21',
     name: '压缩机2-1低压故障',
     title: '检查制冷剂是否泄漏；检查通风机运转是否正常；检查低压开关是否损坏'
    },{
     value: 'dvc_bscflt_comp_u21',
     name: '压缩机2-1高压连锁故障',
     title: '检查冷凝风机运转是否正常；检查高压开关、排气温度保护是否正常；检查反馈线路是否正常'
    },{
     value: 'dvc_bflt_vfd_u22',
     name: '变频器2-2故障',
     title: '检查变频器运行是否正常；检查反馈线路是否正常'
    },{
     value: 'dvc_blpflt_comp_u22',
     name: '压缩机2-2低压故障',
     title: '检查制冷剂是否泄漏；检查通风机运转是否正常；检查低压开关是否损坏'
    },{
     value: 'dvc_bscflt_comp_u22',
     name: '压缩机2-2高压连锁故障',
     title: '检查冷凝风机运转是否正常；检查高压开关、排气温度保护是否正常；检查反馈线路是否正常'
    },{
     value: 'dvc_bflt_eev_u21',
     name: '电子膨胀阀2-1故障',
     title: '检查电子膨胀阀控制器显示参数是否正常；检查故障反馈线路是否正常'
    },{
     value: 'dvc_bflt_eev_u22',
     name: '电子膨胀阀2-2故障',
     title: '检查电子膨胀阀控制器显示参数是否正常；检查故障反馈线路是否正常'
    },{
     value: 'dvc_bflt_fad_u2',
     name: '新风阀U2故障',
     title: '检查风阀执行器动作、反馈指示是否正常；检查全开反馈线路是否正常'
    },{
     value: 'dvc_bflt_rad_u2',
     name: '回风阀U2故障',
     title: '检查风阀执行器动作、反馈指示是否正常；检查全开反馈线路是否正常'
    },{
     value: 'dvc_bflt_airclean_u2',
     name: '空气净化U2故障',
     title: '检查空气净化器运行是否正常；检查故障反馈线路是否正常'
    },{
     value: 'dvc_bflt_expboard_u2',
     name: '扩展模块U2故障',
     title: '检查扩展模块供电是否正常；检查扩展模块通讯线路是否存在虚接情况'
    },{
     value: 'dvc_bflt_frstemp_u2',
     name: '新风温度传感器U2故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_splytemp_u21',
     name: '送风温度传感器2-1故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_splytemp_u22',
     name: '送风温度传感器2-2故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_rnttemp_u2',
     name: '回风温度传感器U2故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_dfstemp_u21',
     name: '融霜温度传感器2-1故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_dfstemp_u22',
     name: '融霜温度传感器2-2故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_vehtemp',
     name: '车厢温度传感器1故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_seattemp',
     name: '车厢温度传感器2故障',
     title: '检查温度传感器线路是否存在短路、虚接情况；测量温度传感器电阻是否正常'
    },{
     value: 'dvc_bflt_emergivt',
     name: '紧急逆变器故障',
     title: '检查紧急逆变器运行是否正常；检查故障反馈线路是否正常'
    },{
     value: 'dvc_bflt_mvbbus',
     name: 'MVB 通讯故障',
     title: '检查控制器供电是否正常；检查MVB连接器是否松动'
    },{
     value: 'dvc_bcomuflt_vfd_u11',
     name: '变频器1-1通讯故障',
     title: '检查变频器是否上电；检查通讯线路是否存在虚接情况'
    },{
     value: 'dvc_bcomuflt_vfd_u12',
     name: '变频器1-2通讯故障',
     title: '检查变频器是否上电；检查通讯线路是否存在虚接情况'
    },{
     value: 'dvc_bcomuflt_vfd_u21',
     name: '变频器2-1通讯故障',
     title: '检查变频器是否上电；检查通讯线路是否存在虚接情况'
    },{
     value: 'dvc_bcomuflt_vfd_u22',
     name: '变频器2-2通讯故障',
     title: '检查变频器是否上电；检查通讯线路是否存在虚接情况'
    },{
     value: 'dvc_bcomuflt_eev_u11',
     name: '电子膨胀阀1-1通讯故障',
     title: '检查电子膨胀阀控制器供电是否正常；检查通讯线路是否存在虚接情况'
    },{
     value: 'dvc_bcomuflt_eev_u12',
     name: '电子膨胀阀1-2通讯故障',
     title: '检查电子膨胀阀控制器供电是否正常；检查通讯线路是否存在虚接情况'
    },{
     value: 'dvc_bcomuflt_eev_u21',
     name: '电子膨胀阀2-1通讯故障',
     title: '检查电子膨胀阀控制器供电是否正常；检查通讯线路是否存在虚接情况'
    },{
     value: 'dvc_bcomuflt_eev_u22',
     name: '电子膨胀阀2-2通讯故障',
     title: '检查电子膨胀阀控制器供电是否正常；检查通讯线路是否存在虚接情况'
    },{
     value: 'dvc_bmcbflt_pwr_u1',
     name: '机组1供电故障',
     title: '检查主回路电路是否短路、漏电流是否过大；检查断路器是否损坏'
    },{
     value: 'dvc_bmcbflt_pwr_u2',
     name: '机组2供电故障',
     title: '检查主回路电路是否短路、漏电流是否过大；检查断路器是否损坏'
    }
]
function dateChange(e){
console.log(e);
const date1 = new Date(e[0]);
const date2 = new Date(e[1]);
const diff = Math.abs(date2 - date1);
const oneYearInMs = 365 * 24 * 60 * 60 * 1000;
const isMoreThanOneYear = diff > oneYearInMs;
if(isMoreThanOneYear){
    ElNotification({
        title: 'ERROR',
        message: '时间范围不可超过一年',
        duration: 4500,
    })
    return false
}
emit('callParentMethod',e);
}
</script>

<style scoped lang="scss">
$s-primary: #2186CF;
$s-danger: #e65355;
$s-warning: #ffa55c;
$s-info: #bfbfbf;
$s-success: #42ad5d;
$s-yellow: #ffe375;
$s-white:#ffffff;

.el-tag {
    background: rgba($s-success, .3);
    color: $s-success;
    border-color: $s-success;
    &.el-tag--plain {
        &.s-tag-yellow {
            background: rgba($s-yellow, .3) !important;
            color: $s-yellow !important;
            border-color: $s-yellow !important;
        }
        &.el-tag--info {
            background: rgba($s-info, .3) !important;
            color: $s-info !important;
            border-color: $s-info !important;
        }
        &.el-tag--danger {
            background: rgba($s-danger, .3);
            color: $s-danger;
            border-color: $s-danger;
        }
        &.el-tag--warning {
            background: rgba($s-warning, .3);
            color: $s-warning;
            border-color: $s-warning;
        }
        &.el-tag--success {
            background: rgba($s-success, .3);
            color: $s-success;
            border-color: $s-success;
        }
    }
}
// .s-tag {
//     border-radius: 20px;
//     background: rgba($s-success, .3);
//     color: $s-success;
//     border-color: $s-success;

//     &.s-tag-danger {
//         background: rgba($s-danger, .3);
//         color: $s-danger;
//         border-color: $s-danger;
//     }

//     &.s-tag-warning {
//         background: rgba($s-warning, .3);
//         color: $s-warning;
//         border-color: $s-warning;
//     }

//     &.s-tag-info {
//         background: rgba($s-info, .3);
//         color: $s-info;
//         border-color: $s-info;
//     }

//     &.s-tag-success {
//         background: rgba($s-success, .3);
//         color: $s-success;
//         border-color: $s-success;
//     }
// }

.el-table,
.el-table tr,
.el-table td,
.el-table th {
    background-color: transparent !important;
}

.mx-1 {
    background-color: rgba(0, 0, 0, 0);
    border-radius: 12px;
}

.el-tag--default {
    height: 20px;
}

// 操作文字
.btn_text {
    font-size: 14px;
    color: #2186CF !important;
}

.warp {
    width: 100%;
    border-radius: 10px;
    background-color: #171c32 !important;
}

.el-table {
    --el-table-border-color: none;
}

/* 标题左边线 */
.header {
    height: 50px;
    box-sizing: border-box;
    top: 0;
    left: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 30px;
    padding-right: 30px;
    > h4 {
        font-size: 16px;
        position: relative;
        margin-left: 30px;
        line-height: 1em;
        &::before {
        position: absolute;
        content: "";
        width: 4px;
        height: 1em;
        background: #fff;
        margin-left: -10px;
    }
    }
    :deep(.el-date-editor) {
        width: 240px;
    }
    :deep(.el-range-editor.is-active) {
        box-shadow: 0 0 0 1px #394153 inset;
    }
    :deep(.el-range-input) {
        color: rgba(255, 255, 255, 0.815);
    }
    :deep(.el-range-separator) {
        color: #a8abb2 !important;
    }
    :deep(.el-range-editor.el-input__inner) {
        background-color: #181F30;
        border: 1px solid #394153;
        box-shadow: 0 0 1px 0 #394153;
        color: #fff !important
    }
    /* 时间弹层 */
    .el-picker-panel__body-wrapper {
        background-color: #fff !important;
    }
}

</style>
