<template>
  <div class="warp">
    <div class="header">
      <h4>空调系统</h4>
    </div>
    <!-- <div>
      <el-table :data="prop.tableData" stripe height="650">
        <el-table-column prop="F_trainID" label="车号" min-width="20%"></el-table-column>
        <el-table-column prop="areaNum" label="设备状态" min-width="28%">
          <template #default="scope">
            
            <el-tag
              class="mx-1"
              :type="tagType(scope.row.start)[0]"
              effect="plain"
            >{{ tagType(scope.row.start)[1] }}</el-tag>
            <el-tag
              class="mx-1"
              :type="tagType(scope.row.start)[0]"
              effect="plain"
            >{{ tagType(scope.row.start)[1] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="areaNum" label="报警数" min-width="23%"></el-table-column>
        <el-table-column prop="paNum" label="预警数" min-width="23%"></el-table-column>
        <el-table-column label="操作" min-width="25%">
          <template #default="scope">
            <el-button
              type="text"
              @click="gotoPath(scope.row.F_trainID)"
              :class="['btn_text', activeClass(scope.row.start)]"
            >查看详情</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <img src="/img/no-data.svg" width="80" height="50" />
          <p style="font-size:12px">当前暂无数据</p>
        </template>
      </el-table>
    </div> -->
    <!-- <el-tag class="s-tag s-tag-info" type="info" effect="dark">离线</el-tag>
    <el-tag class="s-tag s-tag-danger" type="danger" effect="light">告警</el-tag>
    <el-tag class="s-tag s-tag-warning" type="warning" effect="plain">预警</el-tag>
    <el-tag class="s-tag s-tag-success" effect="plain">正常</el-tag> -->
    <div>
      <el-table :data="props.tableData" stripe height="650">
        <el-table-column prop="train_no" label="列车号" min-width="20%"></el-table-column>
        <el-table-column prop="status" label="设备状态" min-width="30%">
          <template #default="scope">
            <el-tag
              :type="scope.row.alarm_count == 0 && (scope.row.warning_count == 0 || scope.row.warning_count == null)?'success':'danger'"
              effect="plain" 
              round
              >
              {{scope.row.alarm_count == 0 && (scope.row.warning_count == 0 || scope.row.warning_count == null)?'正常':'异常'}}
              </el-tag>
          </template>
        </el-table-column>
         <!-- :type="scope.row.status == '正常'?'success':
              scope.row.status == '异常'?'danger':
              scope.row.status == '预警'?'warning':
              scope.row.status == '离线'?'info':''" -->
        <!-- <el-table-column prop="areaNum" label="设备状态" min-width="25%">
          <template #default="scope">
            <el-tag 
              :type="{
                '': tagType(scope.row.start)[0] == '',
                'info': tagType(scope.row.start)[0] == 'info',
                'danger': tagType(scope.row.start)[0] == 'danger',
                'warning': tagType(scope.row.start)[0] == 'warning'
              }"
              effect="plain" 
              round
            >{{ tagType(scope.row.start)[1] }}正常</el-tag> -->
            
            <!--  四种状态  
              <el-tag type="info" effect="plain" round>离线</el-tag>
              <el-tag effect="plain" round>正常</el-tag>
              <el-tag type="danger" effect="plain" round>告警</el-tag>
              <el-tag type="warning" effect="plain" round>预警</el-tag> 
            -->
          <!-- </template>
        </el-table-column> -->
        <el-table-column prop="alarm_count" label="报警数" min-width="22%"></el-table-column>
        <el-table-column prop="warning_count" label="预警数" min-width="22%">
          <template #default="scope">
            {{scope.row.warning_count == null ? 0:scope.row.warning_count}}
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="28%">
          <template #default="scope">
            <el-link type="primary" underline="never" @click="gotoPath(scope.row.train_no)" :class="['btn_text', activeClass(scope.row.start)]">
              查看详情
            </el-link>
          </template>
        </el-table-column>
        <template #empty>
          <img src="/img/no-data.svg" width="80" height="50" />
          <p style="font-size:12px">当前暂无数据</p>
        </template>
      </el-table>
    </div>
  </div>
</template>

<script setup>
// import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
const router = useRouter();

const props = defineProps({
  tableData: {
    type: Array,
    default: []
  }
})
const tagType = (state) => {
  console.log(state)
  let val = []
  switch (state) {
    case 'unstart':
      val = ['info', '离线']
      break;
    case 'runstart':
      val = ['', '正常']
      break;
    case 'ea':
      val = ['danger', '告警']
      break;
    case 'pa':
      val = ['warning', '预警']
      break;
  }
  return val
}

const gotoPath = (trainId) => {
  console.log(trainId)
  router.push({
    name: 'trainInfo',
    query: {
      trainNo:trainId
    }
    // params: {
    //   trainId
    // }
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

const gtableData = [
  {
    train_no: '10551',
    alarm_count: '0',
    warning_count: '0'
  },
  {
    train_no: '10551',
    alarm_count: '12',
    warning_count: '22'
  },
  {
    train_no: '10551',
    alarm_count: '0',
    warning_count: '0'
  },
  {
    train_no: '10551',
    alarm_count: '12',
    warning_count: '22'
  },
  {
    train_no: '10551',
    status: 0,
    alarm_count: '12',
    warning_count: '22'
  },
  {
    train_no: '10551',
    alarm_count: '12',
    warning_count: '22'
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
    &.el-tag--default {
      background: rgba($s-success, .3);
      color: $s-success;
      border-color: $s-success;
    }
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
// 表格
.el-table,
.el-table tr,
.el-table td,
.el-table th
 {
  background: transparent;
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
.header > h4 {
  font-size: 16px;
  position: relative;
  margin-left: 30px;
  line-height: 1em;
}
.header > h4::before {
  position: absolute;
  content: "";
  width: 4px;
  height: 1em;
  background: #fff;
  margin-left: -10px;
}

// 操作文字
.btn_text {
  font-size: 14px;
  color: #2186CF !important;
}
.warp {
  width: 100%;
  // height: 703px;
  border-radius: 10px;
  background-color: #181F30;
  border: 1px solid #394153;
  margin-top: 0;
  margin-bottom: 15px;
}
.not_click {
  color: #909399 !important;
  /* pointer-events: none; */
  font-size: 12px;
}
:deep(.el-table .el-table__cell) {
    padding-left: 10px!important;
}
.cell{

}
</style>
