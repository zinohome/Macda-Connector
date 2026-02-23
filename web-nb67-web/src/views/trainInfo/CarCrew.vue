<template>
    <div class="train-overview-card">
        <div class="section-title">
        </div>
        <div class="train-visual-container">
            <div 
                v-for="(car, index) in trainSelectionData" 
                :key="car.carriage_no" 
                :class="['car-item', { 'is-active': car.carriage_no === props.activeCarriage }]"
                @click="handleSelect(car.carriage_no)"
            >
                <div class="car-image-box">
                    <img :src="getCarImage(index)" class="car-img" />
                </div>
                
                <div class="car-info">
                    <el-icon v-if="!car.hasData" color="#676e82" class="status-mini-icon"><RemoveFilled /></el-icon>
                    <el-icon v-else-if="car.alarm_count > 0" color="#e65355" class="status-mini-icon"><CircleCloseFilled /></el-icon>
                    <el-icon v-else-if="car.warning_count > 0" color="#ffa55c" class="status-mini-icon"><WarnTriangleFilled /></el-icon>
                    <el-icon v-else color="#42ad5d" class="status-mini-icon"><SuccessFilled /></el-icon>
                    <span class="car-name">{{ car.name }}</span>
                </div>

                <div v-if="car.carriage_no === props.activeCarriage" class="active-arrow"></div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { getTrainSelection } from '@/api/api'

const props = defineProps({
    trainId: { type: String, default: '' },
    activeCarriage: { type: String, default: '' }
})

const emit = defineEmits(['select'])

const trainSelectionData = ref([
    { name: 'TC1', carriage_no: '1', alarm_count: 0, warning_count: 0, hasData: false },
    { name: 'MP1', carriage_no: '2', alarm_count: 0, warning_count: 0, hasData: false },
    { name: 'M1', carriage_no: '3', alarm_count: 0, warning_count: 0, hasData: false },
    { name: 'M2', carriage_no: '4', alarm_count: 0, warning_count: 0, hasData: false },
    { name: 'MP2', carriage_no: '5', alarm_count: 0, warning_count: 0, hasData: false },
    { name: 'TC2', carriage_no: '6', alarm_count: 0, warning_count: 0, hasData: false }
])

const getCarImage = (index) => {
    if (index === 0) return '/src/assets/img/carHeader.svg'
    if (index === 5) return '/src/assets/img/carEnd.svg'
    return '/src/assets/img/carBody.svg'
}

const handleSelect = (carriageId) => {
    emit('select', carriageId)
}

const fetchData = () => {
    if (!props.trainId) return
    getTrainSelection(props.trainId).then((res) => {
        const statuses = res.vw_carriage_status || []
        const predicts = res.vw_carriage_predict_status || []
        
        trainSelectionData.value.forEach((car, index) => {
            const carIdx = (index + 1).toString()
            const fullCarId = props.trainId + '0' + carIdx
            
            const status = statuses.find(s => s.carriage_no === fullCarId)
            const predict = predicts.find(p => p.carriage_no === fullCarId)
            
            car.carriage_no = carIdx // 保持与下拉框一致的简易 ID
            car.hasData = !!status // 如果在 vw_carriage_status 中找不到记录，说明完全无报文交互
            car.alarm_count = status ? status.alarm_count : 0
            car.warning_count = predict ? predict.warning_count : 0
        })
    })
}

watch(() => props.trainId, fetchData)

onMounted(fetchData)
</script>

<style scoped lang="scss">
.train-overview-card {
    padding: 5px 20px 10px;
}

.section-title {
    display: none;
}

.train-visual-container {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 4px;
    padding: 0;
}

.car-item {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    position: relative;
    padding: 10px 5px;
    border-radius: 8px;
    

    
    &.is-active {
        background: rgba(33, 134, 207, 0.12);
        border: 1px solid #2186cf;
        box-shadow: inset 0 0 10px rgba(33, 134, 207, 0.2);
        
        .car-name { color: #fff; font-weight: bold; }
        .car-img { filter: drop-shadow(0 0 8px #2186cf); }
    }
}

.car-image-box {
    width: 100%;
    margin-bottom: 10px;
    .car-img {
        width: 100%;
        height: auto;
        object-fit: contain;
    }
}

.car-info {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    
    .status-mini-icon {
        font-size: 16px;
    }

    .car-name {
        font-size: 14px;
        color: #2186cf;
    }
}

.active-arrow {
    position: absolute;
    bottom: -6px;
    width: 0;
    height: 0;
    border-left: 6px solid transparent;
    border-right: 6px solid transparent;
    border-bottom: 6px solid #2186cf;
}


</style>
