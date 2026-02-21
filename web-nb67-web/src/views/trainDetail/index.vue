<template>
<div class="warp">
  <div class="header">
    <BackTop />
    <div style="line-height: 50px;font-weight: bold;">当前列车号：{{ route.params.trainNo || route.query.trainNo }}</div>
    <div class="noimportant"></div>
  </div>
  <div class="air-wrapper">
    <!-- 空调系统 -->
    <div class="state">
      <div class="left">
        <div class="titleheader">
          <h4>空调系统</h4>
          <el-popover placement="right" :width="200" popper-class="popper-help">
            <template #reference>
              <img src="@/assets/img/help.png" width="15" height="15" @click="openHelp()" />
            </template>
            <el-row :gutter="20" align="middle">
              <el-col :span="12">正常</el-col>
              <el-col :span="12" style="text-align: right">
                <el-icon color="#42ad5d" size="20px">
                  <SuccessFilled />
                </el-icon>
              </el-col>
            </el-row>
            <el-divider style="width: 100%" />
            <el-row :gutter="20" align="middle">
              <el-col :span="12">故障</el-col>
              <el-col :span="12" style="text-align: right">
                <el-icon color="#e65355" size="20px">
                  <CircleCloseFilled />
                </el-icon>
              </el-col>
            </el-row>
            <el-divider />
            <el-row :gutter="20" align="middle">
              <el-col :span="12">预警</el-col>
              <el-col :span="12" style="text-align: right">
                <el-icon color="#ffa55c" size="23px">
                  <WarnTriangleFilled />
                </el-icon>
              </el-col>
            </el-row>
          </el-popover>
        </div>
        
      </div>
      <!-- <div class="right">
        <ul class="ul">
          <li v-for="(title, index) in titleList" :key="index" class="title-item">{{ title }}</li>
        </ul>

        <ul class="ul">
          <li
            :class="['img-item', backgroundImg(index)]"
            v-for="(item, index) in trainStateList"
            :key="index"
          >
            <img :src="getIcon(item, index, 'left')" @click="gotoPath(index + 1, 1)" />
            <img :src="getIcon(item, index, 'right')" @click="gotoPath(index + 1, 2)" />
          </li>
        </ul>
      </div> -->
      <div class="system">
        <div class="systeminformation">
          <span>标题1</span>
          <span class="information">XXXXXXXXXXXXX</span>
        </div>
        <div class="systeminformation">
          <span>标题1</span>
          <span class="information">XXXXXXXXXXXXX</span>
        </div>
      </div>
    </div>
    <!-- 运行状态信息 -->
    <div class="state">
      <div class="left">
        <div class="titleheader">
          <h4>运行状态信息</h4>
        </div>
      </div>
      <!-- <div class="right">
        <ul class="ul ul2">
          <li v-for="(item, index) in trainStateList2" :key="index" :class="reverseClass(index)">
            <div class="div" @click="gotoPath(index + 1, 1)">
              <div class="top">1号机组</div>
              <div class="bottom">{{ getAirState(item.last_wOpModeU1) }}</div>
            </div>
            <div class="div" @click="gotoPath(index + 1, 2)">
              <div class="top">2号机组</div>
              <div class="bottom">{{ getAirState(item.last_wOpModeU2) }}</div>
            </div>
          </li>
        </ul>
      </div> -->
      <div class="system">
        <div 
          class="systeminformation" 
          @click="gotoPathcarCrewDetail()" 
          v-for="i in 5" :key="i"
        >
          <span>{{0+i}}号机组</span>
          <span class="information">26°C</span>
        </div>
      </div>
      <div class="m-l-20 m-r-20">
        <RunState :tableData="runStateDate" />
      </div>
    </div>
    <div class="state">
      <div class="left">
        <div class="titleheader">
          <h4>车厢温度信息</h4>
       </div>
      </div>
      <div class="m-l-20 m-r-20 m-t-20">
        <Temperature :tableData="temperatureData" />
      </div>
    </div>
    <!-- <Alarm title="报警信息" :data="eaData" @handleChangePage="handleChangePageEa" />
    <Alarm title="预警信息" :data="paData" @handleChangePage="handleChangePagePa" /> -->
  </div>
</div>
</template>

<script setup>
import {
  getAirStartDataApi,
  getTrainDataApi,
  getEaDataBydvcAddrApi,
} from "@/api/api.js";
import Alarm from "./Alarm.vue";
import RunState from "./RunState.vue";
import Temperature from "./Temperature.vue";
import BackTop from "@/components/BackTop.vue";
import { useRoute, useRouter } from "vue-router";
import { ref, onMounted, provide } from "vue";

let route = useRoute();
let router = useRouter();
let titleList = ["TC1", "MP1", "M1", "M2", "MP2", "TC2"];
let trainStateList = ref([]);
let trainStateList2 = ref([]);
let runStateDate = ref([]);
let temperatureData = ref([]);

let eaData = ref([]); //告警信息
let paData = ref([]); //预警信息
//Alarm参数
let trainId = "";
let defaultTime = "7d";

let page = 1;
let limit = 10;

const setArr = (obj) => {
  let arr = [];
  for (let i in obj) {
    arr.push({ U1: obj[i].U1, U2: obj[i].U2 });
  }
  return arr;
};

const setRunStateData = (arrData) => {
  let names = [
    { name: "压缩机1", EN: "last_cFBK_Comp_U", NO: "1" },
    { name: "压缩机2", EN: "last_cFBK_Comp_U", NO: "2" },
    { name: "通风机", EN: "last_cFBK_EF_U" },
    { name: "冷凝风机", EN: "last_cFBK_CF_U" },
    { name: "新风阀开度", EN: "last_wPosFAD_U" },
    { name: "回风阀开度", EN: "last_wPosRAD_U" },
  ];
  let arr = names.map((item) => {
    let { name, EN, NO } = item;
    NO = NO ? NO : "";
    return {
      name: name,
      equ1_1: arrData[0][`${EN}1${NO}`],
      equ1_2: arrData[0][`${EN}2${NO}`],
      equ2_1: arrData[1][`${EN}1${NO}`],
      equ2_2: arrData[1][`${EN}2${NO}`],
      equ3_1: arrData[2][`${EN}1${NO}`],
      equ3_2: arrData[2][`${EN}2${NO}`],
      equ4_1: arrData[3][`${EN}2${NO}`],
      equ4_2: arrData[3][`${EN}1${NO}`],
      equ5_1: arrData[4][`${EN}2${NO}`],
      equ5_2: arrData[4][`${EN}1${NO}`],
      equ6_1: arrData[5][`${EN}2${NO}`],
      equ6_2: arrData[5][`${EN}1${NO}`],
    };
  });
  return arr;
};

const setTemperatureData = (arrData) => {
  let names = [
    { name: "客室温度", EN: "last_iInnerTemp" },
    { name: "新风温度", EN: "last_iOuterTemp" },
    { name: "目标温度", EN: "last_iSetTemp" },
    // { name: "车厢温度1", EN: "last_iSeatTemp" },
    // { name: "车厢温度2", EN: "last_iVehTemp" },
  ];
  let arr = names.map((item) => {
    let { name, EN } = item;
    return {
      name: name,
      TC1: arrData[0][`${EN}`] / 10 + "℃",
      MP1: arrData[1][`${EN}`] / 10 + "℃",
      M1: arrData[2][`${EN}`] / 10 + "℃",
      M2: arrData[3][`${EN}`] / 10 + "℃",
      MP2: arrData[4][`${EN}`] / 10 + "℃",
      TC2: arrData[5][`${EN}`] / 10 + "℃",
    };
  });
  return arr;
};

const setAlarmData = (data) => {
  let arr = data.data;
  let list = arr.map((item, index) => {
    return {
      train_id: item.train_id,
      type: item.type,
      startTime: item.starttime_str,
      endTime: item.endtime_str,
      desc: item.desc,
      Advice: item.Advice,
      level: item.level ? item.level : null,
      sig_list: JSON.parse(item.sig_list),
      id:item.id
    };
  });

  return {
    count: data.count,
    list,
  };
};

const backgroundImg = (index) => {
  if (index == 0) {
    return "header-bg";
  } else if (index == 5) {
    return "footer-bg";
  } else {
    return "";
  }
};

const reverseClass = (index) => {
  if (index >= 3) {
    return "row-reverse";
  }
};

// const leftIcon = (state) => {
//   let str = "";
//   switch (state) {
//     case "normal":
//       str = "left1";
//       break;
//     case "warning":
//       str = "left2";
//       break;
//     case "alarm":
//       str = "left3";
//       break;
//     default:
//       str = "left1";
//   }
//   return new URL(`/src/assets/img/${str}.svg`, import.meta.url).href;
// };

// const rightIcon = (state) => {
//   let str = "";
//   switch (state) {
//     case "normal":
//       str = "right1";
//       break;
//     case "warning":
//       str = "right2";
//       break;
//     case "alarm":
//       str = "right3";
//       break;
//     default:
//       str = "right1";
//   }
//   return new URL(`/src/assets/img/${str}.svg`, import.meta.url).href;
// };

// const getIcon = (obj, index, direction) => {
//   let url = "";
//   if (index < 3) {
//     if (direction == "left") {
//       url = leftIcon(obj.U1);
//     } else {
//       url = rightIcon(obj.U2);
//     }
//   } else {
//     // 后三位排列顺序相反
//     if (direction == "left") {
//       url = leftIcon(obj.U2);
//     } else {
//       url = rightIcon(obj.U1);
//     }
//   }
//   return url;
// };

// const gotoPath = (num, crew) => {
//   let params = {
//     trainNum: num,
//     trainCrew: crew
//   }
//   router.push({
//     name: "airConditioner",
//     params
//   });
// };

const getAirState = (state) => {
  switch (state) {
    case 0:
      return "停机";
    case 5:
      return "通风";
    case 6:
      return "弱冷";
    case 7:
      return "强冷";
    case 8:
      return "弱暖";
    case 9:
      return "强暖";
    case 10:
      return "紧急通风";
    case 12:
      return "除霜";
    default:
      return "停机";
  }
};

const getEaData = (params) => {
  getEaDataBydvcAddrApi({
    FL_dvcTrain: trainId,
    type: "ea",
    page,
    limit,
    ...params,
  }).then((res) => {
    eaData.value = setAlarmData(res.data);
    // console.log(eaData)
  });
};

const getPaData = (params) => {
  getEaDataBydvcAddrApi({
    FL_dvcTrain: trainId,
    type: "pa",
    page,
    limit,
    ...params,
  }).then((res) => {
    paData.value = setAlarmData(res.data);
    // console.log(paData)
  });
};

const chooseDay = (title, dateParams) => {
  if (title == "报警信息") {
    getEaData(dateParams);
  } else if (title == "预警信息") {
    getPaData(dateParams);
  }
};
provide("provideChooseDay", chooseDay); //传递给TitleHeader组件,TitleHeader进行回调执行

// const handleChangePageEa = (num) => {
//   page = num;
//   getEaData();
// };

// const handleChangePagePa = (num) => {
//   page = num;
//   getPaData();
// };

const gotoPathcarCrewDetail = () => {
  router.push({
    name: 'carCrewDetail',
  })
};

onMounted(() => {
  trainId = route.params.trainNo || route.query.trainNo;
  //空调状态信息
  getAirStartDataApi({ FL_dvcTrain: trainId }).then((res) => {
    let arr = setArr(res.data);
    trainStateList.value = arr;
  });

  // 车厢数据
  getTrainDataApi({ FL_dvcTrain: trainId }).then((res) => {
    trainStateList2.value = res.data;
    runStateDate.value = setRunStateData(res.data); //运行状态信息
    temperatureData.value = setTemperatureData(res.data); // 车厢温度信息
  });

  // 告警信息
  getEaData({ defaultTime });

  // 预警信
  getPaData({ defaultTime });
});
</script>

<style lang="scss" scoped>
// 整个模块背景色
// .warp {
//   width: 100%;
//   border-radius: 10px;
//   background-color: #171c32;
//   margin-bottom: 15px;
// }
.warp {
  background: #181F30;
  border-radius: 20px;
  border: 1px solid #3a3b68
}
.header {
  box-sizing: border-box;
  height: 60px;
  display: flex;
  justify-content: space-between;
  padding-left: 25px;
  padding-right: 25px;
}
.titleheader {
  height: 50px;
  box-sizing: border-box;
  top: 0;
  left: 0;
  display: flex;
  align-items: center;
  gap: 0.5em;
}
.titleheader > h4 {
  font-size: 16px;
  position: relative;
  line-height: 1em;
}
.titleheader > h4::before {
  position: absolute;
  content: "";
  width: 4px;
  height: 1em;
  background: #fff;
  margin-left: -10px;
}

.air-wrapper {
  box-sizing: border-box;
  padding-bottom: 10px;
  width: 100%;
  font-size: 14px;
  .row-reverse {
    flex-flow: row-reverse;
  }
  .bgc {
    background-color: #2a3143;
    position: relative;
    > img {
      display: block;
      position: absolute;
      right: 10px;
      top: 50%;
      transform: translateY(-50%);
    }
  }
  .state {
    padding-bottom: 15px;
    border-bottom: 1px solid #3a3b68;
    .left {
      // width: 100px;
      height: 30px;
      line-height: 30px;
      margin: 10px 30px auto 0;
      display: flex;
      align-items: center;
      margin-left: 40px;
    }
    .right {
      flex: 1;
    }
  }

  .ul {
    margin-bottom: 10px;
    font-size: 12px;
    display: flex;
    justify-content: space-around;
    .title-item {
      width: 60px;
      height: 20px;
      text-align: center;
      line-height: 20px;
      border-radius: 10px;
      background-color: #2a3143;
    }
    .img-item {
      width: 100%;
      height: 55px;
      background-image: url("@/assets/img/center.png");
      background-size: 100% 100%;
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0 10px 0 20px;
      img {
        width: 30%;
        height: 70%;
        cursor: pointer;
      }
    }
    .header-bg {
      background-image: url("@/assets/img/header.png");
      padding: 0 12px 0 28px;
    }
    .footer-bg {
      padding: 0 28px 0 20px;
      background-image: url("@/assets/img/footer.png");
    }
  }

  .ul2 {
    li {
      display: flex;
      justify-content: space-between;
      padding: 0 10px 0 20px;
      flex: 1;
      .div div {
        width: 62px;
        height: 24px;
        line-height: 24px;
        font-size: 12px;
        text-align: center;
        border: 1px solid #fc791f;
        cursor: pointer;
        &.bottom {
          background: #fc791f;
        }
      }
    }
    li:first-child {
      padding: 0 10px 0 28px;
    }
    li:last-child {
      padding: 0 28px 0 20px;
    }
  }
}

// 返回样式
:deep(.header[data-v-e75009dc]) {
  background: none !important;
  padding-top: 10px;
  padding-bottom: 10px;
  border-bottom: 1px solid #3a3b68;
}
// ↓这个是什么不重要
.noimportant{
  background: none;
}
// 系统信息
.system {
  display: flex;
  box-sizing: border-box;
  padding: 10px 20px 20px;
}
.systeminformation {
  display: flex;
  flex-direction: column;
  margin-top: 10px;
  margin-left: 10px;
  margin-right: 20px;
}
.information {
  color: #aaa;
}
.systeminformation {
  cursor: pointer;
}
.systeminformation:hover,.systeminformation:hover .information {
  color: #528fff!important;
}
// 边距
.m-l-10{
  margin-left: 10px;
}
.m-l-20{
  margin-left: 20px;
}
.m-r-20{
  margin-right: 20px;
}
.m-b-20{
  margin-bottom: 20px;
}
.m-t-20{
  margin-top: 20px;
}
</style>

<style scoped>
/* :deep(thead) {
  background: #2b3548;
} */
:deep(.btn_text) {
  color: #2186CF !important;
}
:deep(.title) {
  width: 100%;
  height: 40px;
  line-height: 40px;
  padding-left: 20px;
  box-sizing: border-box;
  border-bottom: 2px solid #1f273c;
  margin-bottom: 10px;
  font-size: 13px;
  font-weight: bold;
}
:deep(.el-table) {
  --el-table-border-color: none;
}
/* nodate table style */
:deep(.el-table__empty-text) {
  margin-top: 15px;
  line-height: 20px !important;
}
</style>
<style>
.popper-help {
  background: #2b3548 !important;
  border: 1px solid #1f273c !important;
  color: #fff !important;
}
.el-divider--horizontal {
  margin: 5px 0;
  border-top: 1px solid #4d535f !important;
}
</style>

