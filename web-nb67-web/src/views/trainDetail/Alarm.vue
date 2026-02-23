<template>
  <div class="warp alarm-page">
    <TitleHeader :title="title" />

    <div style="padding: 0 15px">
      <el-table :data="props.data.list" style="width: 100%">
        <el-table-column prop="train_id" label="车厢号"></el-table-column>
        <el-table-column
          prop="typeIcon"
          label="故障类型"
          v-if="props.title == '报警信息'"
        >
          <template #default="scope">
            <el-tag
              class="mx-1"
              :type="tagType(scope.row.level)[0]"
              effect="plain"
              >{{ tagType(scope.row.level)[1] }}</el-tag
            >
          </template>
        </el-table-column>
        <el-table-column prop="startTime" label="开始时间" sortable>
          <!-- 时间去掉毫秒 -->
          <template #default="scope">{{
            formatTime(scope.row.startTime)
          }}</template>
        </el-table-column>
        <el-table-column prop="endTime" label="结束时间" sortable>
          <!-- 时间去掉毫秒 -->
          <template #default="scope">{{
            formatTime(scope.row.endTime)
          }}</template>
        </el-table-column>
        <el-table-column prop="desc" :label="props.title"></el-table-column>
        <el-table-column label="指导建议" align="center">
          <template #default="scope">
            <el-tooltip
              class="item"
              effect="dark"
              :content="scope.row.Advice"
              placement="left"
            >
              <el-button link class="btn_text">指导建议</el-button>
            </el-tooltip>
            <el-button
              link
              class="btn_text"
              @click="showDetail(scope.row.sig_list, scope.row.id)"
              v-if="
                scope.row.sig_list.length > 0 &&
                Object.keys(scope.row.sig_list[0]).length > 0
              "
              >详细数据</el-button
            >
            <el-tooltip
              class="item"
              effect="dark"
              :content="'暂无数据'"
              placement="left"
              v-else
            >
              <el-button link class="btn_text">N/A</el-button>
            </el-tooltip>
          </template>
        </el-table-column>

        <template #empty>
          <img src="/src/assets/img/no-data.svg" width="80" height="50" />
          <p style="font-size: 12px">当前暂无数据</p>
        </template>
      </el-table>
    </div>

    <div class="demo-pagination-block">
      <el-pagination
        v-model:currentPage="currentPage"
        :small="true"
        :background="true"
        layout="prev, pager, next, jumper"
        :total="props.data.count == undefined ? 0 : props.data.count"
        @current-change="handleCurrentChange"
      />
    </div>

    <el-dialog
      v-model="dialogVisible"
      width="80%"
      @opened="onOpenDialog"
      @closed="closedDialog"
    >
      <div style="height: 60vh" ref="echartId"></div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import * as echarts from "echarts";
import TitleHeader from "@/components/TitleHeader.vue";
import { formatTime } from '@/utils/time'
import { getEaOrPaById } from "@/api/api.js";

let interval;
const emit = defineEmits(["handleChangePage"]);

const props = defineProps({
  data: {
    type: Object,
    default: [],
  },
  title: {
    type: String,
  },
});

let echartId = ref();
let dialogVisible = ref(false);
let echartData = ref({});

//分页参数
const currentPage = ref(1);

const handleSizeChange = (limit) => {
  // emit('handleChangeLimit', limit)
};

const handleCurrentChange = (num) => {
  emit("handleChangePage", num);
};

const tagType = (state) => {
  switch (state) {
    case "中等":
      return ["warning", "中等"];
    case "轻微":
      return ["warning", "轻微"];
    case "严重":
      return ["danger", "严重"];
  }
};

const setEchartData = (list) => {
  let obj = {
    xAxisData: list.map((item) => {
      return formatTime(item["时间"])
    }),
    legendData: [],
    series: [],
  };

  let keys = Object.keys(list[0]);
  obj.legendData = keys.filter((item) => item != "时间");

  function setVal(key, data) {
    let arr = data.map((item) => item[key]);
    return arr;
  }

  obj.series = obj.legendData.map((name) => {
    return {
      name: name,
      type: "line",
      data: setVal(name, list),
    };
  });
  return obj;
};

const showDetail = (data, id) => {
  echartData.value = setEchartData(data);
  dialogVisible.value = true;
  if (interval != undefined) {
    // 清空定时器
    window.clearTimeout(interval);
  }
  //   定时获取对应信号数据
  interval = setInterval(() => {
    getEaOrPaById({ id: id }).then((res) => {
      echartData.value = setEchartData(JSON.parse(res.data[0].sig_list));
    });
  }, 5 * 1000);
};

const initEchart = () => {
  let option = {
    tooltip: {
      trigger: "axis",
      formatter: function (params) {
        let relVal = params[0].name;
        for (let i = 0; i < params.length; i++) {
          let name = params[i].seriesName;
          let dw = "℃";
          if (name.indexOf("频率") != -1) {
            dw = "Hz";
          }
          if (name.indexOf("电流") != -1) {
            dw = "A";
          }
          if (name.indexOf("电压") != -1) {
            dw = "V";
          }
          if (name.indexOf("压力") != -1) {
            dw = "bar";
          }
          if (name.indexOf("热度") != -1) {
            dw = "K";
          }
          if (name.indexOf("开度") != -1) {
            dw = "%";
          }
          relVal +=
            "<br/>" + params[i].marker + name + " : " + params[i].value + dw;
        }
        return relVal;
      },
    },
    grid: {
      x: 50,
      y: 50,
      x2: 30,
      y2: 60,
      borderWidth: 10,
    },

    xAxis: {
      data: echartData.value.xAxisData,
      axisLabel: {
        color: "#fff",
      },
    },
    legend: {
      type: "scroll",
      data: echartData.value.legendData,

      textStyle: {
        fontSize: 12,
        color: "#ffffff",
      },
    },
    yAxis: {
      splitLine: {
        show: true,
        lineStyle: {
          type: "dashed",
        },
      },
      axisLabel: {
        color: "#fff",
      },
    },
    series: echartData.value.series,
  };

  let alarmEcharts = echarts.init(echartId.value);
  alarmEcharts.setOption(option);
};

const onOpenDialog = () => {
  initEchart();
};
const closedDialog = () => {
  if (interval != undefined) {
    // 清空定时器
    window.clearTimeout(interval);
  }
};

onMounted(() => {});
</script>

<style scoped lang="scss">
$s-primary: #2186cf;
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
.not_click {
  cursor: not-allowed;
}

.alarm-page :deep(.el-dialog) {
  background-color: #2b3548;
}

.demo-pagination-block {
  padding: 15px;
  display: flex;
  justify-content: flex-end;
}
:deep(.el-pagination__total) {
  color: #fff;
}
:deep(.el-pagination__jump) {
  color: #fff;
}
:deep(.el-input__inner) {
  color: #fff;
}
</style>
