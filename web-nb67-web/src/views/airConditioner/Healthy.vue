<template>
    <div class="warp">
        <div class="header">
            <h4>健康评估信息</h4>
        </div>
        <!-- <div class="content">
            <el-table :data="props.tableData" stripe style="width: 100%">
                <el-table-column prop="name" label="设备名称" min-width="25%"></el-table-column>
                <el-table-column prop="state" label="健康状态" min-width="25%">
                    <template #default="scope">
                        <el-tag
                            class="mx-1"
                            :type="tagType(scope.row.start)[0]"
                            effect="plain"
                        >{{ tagType(scope.row.start)[1] }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="errorNum1" label="累计运行时间" min-width="25%">
                    <template #default="scope">
                        <span v-if="scope.row.runDate>0">{{scope.row.runDate}} 小时</span>
                         <span v-else>N/A</span></template>
                </el-table-column>
                <el-table-column prop="errorNum2" label="动作次数" min-width="25%">
                    <template #default="scope"> 
                        <span v-if="scope.row.runCount>0">{{scope.row.runCount}} 次</span>
                        <span v-else>N/A</span></template>
                </el-table-column>
                <template #empty>
                    <img src="/src/assets/img/no-data.svg" width="40" />
                    <p style="font-size:12px">当前暂无数据</p>
                </template>
            </el-table>
        </div> -->
        <!-- 
            三个状态tag
            <el-tag class="s-tag s-tag-warning" :type="warning" effect="plain">亚健康</el-tag>
            <el-tag class="s-tag s-tag-danger" :type="danger" effect="plain">非健康</el-tag>
            <el-tag class="s-tag" effect="plain">健康</el-tag> 
        -->
        <div class="content">
            <el-table :data="props.tableData"  stripe style="width: 100%">
                <el-table-column prop="equipment_name" label="设备名称" min-width="25%" style="padding-left: 10px !important;"></el-table-column>
                <el-table-column prop="health_status" label="健康状态" min-width="25%">
                    <!-- <template #default="scope">
                        <el-tag class="s-tag s-tag-warning" :type="warning" effect="plain">亚健康</el-tag>
                        <el-tag class="s-tag s-tag-danger" :type="danger" effect="plain">非健康</el-tag>
                        <el-tag class="s-tag" effect="plain">健康</el-tag>
                    </template> -->
                    <template #default="scope">
                        <el-tag type="scuess" effect="plain" round v-if="scope.row.health_status == '健康'">健康</el-tag>
                        <el-tag type="warning" effect="plain" round v-if="scope.row.health_status == '亚健康'">亚健康</el-tag>
                        <el-tag type="danger" effect="plain" round v-if="scope.row.health_status == '非健康'">非健康</el-tag>
                    </template>
                </el-table-column>
                <el-table-column prop="cumulative_running_time" label="累计运行时间" min-width="25%">
                     <template #default="scope">
                        <span v-if="scope.row.cumulative_running_time !=='-'">{{scope.row.cumulative_running_time}}小时</span>
                        <span v-else>{{scope.row.cumulative_running_time}}</span>
                    </template>
                </el-table-column>
                 
                <el-table-column prop="number_actions" label="动作次数" min-width="25%">
                    <template #default="scope">
                        <span>{{scope.row.number_actions}}次</span>
                    </template>
                </el-table-column>
                <el-table-column label="操作" min-width="15%">
                    <template #default="scope">
                        <el-tooltip class="item" effect="light" :content="scope.row.precautions" placement="left">
                            <el-link v-if="scope.row.health_status !== '健康'" type="primary" underline="never" :class="['btn_text']">
                                指导建议
                            </el-link>
                        </el-tooltip>
                    </template>
                </el-table-column>
                <template #empty>
                    <img src="/src/assets/img/no-data.svg" width="40" />
                    <p style="font-size:12px">当前暂无数据</p>
                </template>
            </el-table>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const props = defineProps({
    tableData: {
        type: Array,
        default: []
    }
})

const tagType = (state) => {
    switch (state) {
        case 'health':
            return ['scuess', '健康'];
        case 'subHealth':
            return ['warning', '亚健康'];
        case 'unhealthy':
            return ['danger', '非健康'];
    }
}

const tableData = [
    {
        equipment_name: '通风机',
        health_status: '健康',
        cumulative_running_time: 2114,
        number_actions: 788
    },{
        equipment_name: '冷凝风机',
        health_status: '亚健康',
        cumulative_running_time: 192,
        number_actions: 190
    },{
        equipment_name: '新风阀',
        health_status: '非健康',
        cumulative_running_time: 1084,
        number_actions: 230
    },{
        equipment_name: '回风阀',
        health_status: '健康',
        cumulative_running_time: 2223,
        number_actions: 190
    },{
        equipment_name: '压缩机1',
        health_status: '健康',
        cumulative_running_time: 1696,
        number_actions: 223
    },{
        equipment_name: '压缩机2',
        health_status: '健康',
        cumulative_running_time: 1193,
        number_actions: 118
    }
]
</script>

<style scoped lang="scss">
$s-primary: #2186CF;
$s-danger: #e65355;
$s-warning: #ffa55c;
$s-info: #bfbfbf;
$s-success: #42ad5d;
$s-yellow: #ffe375;
$s-white: #ffffff;

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
.warp {
    width: 100%;
    border-radius: 10px;
    background-color: #181F30 !important;
    margin-bottom: 15px;
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
    gap: 30px;
}
.header>h4 {
    font-size: 16px;
    position: relative;
    margin-left: 30px;
    line-height: 1em;
}
.header>h4::before {
    position: absolute;
    content: "";
    width: 4px;
    height: 1em;
    background: #fff;
    margin-left: -10px;
}
</style>
