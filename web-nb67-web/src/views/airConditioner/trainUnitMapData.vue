<template>
    <div class="map" id="unitMap" ref="unitMap">
        <!-- 冷凝风机电流 -->
        <div class="condensateFan1 orientation" :class="{ 'is-running': checkRun('cf', 2) }">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.i_cf_u12, 'A') : formatVal(props.crewDetails.i_cf_u22, 'A')  }}</span>A</div>
            </div>
            <div class="popup cardAlign">冷凝风机2运行中</div>
        </div>
        <div class="condensateFan2 orientation" :class="{ 'is-running': checkRun('cf', 1) }">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.i_cf_u11, 'A') : formatVal(props.crewDetails.i_cf_u21, 'A')  }}</span>A</div>
            </div>
            <div class="popup cardAlign">冷凝风机1运行中</div>
        </div>

        <!-- 压缩机 -->
        <div class="compressor2 orientation" :class="{ 'is-running': checkRun('cp', 2) }">
            <div class="cardItem">
                <div class="unitFont compressor2A">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.i_cp_u12, 'A') : formatVal(props.crewDetails.i_cp_u22, 'A')  }}</span>A
                    <div class="popup cardAlign">压缩机2已开启</div>
                </div>
            </div>
            <div class="cardItem">
                <div class="unitFont compressor2V">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.v_cp_u12, 'V') : formatVal(props.crewDetails.v_cp_u22, 'V')  }}</span>V
                    <div class="popup cardAlign">压缩机2电压</div>
                </div>
            </div>
            <div class="cardItem">
                <div class="unitFont compressor2Hz">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.f_cp_u12, 'Hz') : formatVal(props.crewDetails.f_cp_u22, 'Hz')  }}</span>Hz
                    <div class="popup cardAlign">压缩机2频率</div>
                </div>
            </div>
        </div>
        <div class="compressor1 orientation" :class="{ 'is-running': checkRun('cp', 1) }">
            <div class="cardItem">
                <div class="unitFont compressor1A">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.i_cp_u11, 'A') : formatVal(props.crewDetails.i_cp_u21, 'A') }}</span>A
                    <div class="popup cardAlign">压缩机1已开启</div>
                </div>
            </div>
            <div class="cardItem">
                <div class="unitFont compressor1V">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.v_cp_u11, 'V') : formatVal(props.crewDetails.v_cp_u21, 'V') }}</span>V
                    <div class="popup cardAlign">压缩机1电压</div>
                </div>
            </div>
            <div class="cardItem">
                <div class="unitFont compressor1Hz">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.f_cp_u11, 'Hz') : formatVal(props.crewDetails.f_cp_u21, 'Hz') }}</span>Hz
                    <div class="popup cardAlign">压缩机1频率</div>
                </div>
            </div>
        </div>

        <!-- 系统高压与吸气 -->
        <div class="system2HighPressure orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.highpress_u12, 'bar') : formatVal(props.crewDetails.highpress_u22, 'bar') }}</span>bar</div>
            </div>
            <div class="popup cardAlign">系统2高压</div>
        </div>
        <div class="system2inhale orientation">
            <div class="cardItem largeMargin system2inhaleTemperature">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.suckt_u12, '℃') : formatVal(props.crewDetails.suckt_u22, '℃')}}</span>℃</div>
                <div class="popup cardAlign">系统2吸气温度</div>
            </div>
            <div class="cardItem system2inhaleBar">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.suckp_u12, 'bar') : formatVal(props.crewDetails.suckp_u22, 'bar') }}</span>bar</div>
                <div class="popup cardAlign">系统2吸气压力</div>
            </div>
        </div>

        <div class="system1HighPressure orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.highpress_u11, 'bar'): formatVal(props.crewDetails.highpress_u21, 'bar') }}</span>bar</div>
            </div>
            <div class="popup cardAlign">系统1高压</div>
        </div>
        <div class="system1inhale orientation">
            <div class="cardItem largeMargin system1inhaleTemperature">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.suckt_u11, '℃'): formatVal(props.crewDetails.suckt_u21, '℃') }}</span>℃</div>
                <div class="popup cardAlign">系统1吸气温度</div>
            </div>
            <div class="cardItem system1inhaleBar">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.suckp_u11, 'bar') : formatVal(props.crewDetails.suckp_u21, 'bar') }}</span>bar</div>
                <div class="popup cardAlign">系统1吸气压力</div>
            </div>
        </div>

        <!-- 温度参数 -->
        <div class="supplyAirTemperature2 orientation">
            <div class="cardItem">
                <div class="unitFont"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.sas_u12, '℃'): formatVal(props.crewDetails.sas_u22, '℃')}}</span>℃</div>
            </div>
            <div class="popup cardAlign">送风温度2</div>
        </div>
        <div class="fanE2 orientation" :class="{ 'is-running': checkRun('ef', 2) }">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.i_ef_u12, 'A'): formatVal(props.crewDetails.i_ef_u22, 'A') }}</span>A</div>
            </div>
            <div class="popup cardAlign">通风机2运行中</div>
        </div>
        <div class="supplyAirTemperature1 orientation">
            <div class="cardItem">
                <div class="unitFont"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.sas_u11, '℃'): formatVal(props.crewDetails.sas_u21, '℃')}}</span>℃</div>
            </div>
            <div class="popup cardAlign">送风温度1</div>
        </div>
        <div class="fanE1 orientation" :class="{ 'is-running': checkRun('ef', 1) }">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.i_ef_u11, 'A'): formatVal(props.crewDetails.i_ef_u21, 'A')}}</span>A</div>
            </div>
            <div class="popup cardAlign">通风机1运行中</div>
        </div>

        <!-- 膨胀阀 -->
        <div class="system2ExpansionValve orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.eevpos_u12, '步') : formatVal(props.crewDetails.eevpos_u22, '步') }}</span>步</div>
            </div>
            <div class="popup cardAlign">系统2膨胀阀开度</div>
        </div>
        <div class="system1ExpansionValve orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.eevpos_u11, '步') : formatVal(props.crewDetails.eevpos_u21, '步') }}</span>步</div>
            </div>
            <div class="popup cardAlign">系统1膨胀阀开度</div>
        </div>

        <!-- 空气监测网络 -->
        <div class="airMonitoring orientation">
            <div class="airMonitoringCard">
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardT">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.aq_t_u1, '℃'): formatVal(props.crewDetails.aq_t_u2, '℃') }}</span>℃
                        <div class="popup cardAlign">空气监测终端温度</div>
                    </div>
                </div>
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardRH">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.aq_h_u1, 'RH') : formatVal(props.crewDetails.aq_h_u2, 'RH') }}</span>RH
                        <div class="popup cardAlign">空气监测终端湿度</div>
                    </div>
                </div>
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardppm">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.aq_co2_u1, 'ppm') : formatVal(props.crewDetails.aq_co2_u2, 'ppm') }}</span>ppm
                        <div class="popup cardAlign">空气监测终端CO2</div>
                    </div>
                </div>
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardppb">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.aq_tvoc_u1, 'ppb') : formatVal(props.crewDetails.aq_tvoc_u2, 'ppb') }}</span>ppb
                        <div class="popup cardAlign">空气监测终端TVOC</div>
                    </div>
                </div>
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardPM2">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.aq_pm2_5_u1, 'μg/m³') : formatVal(props.crewDetails.aq_pm2_5_u2, 'μg/m³') }}</span>μg/m³
                        <div class="popup cardAlign">空气监测终端PM2.5</div>
                    </div>
                </div>
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardPM10">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.aq_pm10_u1, 'μg/m³') : formatVal(props.crewDetails.aq_pm10_u2, 'μg/m³') }}</span>μg/m³
                        <div class="popup cardAlign">空气监测终端PM10</div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 阀门与滤网 -->
        <div class="newTrendTemperature orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.fas_u1, '℃'): formatVal(props.crewDetails.fas_u2, '℃')}}</span>℃</div>
            </div>
            <div class="popup cardAlign">新风温度</div>
        </div>
        <div class="newTrendOpen orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.fadpos_u1, '%') : formatVal(props.crewDetails.fadpos_u2, '%') }}</span>%</div>
            </div>
            <div class="popup cardAlign">新风阀开度</div>
        </div>
        <div class="returnAirTemperature orientation">
            <div class="cardItem">
                <div class="unitFont"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.ras_u1, '℃'): formatVal(props.crewDetails.ras_u2, '℃')}}</span>℃</div>
            </div>
            <div class="popup cardAlign">回风温度</div>
        </div>
        <div class="returnAirOpne orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.radpos_u1, '%') : formatVal(props.crewDetails.radpos_u2, '%') }}</span>%</div>
            </div>
            <div class="popup cardAlign">回风阀开度</div>
        </div>
        <div class="percolator orientation">
            <div class="cardItem">
                <div class="unitFont"><span class="dataFont">{{ props.UnitNo =='1号机组' ? formatVal(props.crewDetails.presdiff_u1, 'par'): formatVal(props.crewDetails.presdiff_u2, 'par')}}</span>par</div>
            </div>
            <div class="popup cardAlign">滤网前后压差值</div>
        </div>

        <div class="bg"><img :src="mapBg" /></div>
    </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
    crewDetails: {
        type: Object,
        default: () => ({})
    },
    UnitNo: {
        type: String,  
    }
})

const mapBg = computed(() => {
    return Object.keys(props.crewDetails).length > 0 
        ? '/img/TrainUnitMapBg_anim.svg' 
        : '/img/TrainUnitMapBg.svg'
})


const checkRun = (type, subIdx) => {
    if (!props.crewDetails) return false
    // 自动判定当前是1号还是2号机组
    const unitPrefix = props.UnitNo.includes('1') ? '1' : '2'
    const key = `cfbk_${type}_u${unitPrefix}${subIdx}`
    const val = props.crewDetails[key]
    // 兼容多种运行状态判定逻辑 (1, true, 或字符串 "true")
    return val == 1 || val === true || String(val).toLowerCase() === 'true'
}

const formatVal = (val, unit) => {
    if (val === undefined || val === null || val === '') return '-'
    
    // 强制转换为数字类型，容忍后端传回字符串类型的数字 (例如 "219.00")
    const num = Number(val)
    if (isNaN(num)) return val // 如果根本不是数字，直接原样返回
    
    // 过滤传感器无数据的溢出值
    if (num === 32767 || num === 3276.7 || num === 65535) return '-'
    
    // 不缩放的白名单
    const noScaleUnits = ['步', '%', 'ppm', 'ppb', 'μg/m³', 'RH', 'par'];
    if (noScaleUnits.includes(unit)) {
        return num;
    }
    
    // 执行缩放
    return (num / 10.0).toFixed(1);
}
</script>

<style lang="scss" scoped>
    .map {
        position: relative;
        max-width: 100%;
        min-height: 250px;
        padding-bottom: 30px;
        .bg {
            display: flex;
            align-items: center;
            justify-content: center;
            img {
                width: 90%;
            }
        }
        .cardItem {
            display: flex;
            flex-direction: column;
            font-size: 13px; // 固定像素
            margin-bottom: 20px;
            transition: all .2s ease-in-out 0s;
            &:last-child {
                margin-bottom: 0;
            }
            .dataFont {
                font-size: 14px; // 固定像素
                font-weight: bold;
                margin-right: 5px;
                color: white;
            }
            .unitFont {
                font-weight: bold;
                color: #2186cf;
                white-space: nowrap; // 保证单位和数值不换行
            }
            .fontWidth {
                min-width: 50px;
                display: flex;
                flex-direction: row; // 保证并排
                align-items: baseline;
                justify-content: center;
            }
        }
        .cardAlign {
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .orientation {
            position: absolute;
            z-index: 10;
        }
        .largeMargin {
            margin-bottom: 35px;
        }
        
        // 科技感 Hover (用户要求: 透明地，灰色字体，蓝色框)
        .popup {
            display: none;
            position: absolute;
            top: 0;
            left: 110%;
            background: rgba(10, 15, 29, 0.9);
            backdrop-filter: blur(4px);
            border: 1px solid #2186cf;
            color: #ffffff;
            height: 24px;
            font-size: 11px;
            padding: 0 8px;
            white-space: nowrap;
            border-radius: 4px;
            z-index: 999;
            box-shadow: 0 0 10px rgba(33, 134, 207, 0.4);
            align-items: center;
        }

        .is-running {
            .dataFont {
                color: #00fff2 !important;
                text-shadow: 0 0 10px rgba(0, 255, 242, 0.6);
            }
            .unitFont {
                color: #ffffff !important;
            }
        }
        
        // 坐标完全拷贝老项目百分比
        // 冷凝风机1电流
        .condensateFan1 {
            top: 48.5%; // 原47%
            left: 8%;
        }
        .condensateFan1:hover .popup {
            display: flex;
        }
        // 冷凝风机2电流
        .condensateFan2 {
            top: 48.5%; // 原47%
            left: 16.8%;
        }
        .condensateFan2:hover .popup {
            display: flex;
        }
        // 压缩机2
        .compressor2 {
            top: 32.5%;
            left: 34.5%;
            .compressor2A:hover .popup {
                display: flex;
            }
            .compressor2V:hover .popup {
                display: flex;
                top: 42%;
            }
            .compressor2Hz:hover .popup {
                display: flex;
                top: 84%;
            }
        }
        // 压缩机1
        .compressor1 {
            top: 70%;
            left: 34.5%;
            z-index: 2;
            .compressor1A:hover .popup {
                display: flex;
            }
            .compressor1V:hover .popup {
                display: flex;
                top: 42%;
            }
            .compressor1Hz:hover .popup {
                display: flex;
                top: 84%;
            }
        }
        // 系统2高压
        .system2HighPressure {
            top: 5%; // 原3.5%
            left: 43.5%;
        }
        .system2HighPressure:hover .popup {
            display: flex;
        }
        // 系统2吸气
        .system2inhale {
            top: 24.2%; // 原22.7%
            left: 47%;
            .system2inhaleTemperature:hover .popup {
                display: flex;
            }
            .system2inhaleBar:hover .popup {
                display: flex;
                top: 74%;
            }
        }
        // 系统1高压
        .system1HighPressure {
            top: 50.5%; // 原49%
            left: 46.5%;
        }
        .system1HighPressure:hover .popup {
            display: flex;
        }
        // 系统1吸气
        .system1inhale {
            top: 71.5%; // 原70%
            left: 46%;
            .system1inhaleTemperature:hover .popup {
                display: flex;
            }
            .system1inhaleBar:hover .popup {
                display: flex;
                top: 74%;
                left: 120%;
            }
        }
        // 送风温度2
        .supplyAirTemperature2 {
            top: 55.8%;
            right: 39%;
        }
        .supplyAirTemperature2:hover .popup {
            display: flex;
        }
        // 通风机电流2
        .fanE2 {
            top: 63.5%;
            left: 65.5%;
        }
        .fanE2:hover .popup {
            display: flex;
            left: 110%;
        }
        // 送风温度1
        .supplyAirTemperature1 {
            top: 55.8%;
            left: 95.5%;
        }
        .supplyAirTemperature1:hover .popup {
            display: flex;
            left: -110%; // 修改弹出方向，原为 -90% 可能会重叠
            top: 70%;
        }
        // 通风机电流1
        .fanE1 {
            top: 63.5%;
            right: 8%;
        }
        .fanE1:hover .popup {
            display: flex;
        }
        // 系统2膨胀阀开度
        .system2ExpansionValve {
            top: 1.6%;
            left: 77%;
        }
        .system2ExpansionValve:hover .popup {
            display: flex;
        }
        // 系统1膨胀阀开度
        .system1ExpansionValve {
            top: 96%;
            left: 77%;
        }
        .system1ExpansionValve:hover .popup {
            display: flex;
        }
        // 空气监测
        .airMonitoring {
            top: 15%; // 原13.5%
            left: 80%;
            .airMonitoringCard {
                display: grid;
                grid-template-columns: 80px 80px 80px;
                gap: 5px;
                .cardItem {
                    margin-bottom: 5px;
                }
                // 解决靠右边缘被裁减的问题：让悬浮窗朝左侧弹出
                .popup {
                    left: auto;
                    right: 105%;
                }
                .airMonitoringCardT:hover .popup {
                    display: flex;
                }
                .airMonitoringCardRH:hover .popup {
                    display: flex;
                }
                .airMonitoringCardppm:hover .popup {
                    display: flex;
                }
                .airMonitoringCardppb:hover .popup {
                    display: flex;
                }
                .airMonitoringCardPM2:hover .popup {
                    display: flex;
                }
                .airMonitoringCardPM10:hover .popup {
                    display: flex;
                }
            }
        }
        // 新风温度
        .newTrendTemperature {
            top: 33.5%;
            right: 28.5%;
        }
        .newTrendTemperature:hover .popup {
            display: flex;
        }
        // 新风阀开度
        .newTrendOpen {
            top: 24.5%; // 原23%
            left: 69%;
        }
        .newTrendOpen:hover .popup {
            display: flex;
        }
        // 回风温度
        .returnAirTemperature {
            top: 33.5%;
            left: 83%;
        }
        .returnAirTemperature:hover .popup {
            display: flex;
        }
        // 回风阀开度
        .returnAirOpne {
            top: 24.5%; // 原23%
            left: 85.5%;
        }
        .returnAirOpne:hover .popup {
            display: flex;
        }
        // 滤网
        .percolator {
            top: 69%;
            left: 86%;
        }
        .percolator:hover .popup {
            display: flex;
        }
    }
</style>