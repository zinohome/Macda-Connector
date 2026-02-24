<template>
  <div class="warp">
    <!-- <div class="header">
      <div class="img">
        <img src="/img/kongtiao.png" />
      </div>
      <div class="line">
          5号线
      </div>
    </div> -->
    <el-row class="content" >
      <el-col class="header">
        <h4>空调系统</h4>
        <!-- <div>
          <el-select v-model="value" placeholder="请选择">
            <el-option v-for="item in options" :key="item.value" :label="item.label" :value="item.value">
            </el-option>
          </el-select>
        </div> -->
      </el-col>
      <el-col :span="24" class="svgBox">
        <!-- <div class="top">
          <div class="text">{{ props.centerData.online }}/26</div>
        </div> -->
        <div class="svgBoxContent">
          <div class="svgTitle">
            <div class="left">
              <p class="title">在线正常列车</p>
              <div class="text">
                <span class="success">{{ props.centerData.normal }}</span>/{{ props.centerData.totalnumber }}
              </div>
            </div>
            <div class="mid">
              <p class="title">在线预警列车</p>
              <div class="text">
                <span class="warr">{{ props.centerData.warning_count }}</span>/{{ props.centerData.totalnumber }}
              </div>
            </div>
            <div class="right">
              <p class="title">在线报警列车</p>
              <div class="text">
                <span class="error">{{ props.centerData.alarm }}</span>/{{ props.centerData.totalnumber }}
              </div>
            </div>
          </div>
          <div class="svgBg">
            <img src="/img/bg3.svg" />
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  centerData: {
    type: Object,
    default: {
      online: 0, normal: 0, paNum: 0, eaNum: 0
    }
  }
})

// const options = [{
//     value: '1',
//     label: '1号线'
//   }, {
//     value: '2',
//     label: '2号线'
//   }, {
//     value: '3',
//     label: '3号线'
//   }, {
//     value: '4',
//     label: '4号线'
//   }, {
//     value: '5',
//     label: '5号线'
//   }]

// const value = '1'

</script>

<style scoped lang="scss">
.content {
  height: 703px;
  position: relative;
}
.content > .header {
  position: absolute;
  width: 100%;
  height: 50px;
  box-sizing: border-box;
  top: 0;
  left: 0;
  display: flex;
  align-items: center;
  gap: 30px;
}
.content > .header > h4 {
  font-size: 16px;
  position: relative;
  margin-left: 30px;
  line-height: 1em;
}
.content>.header>h4::before {
  position: absolute;
  content: "";
  width: 4px;
  height: 1em;
  background: #fff;
  margin-left: -10px;
}
.success {
  color: #42ad5d;
}
.warr {
  color: #ffa55c;
}
.error {
  color: #e65355;
}
.svgBox {
  display: flex;
  justify-content: center;
  align-items:center
}
.svgBoxContent .svgTitle {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}
.svgBox .svgBoxContent {
  box-sizing: border-box;
  width: 100%;
  max-width: 880px;
  max-height: 650px;
  position: relative;
  text-align: center;
}
.svgBox .svgBoxContent .svgBg > img {
  max-height: 650px;
  -webkit-user-drag: none;
}
.left,
.mid,
.right {
  position: absolute;
  color: #fff;
  text-align: center;
  font-size: 90%;
  background: rgba(255, 255, 255, .5);
  background: rgba(34, 83, 231, .8);
  background: rgba(#006AC1,.8);
  box-shadow: -4px -4px 0 rgba(14, 20, 138, 1);
  box-shadow: -4px -4px 0 rgba(21, 26, 123,.8);
    box-shadow: -4px -4px 0 rgba(#006AC1, .2);
  padding: 0.5em 1em;
  border-radius: 0.3em;
  // transform: skew(0deg, -30deg); 
}

.left>.text,
.mid>.text,
.right>.text {
  font-size: 130%;
  font-weight: 900;
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif
}

.left {
  top: 28%;
  left: 8%;
}

/* .left::after,.mid::after,.right::after {
  position: absolute;
  content: "";
  width: 10px;
  height: 50%;
  left: 50%;
  top: 100%;
  transform: translateX(-50%);
  z-index: 999;
  background: rgba(34, 83, 231, .8);
  box-shadow: -4px -4px 0 rgba(21, 26, 123, .8);
  border-bottom-left-radius: 0.3em;
  border-bottom-right-radius: 0.3em;
} */

.mid {
  top: 61%;
  left: 21%;
}

.right {
  top: 70%;
  left: 65%;
}
:deep(.el-input__inner,.el-input__inner .el-input__inner:focus) {
  background-color: #181F30;
  border: 1px solid #394153;
  box-shadow: 0 0 1px 0 #394153;
  color: #fff!important
}
:deep(.el-input__wrapper) {
  background-color: #181F30 !important;
  border: 1px solid #394153;
  box-shadow: 0 0 1px 0 #394153;
  color: #fff !important;
  .el-input__inner {
    border: none !important;
    color: #fff !important;
  }
}
</style>