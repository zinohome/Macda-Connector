<template>
    <div class="warp" @mouseenter="pauseScroll" @mouseleave="resumeScroll">
        <!-- 实时报警 -->
        <div class="section-header">
            <span class="title-icon"></span>
            <span class="title-text">实时报警</span>
        </div>
        <el-table ref="tableRef" :data="props.ActualWarningData" stripe height="245px" style="width: 100%;">
            <el-table-column prop="carriage_no" label="车厢号" min-width="15%"></el-table-column>
            <el-table-column prop="name" label="报警名称" min-width="45%" show-overflow-tooltip></el-table-column>
            <el-table-column prop="alarm_time" label="报警时间" min-width="25%" show-overflow-tooltip></el-table-column>
            <el-table-column label="操作" min-width="15%">
                <template #default="scope">
                    <el-popover
                        class="item" 
                        popper-class="advice-popover"
                        placement="bottom"
                        title="指导建议"
                        width="240"
                        trigger="hover"
                        :content="scope.row.precautions">
                        <template #reference>
                         <el-link type="primary" underline="never" :class="['btn_text']">
                            指导建议
                        </el-link>
                        </template>
                    </el-popover>
                </template>
            </el-table-column>
            <template #empty>
                <img src="/img/no-data.svg" width="40" />
                <p style="font-size:12px">当前暂无数据</p>
            </template>
        </el-table>
    </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'

const props = defineProps({
  ActualWarningData: {
    type: Array,
    default: []
  }
})
const router = useRouter();
const tableRef = ref(null)
let scrollTimer = null

const startScroll = () => {
  nextTick(() => {
    setTimeout(() => {
      const tableDiv = tableRef.value?.$el.querySelector('.el-scrollbar__wrap')
      if (tableDiv) {
        if(scrollTimer) clearInterval(scrollTimer)
        // 只有当有数据且内容高度 > 容器高度时，才开启自动滚动
        if (props.ActualWarningData.length > 0 && tableDiv.scrollHeight > tableDiv.clientHeight) {
          const scrollFn = () => {
            tableDiv.scrollTop += 1
            if (Math.ceil(tableDiv.scrollTop) >= tableDiv.scrollHeight - tableDiv.clientHeight) {
              tableDiv.scrollTop = 0
            }
          }
          scrollTimer = setInterval(scrollFn, 50)
        }
      }
    }, 500)
  })
}

const pauseScroll = () => {
  if (scrollTimer) clearInterval(scrollTimer)
}

const resumeScroll = () => {
  startScroll()
}

watch(() => props.ActualWarningData, () => {
  resumeScroll()
}, { deep: true, immediate: true })

onUnmounted(() => {
  if(scrollTimer) clearInterval(scrollTimer)
})
</script>

<style scoped>
.el-table,
.el-table tr,
.el-table td,
.el-table th {
    background-color: transparent !important;
}

.warp {
    width: 100%;
}

.el-table {
    --el-table-border-color: none;
}

.section-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
    padding: 0;

    .title-icon {
        width: 4px;
        height: 16px;
        background: #e65355;
        border-radius: 2px;
    }

    .title-text {
        color: #ffffff;
        font-size: 15px;
        font-weight: bold;
    }
}

.btn_text {
    font-size: 14px;
    color: #2186CF !important;
}
</style>
