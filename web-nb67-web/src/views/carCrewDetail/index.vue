<template>
<div class="warp">
  <div class="header">
    <BackTop />
    <div class="titlename">
      <div style="line-height: 30px;font-weight: bold;">当前列车号：{{ route.params.trainNo || route.query.trainNo }}</div>
      <div class="samlltitle">号机组</div>
    </div>
    <div class="noimportant"></div>
  </div>
  <div class="air-wrapper">
    <div class="state">
      <div class="left">
        <div class="titleheader">
          <h4>健康评估信息</h4>
        </div>
      </div>
      <div class="m-l-20 m-r-20 m-t-20">
        <Health :tableData="HealthData" />
        <Temperature :tableData="TemperatureData" />
        <System :tableData="SystemData" />
      </div>
        
    </div>
  </div>
</div>
</template>

<script setup>
import {
  getAirStartDataApi,
  getTrainDataApi,
  getEaDataBydvcAddrApi,
} from "@/api/api.js";
import Health from "./Health.vue";
import Temperature from "./Temperature.vue";
import System from "./System.vue";
import BackTop from "@/components/BackTop.vue";
import { useRoute, useRouter } from "vue-router";
import { ref, onMounted, provide } from "vue";

let route = useRoute();
let router = useRouter();
let titleList = ["TC1", "MP1", "M1", "M2", "MP2", "TC2"];
let trainStateList = ref([]);
let trainStateList2 = ref([]);
let HealthData = ref([]);
let TemperatureData = ref([]);
let SystemData = ref([]);

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

const setHealthData = (arrData) => {
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

const setTemperatureDataData = (arrData) => {
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

const setSystemDataData = (arrData) => {
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

const leftIcon = (state) => {
  let str = "";
  switch (state) {
    case "normal":
      str = "left1";
      break;
    case "warning":
      str = "left2";
      break;
    case "alarm":
      str = "left3";
      break;
    default:
      str = "left1";
  }
  return new URL(`/src/assets/img/${str}.svg`, import.meta.url).href;
};

const rightIcon = (state) => {
  let str = "";
  switch (state) {
    case "normal":
      str = "right1";
      break;
    case "warning":
      str = "right2";
      break;
    case "alarm":
      str = "right3";
      break;
    default:
      str = "right1";
  }
  return new URL(`/src/assets/img/${str}.svg`, import.meta.url).href;
};

const getIcon = (obj, index, direction) => {
  let url = "";
  if (index < 3) {
    if (direction == "left") {
      url = leftIcon(obj.U1);
    } else {
      url = rightIcon(obj.U2);
    }
  } else {
    // 后三位排列顺序相反
    if (direction == "left") {
      url = leftIcon(obj.U2);
    } else {
      url = rightIcon(obj.U1);
    }
  }
  return url;
};

const gotoPath = (num, crew) => {
  let params = {
    trainNum: num,
    trainCrew: crew
  }
  router.push({
    name: "airConditioner",
    params
  });
};

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

const handleChangePageEa = (num) => {
  page = num;
  getEaData();
};

const handleChangePagePa = (num) => {
  page = num;
  getPaData();
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
    HealthData.value = setHealthData(res.data); // 车厢温度信息
  });

    // 温度参数
  getTrainDataApi({ FL_dvcTrain: trainId }).then((res) => {
    trainStateList2.value = res.data;
    TemperatureData.value = setTemperatureData(res.data); // 车厢温度信息
  });

      // 系统参数
  getTrainDataApi({ FL_dvcTrain: trainId }).then((res) => {
    trainStateList2.value = res.data;
    SystemData.value = setSystemData(res.data); // 系统参数
  });

  // 告警信息
  getEaData({ defaultTime });

  // 预警信
  getPaData({ defaultTime });
});
</script>

<style lang="scss" scoped>
.header {
  box-sizing: border-box;
  height: 60px;
  display: flex;
  justify-content: space-between;
  padding-left: 25px;
  padding-right: 25px;
}
.titleheard > h4 {
  font-size: 16px;
  position: relative;
  line-height: 1em;
}
.titleheard > h4::before {
  position: absolute;
  content: "";
  width: 4px;
  height: 1em;
  background: #fff;
  margin-left: -10px;
}
// 整个模块背景色
.warp {
  background: #181F30;
  border-radius: 20px;
  border: 1px solid #3a3b68
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
:deep(.header) {
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
  margin-left: 20px;
  margin-bottom: 30px;
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

// 页面标题
.titlename{
  display: flex;
  flex-direction: column;
  .samlltitle{
    align-self: center;
    font-size:12px;
    color: #aaa;
  }
}
// 边距
.m-l-10{
  margin-left: 10px;
}
.m-l-30{
  margin-left: 30px;
}
.m-r-30{
  margin-right: 30px;
}
.m-b-20{
  margin-bottom: 20px;
}
</style>

<style scoped>
/* :deep(thead) {
  background: #2b3548;
} */
:deep(.btn_text) {
  font-size: 12px;
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
/* 边距 */
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

.p-l-20{
  padding-left: 20px;
}
.p-r-20{
  padding-right: 20px;
}
</style>

