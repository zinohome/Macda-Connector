<template>
    <div class="back-top" @click="back">
        <el-icon :size="18">
            <arrow-left-bold />
        </el-icon>
        <span>返回</span>
    </div>
</template>

<script setup>
import { ArrowLeftBold } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
let router = useRouter();

let back = () => {
    router.go(-1)
}
</script>

<style scoped>
.back-top {
    display: flex;
    align-items: center;
    cursor: pointer;
}
.back-top > span {
    margin-left: 4px;
    font-size: 16px;
}
</style>

