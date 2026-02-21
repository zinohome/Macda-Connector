<script setup>
</script>
<template>
   <router-view />
</template>
<style lang="scss">
$s-primary: #2186CF;
$s-danger: #e65355;
$s-warning: #ffa55c;
$s-info: #bfbfbf;
$s-success: #42ad5d;
$s-yellow: #ffe375;
$s-white: #ffffff;
#app {
   box-sizing: border-box;
   color: #fff;
}

.el-table th {
   color: #fff;
   font-size: 14px;
   background-color: transparent !important;
   /* border: 0; */
   height: 30px;
   line-height: 30px;
   border-bottom: 1px solid #394153 !important;
}

.el-table tr,
.el-table td {
   color: #fff;
   font-size: 14px;
   background-color: transparent !important;
   /* border: 0; */
   height: 30px;
   line-height: 30px;
}

.el-table::before {
   left: 0;
   bottom: 0;
   width: 100%;
   height: 0px !important;
}
/* 表格内容边距 */
.el-table__body {
   padding-top: 10px;
   padding-bottom: 10px;
}
/* 调整所有表格行高 */
.el-table .el-table__cell {
   padding-left: 10px !important;
}

.el-table .cell {
   padding: 0 5px!important;
}
/* 表格标题底部线 */
.el-table__header-wrapper {
   border-bottom: 1px solid #394153 !important;
}

.el-picker-panel {
   color: #606266;
   box-sizing: border-box;
   border: 1px solid #394153 !important;
   box-shadow: 0 2px 12px 0 rgb(0 0 0 / 10%);
   background: #181F30 !important;
   border-radius: 4px;
   line-height: 30px;
   /* margin: 5px 0; */
}
.el-input__inner {
   background-color: #20283C !important;
   border: 1px solid #394153 !important;
   color: #fff;
}
.el-date-editor .el-range-input {
   background: none !important;
}
.el-date-editor .el-range-separator {
   font-size: 12px;
   color: #a8abb2 !important;
}
.el-picker-panel__icon-btn {
   color: #2186CF !important;
}
.el-select-dropdown__item.selected {
    color: #2186CF;
}
/* .el-scrollbar__wrap {
   display: flex;
   align-items: center;
} */

/*定义滚动条高宽及背景 高宽分别对应横竖滚动条的尺寸*/
::-webkit-scrollbar {
   width: 7px;
   height: 7px;
   background-color: #f5f5f5;
   cursor: pointer;
   border-radius: 10px;
}

/*定义滚动条轨道 内阴影+圆角*/
::-webkit-scrollbar-track {
   box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.3);
   -webkit-box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.3);
   border-radius: 10px;
   background-color: #f5f5f5;
   cursor: pointer;
}

/*定义滑块 内阴影+圆角*/
::-webkit-scrollbar-thumb {
   border-radius: 10px;
   box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.1);
   -webkit-box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.1);
   background-color: rgb(136, 136, 136);
   cursor: pointer;
}
.el-popover{
   white-space: pre-line !important;
   line-height: 12px !important;
}
</style>
