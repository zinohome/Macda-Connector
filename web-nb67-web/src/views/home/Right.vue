<template>
  <div>
    <div class="warp">
      <div class="header">
        <h4>实时报警</h4>
      </div>
      <div>
        <el-table :data="props.tableDataEa" stripe style="width: 100%" height="290">
          <el-table-column prop="carriage_no" show-overflow-tooltip label="车厢号" min-width="20%"></el-table-column>
          <el-table-column
            prop="name"
            show-overflow-tooltip
            label="报警名称"
            min-width="30%"
          ></el-table-column>
          <el-table-column
            prop="alarm_time"
            show-overflow-tooltip
            align="left"
            label="发生时间"
            min-width="25%"
          >
            <!-- 时间去掉毫秒 -->
            <template #default="scope">
              {{ (scope.row.alarm_time)}}
              <!-- {{ (scope.row.starttime_str).split(".")[0]}} -->
            </template>
          </el-table-column>
          <el-table-column label="操作" min-width="25%">
            <template #default="scope">
              <el-popover
                        class="item" effect="light"
                        placement="left"
                        title="指导建议"
                        width="200"
                        trigger="hover"
                        :content="scope.row.precautions">
                        <template #reference>
                         <el-link type="primary" :underline="false" :class="['btn_text']">
                            指导建议
                        </el-link>
                        </template>
                    </el-popover>
              <!-- <el-tooltip class="item" effect="light" :content="scope.row.precautions" placement="left">
                <el-link type="primary" :underline="false" @click="gotoPathTrainInfo()" class="btn_text">
                  处理措施
                </el-link>
              </el-tooltip> -->
            </template>
          </el-table-column>
          <template #empty>
            <img src="/src/assets/img/no-data.svg" width="80" height="50" />
            <p style="font-size:12px">当前暂无数据</p>
          </template>
        </el-table>
        
        <!-- <el-table :data="props.tableDataEa" stripe style="width: 100%" height="290">
          <el-table-column prop="FL_dvcTrain" label="车号" min-width="20%"></el-table-column>
          <el-table-column
            prop="descText"
            show-overflow-tooltip
            label="报警名称"
            min-width="30%"
          ></el-table-column>
          <el-table-column
            prop="starttime_str"
            show-overflow-tooltip
            align="starttime_str"
            label="发生时间"
            min-width="30%"
          >
            <template #default="scope">
              {{ (scope.row.starttime_str).split(".")[0]}}
            </template>
          </el-table-column>
          <el-table-column label="操作" min-width="20%">
            <template #default="scope">
              <el-tooltip class="item" effect="dark" :content="scope.row.Advice" placement="left">
                <el-button type="text" class="btn_text">处理措施</el-button>
              </el-tooltip>
            </template>
          </el-table-column>
          <template #empty>
            <img src="/src/assets/img/no-data.svg" width="80" height="50" />
            <p style="font-size:12px">当前暂无数据</p>
          </template>
        </el-table> -->
      </div>
    </div>
    <div class="warp">
      <div class="header">
        <h4>实时预警</h4>
      </div>
      <div>
        <el-table :data="props.tableDataPa" stripe style="width: 100%" height="290">
          <el-table-column prop="carriage_no" show-overflow-tooltip label="车厢号" min-width="20%"></el-table-column>
          <el-table-column
            prop="name"
            show-overflow-tooltip
            label="预警名称"
            min-width="30%"
          ></el-table-column>
          <el-table-column
            prop="warning_time"
            show-overflow-tooltip
            align="left"
            label="发生时间"
            min-width="25%"
          >
            <template #default="scope">
              {{ (scope.row.warning_time)}}
              <!-- {{ (scope.row.starttime_str).split(".")[0]}} -->
            </template>
          </el-table-column>
          <el-table-column label="操作" align="center" min-width="25%">
            <template #default="scope">
              <el-popover
                        class="item" effect="light"
                        placement="left"
                        title="处理措施"
                        width="200"
                        trigger="hover"
                        :content="scope.row.precautions">
                        <template #reference>
                         <el-link type="primary" :underline="false" :class="['btn_text']">
                          处理措施
                        </el-link>
                        </template>
                    </el-popover>
              <!-- <el-tooltip class="item" effect="light" content="运营结束后回库处理" placement="left">
                <el-link type="primary" :underline="false" class="btn_text">
                  处理措施
                </el-link>
              </el-tooltip> -->
            </template>
          </el-table-column>
          <template #empty>
            <img src="/src/assets/img/no-data.svg" width="80" height="50" />
            <p style="font-size:12px">当前暂无数据</p>
          </template>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
const router = useRouter();
// import { ref, onMounted } from 'vue'
const props = defineProps({
  tableDataEa: {
    type: Array,
    default: []
  },
  tableDataPa: {
    type: Array,
    default: []
  }
})
const gtableData = [
  {
    F_trainID: '05003',
    descText: '空调蒸发器A报警',
    starttime_str: '23:52:11'
  }, {
    F_trainID: '05003',
    descText: '空调蒸发器B报警',
    starttime_str: '23:52:11'
  }, {
    F_trainID: '05003',
    descText: '空调蒸发器B报警',
    starttime_str: '23:52:11'
  }, {
    F_trainID: '05003',
    descText: '空调蒸发器B报警',
    starttime_str: '23:52:11'
  }, {
    F_trainID: '05003',
    descText: '空调蒸发器B报警',
    starttime_str: '23:52:11'
  }, {
    F_trainID: '05003',
    descText: '空调蒸发器B报警',
    starttime_str: '23:52:11'
  }, {
    F_trainID: '05003',
    descText: '空调蒸发器B报警',
    starttime_str: '23:52:11'
  }, {
    F_trainID: '05003',
    descText: '空调蒸发器B报警',
    starttime_str: '23:52:11'
  }
]
// const gtableData = [
//   {
//     F_trainID: '05003',
//     descText: '空调蒸发器A报警',
//     starttime_str: '2022-10-19'
//   }, {
//     F_trainID: '05003',
//     descText: '空调蒸发器B报警',
//     starttime_str: '2022-10-19'
//   }
// ]

const gotoPathTrainDetail = (trainId) => {
  router.push({
    name: 'trainDetail',
    params: {
      trainId
    }
  })
}

const activeClass = (state) => {
  let className = ''
  if (state == 'unstart') {
    className = 'not_click'
  } else {
    className = ''
  }
  return className
}

</script>

<style scoped>
.el-table,
.el-table tr,
.el-table td,
.el-table th {
  background-color: transparent !important;
}
.el-table {
  --el-table-border-color: none;
}

.btn_text {
  font-size: 14px;
  color: #2186CF !important
}

.warp {
  width: 100%;
  height: 341.5px;
  border-radius: 10px;
  background-color: #181F30;
  border: 1px solid #394153;
  margin-top: 0;
  margin-bottom: 15px;
}
.warp1 {
  height: 345px;
  padding: 0 8px;
  border-radius: 10px;
  background-color: #171c32;
  padding-left: 10px;
}
.top-10 {
  margin-top: 13px;
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
:deep(.el-table .el-table__cell) {
    padding: 10px !important;
}
</style>
