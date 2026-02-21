<template>
  <div class="title-header">
    <div class="left">
      <div class="header">
        <h4>{{ props.title }}</h4>
      </div>
    </div>
    <div class="right">
      <el-button
        type="text"
        :class="['btn_text1', day == '7d' ? 'btn_text' : '']"
        @click="chooseDay('7d')"
        >近7天</el-button
      >
      <el-button
        type="text"
        :class="['btn_text1', day == '30d' ? 'btn_text' : '']"
        @click="chooseDay('30d')"
        >近30天</el-button
      >
      <el-button
        type="text"
        :class="['btn_text1', day == '365d' ? 'btn_text' : '']"
        @click="chooseDay('365d')"
        >近12个月</el-button
      >&nbsp;&nbsp;&nbsp;
      <el-date-picker
        v-model="dateValue"
        type="daterange"
        range-separator="~"
        format="YYYY/MM/DD"
        value-format="YYYY-MM-DD"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        @change="dateChange"
      ></el-date-picker>
      <!-- <el-button type="text" class="btn_text" @click="reset">
        &nbsp;&nbsp;重置
      </el-button> -->
    </div>
  </div>
</template>

<script setup>
import { ref, inject, onMounted } from "vue";

let dateValue = ref("");
let day = ref("7d");

let SonChangeDayfn = inject("provideChooseDay");

const props = defineProps({
  title: {
    type: String,
  },
});
const reset = () => {
  dateValue.value = "";
  day.value = "7d";
  SonChangeDayfn(props.title, { defaultTime: "7d" });
};
const chooseDay = (date) => {
  console.log(date);
  day.value = date;
  if(date=='7d'){
    var dates = new Date();
    const startTime = getData(-7) + ' 00:00:00'
    const endTime = dates.getFullYear() + '-' + (dates.getMonth() + 1) + '-' + dates.getDate() + " 23:59:59"
    SonChangeDayfn(props.title, { startTime,endTime });
  } else if(date=='30d'){
     var dates = new Date();
    const startTime = getData(-30) + ' 00:00:00'
    const endTime = dates.getFullYear() + '-' + (dates.getMonth() + 1) + '-' + dates.getDate() + " 23:59:59"
    SonChangeDayfn(props.title, { startTime,endTime });
  } else {
     var dates = new Date();
    const startTime = getData(-365) + ' 00:00:00'
    const endTime = dates.getFullYear() + '-' + (dates.getMonth() + 1) + '-' + dates.getDate() + " 23:59:59"
    SonChangeDayfn(props.title, { startTime,endTime });
  }
};

const dateChange = () => {
  // 判断日期是否为空
  if (dateValue.value != null) {
    day.value = "";
    console.log();
    let params = {
      startTime: dateValue.value[0]+" 00:00:00",
      endTime: dateValue.value[1]+" 23:59:59",
    };
    console.log(params);
    SonChangeDayfn(props.title, params);
  } else {
    // 代表日期已经清空
    dateValue.value = "";
    day.value = "7d";
    var date = new Date();
    const startTime = getData(-7) + ' 00:00:00'
    const endTime = date.getFullYear() + '-' + (date.getMonth() + 1) + '-' + date.getDate() + " 23:59:59"
    SonChangeDayfn(props.title, { startTime,endTime });
  }
};
function getData(day){
    var today=new Date()
    var targetday=today.getTime() +1000*60*60*24* day
    today.setTime(targetday)
    var tYear=today.getFullYear()
    var tMonth=today.getMonth()
    var tDate=today.getDate()
    tMonth=doHandMonth(tMonth+1)
    tDate=doHandMonth(tDate)
    return tYear +"-" + tMonth+"-"+tDate
}
  
  
function doHandMonth(month){
    var m=month
    if(month.toString().length==1){
    m="0"+month
    }
    return m
}
onMounted(()=>{
  var date = new Date();
  const startTime = getData(-7) + ' 00:00:00'
  const endTime = date.getFullYear() + '-' + (date.getMonth() + 1) + '-' + date.getDate() + " 23:59:59"
  SonChangeDayfn(props.title, { startTime,endTime });
})
</script>

<style scoped>
.title-header {
  width: 100%;
  height: 60px;
  padding: 0 30px 0 10px;
  box-sizing: border-box;
  font-size: 16px;
  border-bottom: 2px solid #1f273c;
  margin-bottom: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center
}
.btn_text {
  color: #2186CF !important;
}

.btn_text1 {
  color: rgba(255, 255, 255, 0.815);
}

/* .left {
  float: left;
  margin-left: 10px;
  margin-top: 28px;
}
*/
.right {
  display: flex;
  gap: 20px;
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
</style>
<style>
/* 时间弹层 */
.el-picker-panel__body-wrapper {
  background-color: #fff !important;
}
/* 标题左边线 */
.header {
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
