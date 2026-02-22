<template>
  <div>
    <div class="warp">
      <div class="header">
        <h4>实时报警</h4>
      </div>
      <div @mouseenter="pauseScroll(1)" @mouseleave="resumeScroll(1)">
        <el-table ref="tableRefEA" :data="props.tableDataEa" stripe style="width: 100%" height="290">
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
            <template #default="scope">
              {{ (scope.row.alarm_time)}}
            </template>
          </el-table-column>
          <el-table-column label="操作" min-width="25%">
            <template #default="scope">
              <el-popover
                        class="item" 
                        popper-class="advice-popover"
                        placement="left"
                        title="指导建议"
                        width="240"
                        trigger="hover"
                        :content="scope.row.precautions">
                        <template #reference>
                         <el-link type="primary" underline="never" class="btn_text">
                            指导建议
                        </el-link>
                        </template>
                    </el-popover>
            </template>
          </el-table-column>
          <template #empty>
            <img src="/src/assets/img/no-data.svg" width="80" height="50" />
            <p style="font-size:12px">当前暂无数据</p>
          </template>
        </el-table>
      </div>
    </div>
    <div class="warp">
      <div class="header">
        <h4>实时预警</h4>
      </div>
      <div @mouseenter="pauseScroll(2)" @mouseleave="resumeScroll(2)">
        <el-table ref="tableRefPA" :data="props.tableDataPa" stripe style="width: 100%" height="290">
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
            </template>
          </el-table-column>
          <el-table-column label="操作" align="center" min-width="25%">
            <template #default="scope">
              <el-popover
                        class="item" 
                        popper-class="advice-popover"
                        placement="left"
                        title="指导建议"
                        width="240"
                        trigger="hover"
                        :content="scope.row.precautions">
                        <template #reference>
                         <el-link type="primary" underline="never" class="btn_text">
                          处理措施
                        </el-link>
                        </template>
                    </el-popover>
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
import { ref, watch, onUnmounted, nextTick } from 'vue'

const router = useRouter();
const props = defineProps({
  tableDataEa: {
    type: Array,
    default: () => []
  },
  tableDataPa: {
    type: Array,
    default: () => []
  }
})

// === 自动滚动逻辑 ===
const tableRefEA = ref(null)
const tableRefPA = ref(null)
let scrollTimerEA = null
let scrollTimerPA = null

const startScroll = (tableRef, timerRefId) => {
  nextTick(() => {
    // 延迟更久等 DOM 完全加载
    setTimeout(() => {
      const tableDiv = tableRef.value?.$el.querySelector('.el-scrollbar__wrap')
      if (tableDiv) {
        // 先清除之前的防止叠加
        if(timerRefId === 1 && scrollTimerEA) clearInterval(scrollTimerEA)
        if(timerRefId === 2 && scrollTimerPA) clearInterval(scrollTimerPA)
        
        const dataLength = timerRefId === 1 ? props.tableDataEa.length : props.tableDataPa.length
        
        // 只有当有数据且内容高度 > 容器高度时，才开启自动滚动
        if (dataLength > 0 && tableDiv.scrollHeight > tableDiv.clientHeight) {
          const scrollFn = () => {
            tableDiv.scrollTop += 1
            // 当滚动到底部时返回顶部无缝循环
            if (Math.ceil(tableDiv.scrollTop) >= tableDiv.scrollHeight - tableDiv.clientHeight) {
              tableDiv.scrollTop = 0
            }
          }
          if(timerRefId === 1) {
            scrollTimerEA = setInterval(scrollFn, 50)
          } else {
            scrollTimerPA = setInterval(scrollFn, 50)
          }
        }
      }
    }, 500)
  })
}

const pauseScroll = (id) => {
  if (id === 1 && scrollTimerEA) clearInterval(scrollTimerEA)
  if (id === 2 && scrollTimerPA) clearInterval(scrollTimerPA)
}

const resumeScroll = (id) => {
  if (id === 1) startScroll(tableRefEA, 1)
  if (id === 2) startScroll(tableRefPA, 2)
}

watch(() => props.tableDataEa, () => {
  resumeScroll(1)
}, { deep: true, immediate: true })

watch(() => props.tableDataPa, () => {
  resumeScroll(2)
}, { deep: true, immediate: true })

onUnmounted(() => {
  if(scrollTimerEA) clearInterval(scrollTimerEA)
  if(scrollTimerPA) clearInterval(scrollTimerPA)
})
// =================

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
