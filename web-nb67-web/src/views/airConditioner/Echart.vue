<template>
    <div class="warp">
        <div class="header">
            <h4>温度趋势</h4>
            <div class="right">
      <el-button
        type="text"
        :class="['btn_text1', day == 'hourly' ? 'btn_text' : '']"
        @click="chooseDay('hourly')"
        >时</el-button
      >
      <el-button
        type="text"
        :class="['btn_text1', day == 'daily' ? 'btn_text' : '']"
        @click="chooseDay('daily')"
        >天</el-button
      >
      <el-button
        type="text"
        :class="['btn_text1', day == 'monthly' ? 'btn_text' : '']"
        @click="chooseDay('monthly')"
        >月</el-button
      >&nbsp;&nbsp;&nbsp;
    </div>
        </div>
        <div class="content" style="padding:0 15px">
            <div class="echart2" id="airEChart" ref="airEChart" v-if="showEchart"></div>

            <div style="padding:30px 0;text-align:center;" v-else>
                <img src="/src/assets/img/no-data.svg" width="40" height="50" />
                <p style="font-size:12px">当前暂无数据</p>
            </div>
        </div>
    </div>
</template>

<script setup>
import { getThDataByDvcAddrApi } from '@/api/api.js'
import { ref, onMounted, watchEffect, nextTick,defineExpose } from 'vue'
import * as echarts from 'echarts'
let day = ref("hourly");
let showEchart = ref(false)
    function getLocalTime(i,data) { 
    if (typeof i !== 'number') return;
 
   var d = new Date(data); 
//   2023-02-20T13:50:20.954Z
//   2023-02-08T00:00:00+00:00
    //得到1970年一月一日到现在的秒数 
    var len = d.getTime();
    
    //本地时间与GMT时间的时间偏移差(注意：GMT这是UTC的民间名称。GMT=UTC）
    var offset = d.getTimezoneOffset() * 60000;
 
    //得到现在的格林尼治时间
    var utcTime = len + offset;
 
    return new Date(utcTime + 3600000 * i);
}
function newDate(time) {
    let temp = time.split('T')
    console.log(temp);
    const year = temp[0].slice(0,4)
    const month = temp[0].slice(5,7)
     const day = temp[0].slice(8,10)
     const times = temp[1].slice(0,8)
     return `${year}-${month}-${day}-${times}`
    // var date = new Date(time)
    // var y = date.getFullYear()
    // var m = date.getMonth() + 1
    // m = m < 10 ? '0' + m : m
    // var d = date.getDate()
    // d = d < 10 ? '0' + d : d
    // var h = date.getHours()
    // h = h < 10 ? '0' + h : h
    // var minute = date.getMinutes()
    // minute = minute < 10 ? '0' + minute : minute
    // var s = date.getSeconds()
    // s = s < 10 ? '0' + s : s
    // return y + '-' + m + '-' + d + ' ' + h + ':' + minute + ':' + s
}
const props = defineProps({
    carID: {
        type: String
    }
})
const chooseDay = (date) => {
    day.value = date;
    initEchart(props.carID,'res'+date)
}
const setEchartData = (arrData) => {
    let obj = {
        iInnerTemp: [],     //客室温度
        iOuterTemp: [],     //新风温度
        iSetTemp: [],       //目标温度
        time: []
    }
    console.log(arrData);
    arrData.map(item => {
        // obj.iInnerTemp.push((item.dvc_i_veh_temp / 10).toFixed(2))
        // obj.iOuterTemp.push((item.dvc_i_outer_temp / 10).toFixed(2))
        // obj.iSetTemp.push((item.dvc_i_set_temp / 10).toFixed(2))
        obj.iInnerTemp.push(item.ras_sys)
        obj.iOuterTemp.push(item.fas_sys)
        obj.iSetTemp.push(item.tic)
        // 处理时间
        // let d = item.time.split("+")
        // let time = [];
        // time.push(item.bucket)
        // console.log(time,item.bucket);
        // if (d.length > 0) {
            // item.bucket = item.bucket.split("+")[0];
            // item.bucket = item.bucket.split("T");
        // }

        // 原始数据
        obj.time.push(newDate(item.bucket))
        // obj.time.push(item.bucket)

    })
    return obj
}

const initEchart = (carID,data) => {
    console.log('进来');
    if(!carID || carID === "") {
      return
    }
    getThDataByDvcAddrApi({ carriageNo: carID, limit: 5 }).then(res => {
        if (res.length == 0) {
            return
        } else {
            showEchart.value = true
        }

        nextTick(() => {
            let airEChart = ref()
            airEChart.value = document.getElementById('airEChart')
            let echartData
            if(day.value == 'daily'){
                res.daily = res.daily.sort((a, b) => new Date(a.bucket) - new Date(b.bucket));
                echartData = setEchartData(res.daily)
            } else if(day.value == 'hourly') {
                res.hourly = res.hourly.sort((a, b) => new Date(a.bucket) - new Date(b.bucket));
                echartData = setEchartData(res.hourly)
            } else {
                res.monthly = res.monthly.sort((a, b) => new Date(a.bucket) - new Date(b.bucket));
                echartData = setEchartData(res.monthly)
            }

            let option = {
                tooltip: {
                    trigger: 'axis',
                    formatter: function (params) {
                        let relVal = params[0].name
                        for (let i = 0; i < params.length; i++) {
                            relVal += '<br/>' + params[i].marker + params[i].seriesName + ' : ' + params[i].value + '℃'
                        }
                        return relVal
                    }
                },
                grid: {
                    x: 50,
                    y: 50,
                    x2: 30,
                    y2: 60,
                    borderWidth: 10
                },

                xAxis: {
                    data: echartData.time,  //set
                    axisLabel: {
                        textStyle: {
                            color: '#fff'
                        }
                    }
                },
                legend: {
                    data: ['客室温度', '新风温度', '目标温度'],
                    textStyle: {
                        fontSize: 14,
                        color: '#ffffff'
                    }
                },
                yAxis: {
                    splitLine: {
                        show: true,
                        lineStyle: {
                            type: 'dashed'
                        }
                    },
                    axisLabel: {
                        textStyle: {
                            color: '#fff'
                        }
                    }
                },
                series: [
                    {
                        name: "客室温度",
                        type: "line",
                        data: echartData.iInnerTemp //set
                    },
                    {
                        name: "新风温度",
                        type: "line",
                        data: echartData.iOuterTemp //set
                    },
                    {
                        name: "目标温度",
                        type: "line",
                        data: echartData.iSetTemp //set
                    }
                ]
            }

            let myChart2 = echarts.init(airEChart.value);

            myChart2.setOption(option);
            window.onresize = function () {
                myChart2.resize();
            };
        })

    })
}
defineExpose({
    initEchart
})
onMounted(() => {
    // initEchart(props.carID,'res.hourly')
})

watchEffect(() => {
    // initEchart(props.carID,'res.hourly')
})


</script>

<style scoped>
.warp {
    width: 100%;
    border-radius: 10px;
    background-color: #171c32 !important;
    margin-bottom: 15px;
}

.echart2 {
    width: 100%;
    height: 400px;
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
.content {
    border-top: 1px solid #394153
}
</style>
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
