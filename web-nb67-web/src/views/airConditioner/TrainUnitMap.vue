<template>
    <div class="warp">
        <div class="header">
            <h4>空调机组详情</h4>
        </div>
        <!-- <div id="mapWrapper">
            <canvas ref="canvas"></canvas>
        </div> -->
        <div class="content">
            
            
            <!-- <div class="data" :style="fontSize = normalFontSize">
                <div class="item1">
                    <p class="scalePrice"  data-max-width="200"><span>系统2膨胀阀开度</span><span>21<small>%</small></span></p>
                </div>
            </div> -->
            <trainUnitMapData :crewDetails="props.crewDetails" :UnitNo="props.UnitNo" />

            <!-- <div style="padding:30px 0;text-align:center;">
                <img src="/src/assets/img/no-data.svg" width="40" height="50" />
                <p style="font-size:12px">当前暂无数据</p>
            </div> -->
        </div>
    </div>
</template>

<script setup>
import { getThDataByDvcAddrApi } from '@/api/api.js'
import { ref, onMounted, watchEffect, nextTick } from 'vue'
import trainUnitMapData from "./trainUnitMapData.vue";

let canvas=ref(null);
let showUnitMap = ref(true);
let normalFontSize = ref('16px');
let mapArea = document.getElementById('unitMap')
const props = defineProps({
    carID: {
        type: String
    },
    UnitNo: {
        type: String
    },
    crewDetails: {
        type: Object,
        default: {}
    }
})
console.log(props.crewDetails);
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
    arrData.map(item => {
        // obj.iInnerTemp.push((item.dvc_i_veh_temp / 10).toFixed(2))
        // obj.iOuterTemp.push((item.dvc_i_outer_temp / 10).toFixed(2))
        // obj.iSetTemp.push((item.dvc_i_set_temp / 10).toFixed(2))
        obj.iInnerTemp.push(item.dvc_i_veh_temp)
        obj.iOuterTemp.push(item.dvc_i_outer_temp)
        obj.iSetTemp.push(item.dvc_i_set_temp)
        // 处理时间
        // let d = item.time.split("+")
        // let time = [];
        // time.push(item.bucket)
        // console.log(time,item.bucket);
        // if (d.length > 0) {
            // item.bucket = item.bucket.split("+")[0];
            // item.bucket = item.bucket.split("T");
        // }
        obj.time.push(newDate(item.bucket))
    })
    return obj
}


</script>

<style lang="scss" scoped>
$s-primary: #2186CF;
$s-danger: #e65355;
$s-warning: #ffa55c;
$s-info: #bfbfbf;
$s-success: #42ad5d;
$s-yellow: #ffe375;
$s-white: #ffffff;
$s-card-bg: #181F30;
$s-line: #394153;
$s-font-weight: 500;
$s-font-size: 1em;
$s-map-card-bg: #2b3247;
.warp {
    width: 100%;
    border-radius: 10px;
    background-color: $s-card-bg !important;
    margin-bottom: 15px;
    /* 标题左边线 */
    .header {
        height: 50px;
        box-sizing: border-box;
        top: 0;
        left: 0;
        display: flex;
        align-items: center;
        gap: 30px;
        > h4 {
            font-size: 16px;
            position: relative;
            margin-left: 30px;
            line-height: 1em;
            &::before {
                position: absolute;
                content: "";
                width: 4px;
                height: 1em;
                background: #fff;
                margin-left: -10px;
            }
        }
    }
    .content {
        border-top: 1px solid $s-line;
        padding: 30px;
        .map {
            position: relative;
            // width: 80vw;
            max-width: 100%;
            min-height: 250px;
            // background: pink;
            .data {
                position: absolute;
                margin-top: 50%;
                margin-left: 50%;
                width: 100%;
                height: 100%;
                transform: translate(-50%, -50%);
                // background: rgba($color: $s-white, $alpha: 0.3);
                .item1 {
                    position: absolute;
                    display: inline-block;
                    background: rgba($color: $s-map-card-bg, $alpha: 1);
                    border-radius: 10px;
                    padding: $s-font-size * 0.3 $s-font-size * 0.7;
                    top: 1%;
                    right: 24%;
                    border-top: 2px solid rgba($color: $s-white, $alpha: 0.2);
                    box-shadow: 0 0 5px 0 rgba($color: $s-white, $alpha: 0.2);
                    > p {
                        display: flex;
                        flex-direction: column;
                        > span {
                            &:first-child {
                                font-size: $s-font-size * 0.8;
                                color: rgba($color: $s-white, $alpha: 0.8);
                            }
                            &:last-child {
                                font-size: $s-font-size * 1.2;
                                font-weight: $s-font-weight;
                                color: $s-danger;
                                > small {
                                    margin-left: $s-font-size * 0.2;
                                    font-size: 60%;
                                    color: rgba($color: $s-white, $alpha: 0.5);
                                }
                            }
                        }
                    }
                }
            }
            .bg {
                img {
                    width: 100%;
                }
            }
        }
    }
}

.right {
  display: flex;
  gap: 20px;
} 
</style>
