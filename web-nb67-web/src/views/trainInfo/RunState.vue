<template>
    <div class="warp">
        <div class="header">
            <h4>运行状态信息</h4>
        </div>
        <div>
            <el-table :data="props.RunStateData" stripe style="width: 100%">
                <el-table-column fixed prop="equipmentName" label="设备名称" min-width="25%"></el-table-column>
                <!-- <el-table-column :prop="index == 'equipmentName'?null:index" :label="(index == 'equipmentName'? null:index.slice(-1)%2==0?'2号机组':'1号机组')" min-width="25%" v-for="(item,index) in props.RunStateData[0]" :key="index"></el-table-column> -->
                <el-table-column prop="comp_u110" label="1号机组" min-width="25%"></el-table-column>
                <el-table-column prop="comp_u120" label="2号机组" min-width="25%"></el-table-column>
                <el-table-column prop="comp_u111" label="1号机组" min-width="25%"></el-table-column>
                <el-table-column prop="comp_u121" label="2号机组" min-width="25%"></el-table-column>
                <el-table-column prop="comp_u112" label="1号机组" min-width="25%"></el-table-column>
                <el-table-column prop="comp_u122" label="2号机组" min-width="25%"></el-table-column>
                <el-table-column prop="comp_u113" label="1号机组" min-width="25%"></el-table-column>
                <el-table-column prop="comp_u123" label="2号机组" min-width="25%"></el-table-column>
                <el-table-column prop="comp_u114" label="1号机组" min-width="25%"></el-table-column>
                <el-table-column prop="comp_u124" label="2号机组" min-width="25%"></el-table-column>
                <el-table-column prop="comp_u115" label="1号机组" min-width="25%"></el-table-column>
                <el-table-column prop="comp_u125" label="2号机组" min-width="25%"></el-table-column>
               
                <!-- <el-table-column prop="carCrew14" label="2号机组" min-width="25%"></el-table-column>
                <el-table-column prop="carCrew15" label="1号机组" min-width="25%"></el-table-column>
                <el-table-column prop="carCrew16" label="2号机组" min-width="25%"></el-table-column> -->
                <template #empty>
                    <img src="/src/assets/img/no-data.svg" width="40" />
                    <p style="font-size:12px">当前暂无数据</p>
                </template>
            </el-table>
        </div>
    </div>
</template>

<script setup>
const props = defineProps({
  RunStateData: {
    type: Array,
    default: []
  }
})
console.log(props.RunStateData.slice(1), props.RunStateData);
const RunStateData = [
    {
        equipmentName: '压缩机1',
        comp_u110: '运行',
        comp_u120: '运行',
        comp_u111: '运行',
        comp_u121: '运行',
        comp_u112: '运行',
        comp_u122: '运行',
        comp_u113: '运行',
        comp_u123: '运行',
        comp_u114: '停机',
        comp_u124: '运行',
        comp_u115: '运行',
        comp_u125: '运行',
        comp_u116: '运行',
        comp_u126: '运行',
        comp_u117: '运行',
        comp_u127: '运行'
    }, {
        equipmentName: '压缩机2',
        comp_u110: '运行',
        comp_u120: '运行',
        comp_u111: '停机',
        comp_u121: '运行',
        comp_u112: '运行',
        comp_u122: '运行',
        comp_u113: '运行',
        comp_u123: '运行',
        comp_u114: '运行',
        comp_u124: '运行',
        comp_u115: '运行',
        comp_u125: '运行',
        comp_u116: '运行',
        comp_u126: '运行',
        comp_u117: '运行',
        comp_u127: '运行'
    }, {
        equipmentName: '通风机',
        comp_u110: '运行',
        comp_u120: '运行',
        comp_u111: '运行',
        comp_u121: '运行',
        comp_u112: '运行',
        comp_u122: '运行',
        comp_u113: '运行',
        comp_u123: '运行',
        comp_u114: '运行',
        comp_u124: '运行',
        comp_u115: '运行',
        comp_u125: '运行',
        comp_u116: '运行',
        comp_u126: '运行',
        comp_u117: '运行',
        comp_u127: '运行'
    },
    // {
    //     equipmentName: '通风机2',
    //     comp_u110: '运行',
    //     comp_u120: '运行',
    //     comp_u111: '运行',
    //     comp_u121: '运行',
    //     comp_u112: '运行',
    //     comp_u122: '运行',
    //     comp_u113: '运行',
    //     comp_u123: '运行',
    //     comp_u114: '运行',
    //     comp_u124: '运行',
    //     comp_u115: '运行',
    //     comp_u125: '运行',
    //     comp_u116: '运行',
    //     comp_u126: '运行',
    //     comp_u117: '运行',
    //     comp_u127: '运行'
    // },
     {
        equipmentName: '冷凝风机',
        comp_u110: '运行',
        comp_u120: '运行',
        comp_u111: '运行',
        comp_u121: '运行',
        comp_u112: '运行',
        comp_u122: '运行',
        comp_u113: '停机',
        comp_u123: '运行',
        comp_u114: '运行',
        comp_u124: '停机',
        comp_u115: '运行',
        comp_u125: '运行',
        comp_u116: '停机',
        comp_u126: '运行',
        comp_u117: '停机',
        comp_u127: '运行'
    },
    //  {
    //     equipmentName: '冷凝风机2',
    //     comp_u110: '运行',
    //     comp_u120: '运行',
    //     comp_u111: '运行',
    //     comp_u121: '运行',
    //     comp_u112: '运行',
    //     comp_u122: '运行',
    //     comp_u113: '停机',
    //     comp_u123: '运行',
    //     comp_u114: '运行',
    //     comp_u124: '停机',
    //     comp_u115: '运行',
    //     comp_u125: '运行',
    //     comp_u116: '停机',
    //     comp_u126: '运行',
    //     comp_u117: '停机',
    //     comp_u127: '运行'
    // },
     {
        equipmentName: '新风阀开度',
        comp_u110: '30%',
        comp_u120: '100%',
        comp_u111: '100%',
        comp_u121: '100%',
        comp_u112: '100%',
        comp_u122: '100%',
        comp_u113: '60%',
        comp_u123: '100%',
        comp_u114: '100%',
        comp_u124: '100%',
        comp_u115: '90%',
        comp_u125: '100%',
        comp_u116: '20%',
        comp_u126: '100%',
        comp_u117: '100%',
        comp_u127: '100%'
    },{
        equipmentName: '回风阀开度',
        comp_u110: '100%',
        comp_u120: '100%',
        comp_u111: '100%',
        comp_u121: '100%',
        comp_u112: '100%',
        comp_u122: '100%',
        comp_u113: '10%',
        comp_u123: '100%',
        comp_u114: '100%',
        comp_u124: '50%',
        comp_u115: '100%',
        comp_u125: '100%',
        comp_u116: '100%',
        comp_u126: '70%',
        comp_u117: '100%',
        comp_u127: '100%'
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
