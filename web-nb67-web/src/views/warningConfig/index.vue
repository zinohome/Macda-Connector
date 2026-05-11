<template>
    <div class="warp2">
        <div class="header-monitor">
            <div class="header-left">
                <el-button type="primary" class="nav-btn" icon="ArrowLeft" @click="goBack">返回</el-button>
            </div>
        </div>

        <div class="monitor-container">
            <div class="section-box">
                <div class="section-header">
                    <span class="title-icon"></span>
                    <span class="title-text">预警条件设置</span>
                    <span class="hint">修改阈值后重启 event-builder 容器生效</span>
                </div>
                <el-table :data="tableData" border stripe style="width:100%" v-loading="loading">
                    <el-table-column prop="component_name" label="组件名称" min-width="220" show-overflow-tooltip />
                    <el-table-column prop="threshold_good"   label="良好" width="120" align="center">
                        <template #default="s"><span class="level-good">{{ s.row.threshold_good || '-' }}</span></template>
                    </el-table-column>
                    <el-table-column prop="threshold_normal" label="一般" width="120" align="center">
                        <template #default="s"><span class="level-normal">{{ s.row.threshold_normal || '-' }}</span></template>
                    </el-table-column>
                    <el-table-column prop="threshold_bad"    label="差" width="120" align="center">
                        <template #default="s"><span class="level-bad">{{ s.row.threshold_bad || '-' }}</span></template>
                    </el-table-column>
                    <el-table-column prop="unit"             label="单位" width="90" align="center" />
                    <el-table-column prop="duration_seconds" label="持续时间" width="100" align="center">
                        <template #default="s">{{ formatDuration(s.row.duration_seconds) }}</template>
                    </el-table-column>
                    <el-table-column label="操作" width="80" align="center" fixed="right">
                        <template #default="s">
                            <el-link type="primary" @click="openEdit(s.row)">编辑</el-link>
                        </template>
                    </el-table-column>
                </el-table>
            </div>
        </div>

        <!-- 编辑弹窗 -->
        <el-dialog v-model="editVisible" title="编辑预警条件" width="500px" :append-to-body="true" class="warning-config-dialog">
            <el-form :model="editForm" label-width="100px" size="default">
                <el-form-item label="组件名称">
                    <span class="form-readonly">{{ editForm.component_name }}</span>
                </el-form-item>
                <el-form-item label="良好范围">
                    <el-input v-model="editForm.threshold_good" placeholder="如 0~75" />
                </el-form-item>
                <el-form-item label="一般范围">
                    <el-input v-model="editForm.threshold_normal" placeholder="如 75~150" />
                </el-form-item>
                <el-form-item label="差阈值">
                    <el-input v-model="editForm.threshold_bad" placeholder="如 ≥150" />
                </el-form-item>
                <el-form-item label="触发算符">
                    <el-select v-model="editForm.trigger_operator" style="width:100px">
                        <el-option v-for="op in ['>', '>=', '<', '<=']" :key="op" :label="op" :value="op" />
                    </el-select>
                </el-form-item>
                <el-form-item label="触发阈值">
                    <el-input-number v-model="editForm.trigger_value" :precision="3" :step="0.1" />
                    <span style="margin-left:8px;color:#676e82">{{ editForm.unit }}</span>
                </el-form-item>
                <el-form-item label="消除阈值">
                    <el-input-number v-model="editForm.clear_value" :precision="3" :step="0.1" />
                </el-form-item>
                <el-form-item label="持续时间(秒)">
                    <el-input-number v-model="editForm.duration_seconds" :min="0" :step="60" />
                </el-form-item>
                <el-form-item label="预警策略">
                    <el-input v-model="editForm.strategy" type="textarea" :rows="3" placeholder="指导意见文字" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="editVisible = false">取消</el-button>
                <el-button type="primary" :loading="saving" @click="saveEdit">确认保存</el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getWarningConfigs, updateWarningConfig } from '@/api/api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()

const loading   = ref(false)
const tableData = ref([])

const editVisible = ref(false)
const saving = ref(false)
const editForm = reactive({
    id: null, component_name: '', threshold_good: '', threshold_normal: '', threshold_bad: '',
    trigger_operator: '>', trigger_value: 0, clear_value: 0,
    duration_seconds: 0, unit: '', strategy: ''
})

const formatDuration = (sec) => {
    if (!sec) return '立即'
    if (sec < 60) return `${sec}秒`
    if (sec < 3600) return `${Math.floor(sec/60)}分钟`
    return `${Math.floor(sec/3600)}小时`
}

const fetchData = async () => {
    loading.value = true
    try {
        const res = await getWarningConfigs()
        if (res?.code === 200) tableData.value = res.data || []
    } catch (e) {
        ElMessage.error('加载配置失败')
    } finally {
        loading.value = false
    }
}

const openEdit = (row) => {
    Object.assign(editForm, {
        id: row.id,
        component_name: row.component_name,
        threshold_good: row.threshold_good || '',
        threshold_normal: row.threshold_normal || '',
        threshold_bad: row.threshold_bad || '',
        trigger_operator: row.trigger_operator || '>',
        trigger_value: Number(row.trigger_value) || 0,
        clear_value: Number(row.clear_value) || 0,
        duration_seconds: row.duration_seconds || 0,
        unit: row.unit || '',
        strategy: row.strategy || ''
    })
    editVisible.value = true
}

const saveEdit = async () => {
    saving.value = true
    try {
        const res = await updateWarningConfig(editForm.id, {
            threshold_good: editForm.threshold_good,
            threshold_normal: editForm.threshold_normal,
            threshold_bad: editForm.threshold_bad,
            trigger_operator: editForm.trigger_operator,
            trigger_value: editForm.trigger_value,
            clear_value: editForm.clear_value,
            duration_seconds: editForm.duration_seconds,
            strategy: editForm.strategy
        })
        if (res?.code === 200) {
            ElMessage.success('保存成功')
            editVisible.value = false
            fetchData()
        } else {
            ElMessage.error(res?.message || '保存失败')
        }
    } catch (e) {
        ElMessage.error('保存失败')
    } finally {
        saving.value = false
    }
}

const goBack = () => router.push({
    name: 'historyWarning',
    query: { trainNo: route.query.trainNo, trainCoach: route.query.trainCoach }
})
onMounted(() => fetchData())
</script>

<style scoped lang="scss">
.warp2 { background:#0a0f1d; min-height:100vh; padding:0; color:#fff; }
.header-monitor {
    height:50px; display:flex; align-items:center;
    padding:0 20px;
    background: linear-gradient(180deg, #181f30 0%, #0a0f1d 100%);
    position:sticky; top:0; z-index:1000;
    border-bottom:1px solid rgba(33,134,207,0.2);
    .header-left { display:flex; align-items:center; gap:8px; }
}
.monitor-container { padding:15px; }
.section-box { background:#141a2e; border-radius:8px; border:1px solid #262e45; padding:15px; }
.section-header { display:flex; align-items:center; gap:8px; margin-bottom:15px;
    .title-icon { width:4px; height:16px; background:#2186cf; border-radius:2px; }
    .title-text { color:#fff; font-size:15px; font-weight:bold; }
    .hint { font-size:11px; color:#676e82; margin-left:8px; }
}
.nav-btn {
    background:#0a0f1d!important; border:1px solid #2186cf!important; color:#fff!important;
    font-size:13px!important; height:32px!important; padding:0 12px!important; border-radius:4px!important;
    &:hover { background:rgba(33,134,207,0.2)!important; }
}
.level-good   { color:#42ad5d; font-size:12px; }
.level-normal { color:#ffa55c; font-size:12px; }
.level-bad    { color:#e65355; font-size:12px; }
.form-readonly { color:#d1d9e7; font-size:13px; }
:deep(.el-table) {
    background:transparent!important; color:#d1d9e7;
    --el-table-border-color:#262e45; --el-table-header-bg-color:#1a2234;
    --el-table-tr-bg-color:transparent; --el-table-row-hover-bg-color:rgba(33,134,207,0.1);
    th.el-table__cell { background:#1a2234!important; color:#2186cf; font-weight:bold; }
}
</style>

<style lang="scss">
/* 编辑预警条件弹窗 — 全局非 scoped，覆盖 teleport 到 body 的 el-dialog */
.warning-config-dialog.el-dialog {
    background: #141a2e !important;
    border: 1px solid #262e45 !important;
    border-radius: 8px;

    .el-dialog__header {
        border-bottom: 1px solid #262e45;
        padding: 12px 20px;
        .el-dialog__title {
            color: #ffffff;
            font-size: 14px;
            font-weight: bold;
        }
        .el-dialog__headerbtn .el-dialog__close { color: #676e82; &:hover { color: #fff; } }
    }

    .el-dialog__body { padding: 20px; background: #141a2e; }
    .el-dialog__footer { border-top: 1px solid #262e45; padding: 12px 20px; background: #141a2e; }

    /* 表单标签 */
    .el-form-item__label { color: #d1d9e7 !important; font-size: 13px; }

    /* 文本/数字输入框 */
    .el-input__wrapper {
        background: #0a0f1d !important;
        box-shadow: 0 0 0 1px #394153 inset !important;
        &:hover { box-shadow: 0 0 0 1px #2186cf inset !important; }
        &.is-focus { box-shadow: 0 0 0 1px #2186cf inset !important; }
    }
    .el-input__inner { color: #ffffff !important; background: transparent !important; font-size: 13px; }

    /* textarea */
    .el-textarea__inner {
        color: #ffffff !important; background: #0a0f1d !important; font-size: 13px;
        box-shadow: 0 0 0 1px #394153 inset !important;
        &:focus { box-shadow: 0 0 0 1px #2186cf inset !important; }
    }

    /* 数字输入框加减按钮 */
    .el-input-number__decrease, .el-input-number__increase {
        background: #1a2234 !important; border-color: #394153 !important; color: #d1d9e7 !important;
        &:hover { color: #2186cf !important; }
    }

    /* 下拉选择器 */
    .el-select .el-input__wrapper { background: #0a0f1d !important; }

    /* 只读文本 */
    .form-readonly { color: #d1d9e7; font-size: 13px; }

    /* 按钮 */
    .el-button--primary { background: #2186cf !important; border-color: #2186cf !important; color: #fff !important; }
    .el-button:not(.el-button--primary) {
        background: #1a2234 !important; border-color: #394153 !important; color: #d1d9e7 !important;
        &:hover { border-color: #2186cf !important; color: #fff !important; }
    }
}
</style>
