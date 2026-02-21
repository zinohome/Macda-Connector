<template>
    <div class="warp">
        <div class="header">
            <h4>车厢选择</h4>
            <el-popover placement="right" :width="160" popper-class="popper-help">
                <template #reference>
                    <el-icon size="16px" @click="openHelp()">
                        <QuestionFilled />
                    </el-icon>
                </template>
                <el-row :gutter="10" justify="space-between" align="middle">
                    <el-col :span="12">正常</el-col>
                    <el-col :span="12" style="text-align: right">
                        <el-icon color="#42ad5d" size="20px">
                            <SuccessFilled />
                        </el-icon>
                    </el-col>
                </el-row>
                <!-- <el-divider /> -->
                <el-row :gutter="10" justify="space-between" align="middle">
                    <el-col :span="12">故障</el-col>
                    <el-col :span="12" style="text-align: right">
                        <el-icon color="#e65355" size="20px">
                            <CircleCloseFilled />
                        </el-icon>
                    </el-col>
                </el-row>
                <!-- <el-divider /> -->
                <el-row :gutter="10" justify="space-between" align="middle">
                    <el-col :span="12">预警</el-col>
                    <el-col :span="12" style="text-align: right">
                        <el-icon color="#ffa55c" size="20px">
                            <WarnTriangleFilled />
                        </el-icon>
                    </el-col>
                </el-row>
            </el-popover>
        </div>
        <div class="container">
            <!-- <div class="c-l">车厢选择</div> -->
            <div class="c-r">
                <div class="carWarpper">
                    <div class="carItem" @click="gotoPathAirConditioner(head.carriage_no)">
                        <div>
                            <img src="/src/assets/img/carHeader.svg" />
                        </div>
                        <el-space warp class="itemAirCondition">
                            <el-icon color="#ffa55c" size="23px" v-if="(head.warning_count !== 0 && head.warning_count !== undefined && head.warning_count !== null) && head.alarm_count == 0">
                                <WarnTriangleFilled />
                            </el-icon>
                            <el-icon color="#42ad5d" size="20px" v-if="head.alarm_count == 0 && (head.warning_count == 0 || head.warning_count == undefined || head.warning_count == null)">
                                <SuccessFilled />
                            </el-icon>
                            <el-icon color="#e65355" size="20px" v-if="head.alarm_count !== 0">
                                <CircleCloseFilled />
                            </el-icon>
                            <span>{{ carriage_name[0]}}</span>
                        </el-space>
                    </div>
                    <div class="carItem" v-for="i in center" :key="i" @click="gotoPathAirConditioner(i.carriage_no)">
                        <div>
                            <img src="/src/assets/img/carBody.svg" />
                        </div>
                        <el-space warp class="itemAirCondition">
                            <el-icon color="#ffa55c" size="23px" v-if="(i.warning_count !== 0 && i.warning_count !== undefined && i.warning_count !== null) && i.alarm_count == 0">
                                <WarnTriangleFilled />
                            </el-icon>
                            <el-icon color="#42ad5d" size="20px" v-if="i.alarm_count == 0 && (i.warning_count == 0 || i.warning_count == undefined || i.warning_count == null)">
                                <SuccessFilled />
                            </el-icon>
                            <el-icon color="#e65355" size="20px" v-if="i.alarm_count !== 0">
                                <CircleCloseFilled />
                            </el-icon>
                            <span>{{i.name}}</span>
                        </el-space>
                    </div>
                    <div class="carItem" @click="gotoPathAirConditioner(tail.carriage_no)">
                        <div>
                            <img src="/src/assets/img/carEnd.svg" />
                        </div>
                        <el-space warp class="itemAirCondition">
                            <el-icon color="#ffa55c" size="23px" v-if="(tail.warning_count !== 0 && tail.warning_count !== undefined && tail.warning_count !== null) && tail.alarm_count == 0">
                                <WarnTriangleFilled />
                            </el-icon>
                            <el-icon color="#42ad5d" size="20px" v-if="tail.alarm_count == 0 && (tail.warning_count == 0 || tail.warning_count == undefined || tail.warning_count == null)">
                                <SuccessFilled />
                            </el-icon>
                            <el-icon color="#e65355" size="20px" v-if="tail.alarm_count !== 0">
                                <CircleCloseFilled />
                            </el-icon>
                            <span>{{ tail.name}}</span>
                        </el-space>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { useRouter, useRoute } from 'vue-router'
import { ref, onMounted, onUnmounted } from 'vue'
import { getTrainSelection } from '@/api/api'
let trainSelectionData = ref([{
  name: 'Tc1',
},{
  name: 'Mp1',
},{
  name: 'M1',
},{
  name: 'M2',
},{
  name: 'MP2',
},{
  name: 'Tc2',
}
])
// let center = ref([])
let route = useRoute()
const router = useRouter();
// const head = ref({})
// const tail = ref({})
const carriage_name = ref(['Tc1','Mp1','M1','M2','Mp2','Tc2'])
const center_name = ref(['M1','M2','Mp2','Tc2'])

let center = ref([
    {
      name: 'Mp1',
      alarm_count: 0,
      warning_count: 0,
      carriage_no: '2'
    },{
      name: 'M1',
      alarm_count: 0,
      warning_count: 1,
      carriage_no: '3'
    },{
      name: 'M2',
      alarm_count: 0,
      warning_count: 1,
      carriage_no: '4'
    },{
      name: 'MP2',
      alarm_count: 0,
      warning_count: 0,
      carriage_no: '5'
    }
])
let head = ref(
    {
      name: 'Tc1',
      alarm_count: 0,
      warning_count: 0,
      carriage_no: '1'
    }
)
let tail = ref(
    {
      name: 'Tc2',
      alarm_count: 0,
      warning_count: 0,
      carriage_no: '6'
    }
)

const props = defineProps({
  trainId: {
    type: String,
    default: ''
  }
})
function getapi () {
getTrainSelection(props.trainId).then((res)=>{
        res.vw_carriage_status.forEach((value)=>{
            trainSelectionData.value.forEach((item)=>{
                if(value.carriage_no == item.carriage_no){
                  item.alarm_count = value.alarm_count
                }
            })
        })
        res.vw_carriage_predict_status.forEach((value)=>{
            trainSelectionData.value.forEach((item)=>{
                if(value.carriage_no == item.carriage_no){
                  item.warning_count = value.warning_count
                }
            })
        })
        console.log(trainSelectionData.value,'55555')
        head.value = trainSelectionData.value[0]
        console.log(head.value.warning_count,"head.value.warning_count");
        tail.value = trainSelectionData.value[5]
        console.log(tail.value.carriage_no,'tail.value');
        center.value = trainSelectionData.value.slice(1,5)
        console.log(center[0].value.carriage_no,'center.value');
        console.log(trainSelectionData.value[0],trainSelectionData.value[5],trainSelectionData.value.slice(1,5))
        if(trainSelectionData.value.slice(2).length == 0){
            center_name.value.forEach((items)=>{
            // item.name = items
            // item.air_conditioning_status = ''
            center.value.push({
                name: items,
                air_conditioning_status: ''
            })
           })
        // center.value.forEach((item)=>{
        //    center_name.value.forEach((items)=>{
        //     item.name = items
        //     item.air_conditioning_status = ''
        //    })
        // })
        console.log(center.value)
        } else {
            center.value = trainSelectionData.value.slice(2)
            center.value.forEach((item)=>{
            center_name.value.forEach((items)=>{
            item.name = items
           })
        })
        }
        console.log(center)
    })
}
const gotoPathAirConditioner = (carriageID) => {
  // alert(carriageID)
  if(!carriageID){
       ElNotification({
      title: 'ERROR',
      message: '数据获取失败，请返回或刷新页面',
      duration: 4500,
  })
      return false
  }
  // console.log(data)
  router.push({
      name: 'airConditioner',
      query: {
      trainNo:props.trainId,
      trainCoach:carriageID
      }
  })
}
function initTrainSelection() {
  trainSelectionData.value.forEach((item,index)=>{
    item.trainNo = props.trainId
    item.carriage_no = props.trainId+'0'+(index+1)
  })
}

onMounted(() => {
    // initWebSocket()
  initTrainSelection()
  getapi()
})
let time = setInterval(() => {
   getapi()
   console.log('请求')
  }, 1000*60*2);
  onUnmounted(()=>{
    clearInterval(time)
})
</script>

<style scoped lang="scss">
:deep(.el-divider .el-divider--horizontal){
    margin: 5px 0!important;
}

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
    border-bottom: 1px solid #394153 !important;
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
.container {
    display: flex;
    align-items: center;

    .c-l {
        text-align: center;
        width: 100px;
    }

    .c-r {
        flex-grow: 1;
        // background: gray;
        box-sizing: border-box;
        width: calc(100% - 100px);
        overflow: hidden;
        padding: 20px;

        .carWarpper {
            display: flex;
            justify-content: center;
            gap: 4px;
            text-align: center;

            .carItem {
                padding-top: 13px;
                cursor: pointer;
                border-radius: 10px;
                &:hover {
                    background: rgba($s-primary, .2)
                }

                img {
                    width: 100%;
                    max-width: 180px;
                }
                .itemTitle {margin-bottom: 10px;}
                .itemAirCondition {
                    .el-space__item {
                        margin-right: 4px !important;
                    }
                }
            }
        }
    }
}
</style>
