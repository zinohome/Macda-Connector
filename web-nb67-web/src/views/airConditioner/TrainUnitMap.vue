<template>
    <div class="train-unit-map-container">
        <div class="section-header">
            <span class="title-icon"></span>
            <span class="title-text">空调机组详情</span>
        </div>
        <div class="train-map-card">
            <!-- 内部悬浮切换按钮 (参考老图红圈位置) -->
            <div class="unit-toggle-overlay">
                <div 
                    class="toggle-item" 
                    :class="{ active: selectedUnit === 1 }" 
                    @click="selectedUnit = 1"
                >机组一</div>
                <div 
                    class="toggle-item" 
                    :class="{ active: selectedUnit === 2 }" 
                    @click="selectedUnit = 2"
                >机组二</div>
            </div>

            <div class="map-wrapper">
                <trainUnitMapData 
                    :crewDetails="crewDetails" 
                    :UnitNo="`${selectedUnit}号机组`"
                />
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, watch, onUnmounted } from 'vue'
import { gettemperature } from '@/api/api.js'
import trainUnitMapData from "./trainUnitMapData.vue";

const props = defineProps({
    carriageId: { type: String, default: '' }
})

const selectedUnit = ref(1)
const crewDetails = ref({})
let refreshTimer = null

const fetchData = () => {
    if (!props.carriageId) return
    
    // 1. 先重置数据，防止无数据时残留上一节车厢的旧数据
    // crewDetails.value = {} // 自动刷新时不需要重置，避免闪烁

    gettemperature(props.carriageId).then(res => {
        if (res && res.vw_system_info && res.vw_system_info.length > 0) {
            const raw = res.vw_system_info[0]
            // 2. 只有在有数据时才建立 Proxy 映射
            crewDetails.value = new Proxy(raw, {
                get(target, prop) {
                    const key = String(prop)
                    if (key in target) return target[key]
                    const search = key.toLowerCase().replace(/_/g, '')
                    const foundKey = Object.keys(target).find(k => 
                        k.toLowerCase().replace(/_/g, '') === search
                    )
                    return foundKey ? target[foundKey] : undefined
                }
            })
        }
    })
}

watch(() => props.carriageId, () => {
    selectedUnit.value = 1
    crewDetails.value = {} // 车厢切换时重置
    fetchData()
})

import { MONITOR_CONFIG } from '@/config/monitorConfig.js'

onMounted(() => {
    fetchData()
    refreshTimer = setInterval(fetchData, MONITOR_CONFIG.refreshInterval)
    console.log('[TrainUnitMap] Timer Started')
})

onUnmounted(() => {
    if (refreshTimer) {
        clearInterval(refreshTimer)
        console.log('[TrainUnitMap] Timer Cleared')
    }
})
</script>

<style lang="scss" scoped>
.train-unit-map-container {
    padding: 20px;
}

.section-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 20px;

    .title-icon {
        width: 4px;
        height: 16px;
        background: #ffffff;
        border-radius: 2px;
    }

    .title-text {
        color: #ffffff;
        font-size: 16px;
        font-weight: bold;
    }
}

.train-map-card {
    position: relative;
    width: 100%;
    background: #181f30;
    border-radius: 8px;
    border: 1px solid #394153;
    overflow: hidden;
}

.unit-toggle-overlay {
    position: absolute;
    top: 20px;
    left: 20px;
    z-index: 100;
    display: flex;
    flex-direction: column;
    gap: 10px;

    .toggle-item {
        color: #676e82;
        font-size: 14px;
        cursor: pointer;
        padding: 4px 10px;
        border-radius: 4px;
        transition: all 0.3s;
        border: 1px solid transparent;

        &:hover { color: #fff; }
        &.active {
            color: #2186cf;
            font-weight: bold;
            border: 1px solid rgba(33, 134, 207, 0.4);
            background: rgba(33, 134, 207, 0.1);
        }
    }
}

.map-wrapper {
    width: 100%;
    height: auto;
}
</style>
