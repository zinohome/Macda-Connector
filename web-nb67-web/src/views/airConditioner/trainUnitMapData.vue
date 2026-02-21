<template>
    <div class="map" id="unitMap" ref="unitMap">
        <div class="condensateFan1 orientation">
            <div class="cardItem ">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.i_cf_u12 : props.crewDetails.i_cf_u22  }}</span>A</div>
            </div>
            <div class="popup cardAlign">冷凝风机2电流</div>
        </div>
        <div class="condensateFan2 orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.i_cf_u11 : props.crewDetails.i_cf_u21  }}</span>A</div>
            </div>
            <div class="popup cardAlign">冷凝风机1电流</div>
        </div>
        <div class="compressor2 orientation">
            <div class="cardItem">
                <div class="unitFont compressor2A">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.i_cp_u12 : props.crewDetails.i_cp_u22  }}</span>A
                    <div class="popup cardAlign">压缩机2电流</div>
                </div>
            </div>
            <div class="cardItem">
                <div class="unitFont compressor2V">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.v_cp_u12  : props.crewDetails.v_cp_u22  }}</span>V
                    <div class="popup cardAlign">压缩机2电压</div>
                </div>
            </div>
            <div class="cardItem">
                <div class="unitFont compressor2Hz">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.f_cp_u12  : props.crewDetails.f_cp_u22  }}</span>Hz
                    <div class="popup cardAlign">压缩机2频率</div>
                </div>
            </div>
        </div>
        <div class="compressor1 orientation">
            <div class="cardItem">
                <div class="unitFont compressor1A">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.i_cp_u11 : props.crewDetails.i_cp_u21 }}</span>A
                    <div class="popup cardAlign">压缩机1电流</div>
                </div>
            </div>
            <div class="cardItem">
                <div class="unitFont compressor1V">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.v_cp_u11 : props.crewDetails.v_cp_u21 }}</span>V
                    <div class="popup cardAlign">压缩机1电压</div>
                </div>
            </div>
            <div class="cardItem">
                <div class="unitFont compressor1Hz">
                    <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.f_cp_u11 : props.crewDetails.f_cp_u21 }}</span>Hz
                    <div class="popup cardAlign">压缩机1频率</div>
                </div>
            </div>
        </div>
        <div class="system2HighPressure orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.highpress_u12 : props.crewDetails.highpress_u22 }}</span>bar</div>
            </div>
            <div class="popup cardAlign">系统2高压</div>
        </div>
        <div class="system2inhale orientation">
            <div class="cardItem largeMargin system2inhaleTemperature">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.suckt_u12 : props.crewDetails.suckt_u22}}</span>℃</div>
                <div class="popup cardAlign">系统2吸气温度</div>
            </div>
            <div class="cardItem system2inhaleBar">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.suckp_u12 : props.crewDetails.suckp_u22 }}</span>bar</div>
                <div class="popup cardAlign">系统2吸气压力</div>
            </div>
        </div>
        <div class="system1HighPressure orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.highpress_u11: props.crewDetails.highpress_u21 }}</span>bar</div>
            </div>
            <div class="popup cardAlign">系统1高压</div>
        </div>
        <div class="system1inhale orientation">
            <div class="cardItem largeMargin system1inhaleTemperature">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.suckt_u11: props.crewDetails.suckt_u21 }}</span>℃</div>
                <div class="popup cardAlign">系统1吸气温度</div>
            </div>
            <div class="cardItem system1inhaleBar">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.suckp_u11 : props.crewDetails.suckp_u21 }}</span>bar</div>
                <div class="popup cardAlign">系统1吸气压力</div>
            </div>
        </div>
        <div class="supplyAirTemperature2 orientation">
            <div class="cardItem">
                <div class="unitFont"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.sas_u12: props.crewDetails.sas_u22}}</span>℃</div>
            </div>
            <div class="popup cardAlign">送风温度2</div>
        </div>
        <div class="fanE2 orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.i_ef_u12: props.crewDetails.i_ef_u22 }}</span>A</div>
            </div>
            <div class="popup cardAlign">通风机2电流</div>
        </div>
        <div class="supplyAirTemperature1 orientation">
            <div class="cardItem">
                <div class="unitFont"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.sas_u11: props.crewDetails.sas_u21}}</span>℃</div>
            </div>
            <div class="popup cardAlign">送风温度1</div>
        </div>
        <div class="fanE1 orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.i_ef_u11: props.crewDetails.i_ef_u21}}</span>A</div>
            </div>
            <div class="popup cardAlign">通风机1电流</div>
        </div>
        <div class="system2ExpansionValve orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.eevpos_u12 : props.crewDetails.eevpos_u22 }}</span>步</div>
            </div>
            <div class="popup cardAlign">系统2膨胀阀开度</div>
        </div>
        <div class="system1ExpansionValve orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.eevpos_u11 : props.crewDetails.eevpos_u21 }}</span>步</div>
            </div>
            <div class="popup cardAlign">系统1膨胀阀开度</div>
        </div>
        <div class="airMonitoring orientation">
            <div class="airMonitoringCard">
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardT">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.aq_t_u1: props.crewDetails.aq_t_u2 }}</span>℃
                        <div class="popup cardAlign">空气监测终端温度</div>
                    </div>
                </div>
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardRH">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.aq_h_u1 : props.crewDetails.aq_h_u2 }}</span>RH
                        <div class="popup cardAlign">空气监测终端湿度</div>
                    </div>
                </div>
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardppm">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.aq_co2_u1 : props.crewDetails.aq_co2_u2 }}</span>ppm
                        <div class="popup cardAlign">空气监测终端CO2</div>
                    </div>
                </div>
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardppb">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.aq_tvoc_u1 : props.crewDetails.aq_tvoc_u2 }}</span>ppb
                        <div class="popup cardAlign">空气监测终端TVOC</div>
                    </div>
                </div>
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardPM2">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.aq_pm2_5_u1 : props.crewDetails.aq_pm2_5_u2 }}</span>μg/m³
                        <div class="popup cardAlign">空气监测终端PM2.5</div>
                    </div>
                </div>
                <div class="cardItem">
                    <div class="unitFont airMonitoringCardPM10">
                        <span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.aq_pm10_u1 : props.crewDetails.aq_pm10_u2 }}</span>μg/m³
                        <div class="popup cardAlign">空气监测终端PM10</div>
                    </div>
                </div>
            </div>
        </div>
        <div class="newTrendTemperature orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.fas_u1: props.crewDetails.fas_u2}}</span>℃</div>
            </div>
            <div class="popup cardAlign">新风温度</div>
        </div>
        <div class="newTrendOpen orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.fadpos_u1 : props.crewDetails.fadpos_u2 }}</span>%</div>
            </div>
            <div class="popup cardAlign">新风阀开度</div>
        </div>
        <div class="returnAirTemperature orientation">
            <div class="cardItem">
                <div class="unitFont"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.ras_u1: props.crewDetails.ras_u2}}</span>℃</div>
            </div>
            <div class="popup cardAlign">回风温度</div>
        </div>
        <div class="returnAirOpne orientation">
            <div class="cardItem">
                <div class="unitFont fontWidth"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.radpos_u1 : props.crewDetails.radpos_u2 }}</span>%</div>
            </div>
            <div class="popup cardAlign">回风阀开度</div>
        </div>
        <div class="percolator orientation">
            <div class="cardItem">
                <div class="unitFont"><span class="dataFont">{{ props.UnitNo =='1号机组' ? props.crewDetails.presdiff_u1: props.crewDetails.presdiff_u2}}</span>par</div>
            </div>
            <div class="popup cardAlign">滤网前后压差值</div>
        </div>
        <div class="bg"><img src="/src/assets/img/TrainUnitMapBg.svg" /></div>
    </div>
</template>
<script setup>
const props = defineProps({
    crewDetails: {
        type: Object,
        default: {}
    },
    UnitNo: {
        type: String,  
    }
})
console.log(props.crewDetails);
</script>
<style lang="scss">
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
            font-size: 0.9vw;
            margin-bottom: 1.5vw;
            transition: all .2s ease-in-out 0s;
            &:last-child {
                margin-bottom: 0;
            }
            .dataTitle {
                margin-bottom: 0.3vw;
            }
            .dataFont {
                font-size: 1.2vw;
                font-weight: bold;
                margin-right: 5px;
                color: white;
            }
            .unitFont {
                font-weight: bold;
                color: #676e82;
            }
            .fontWidth {
                width: 4vw;
                display: flex;
                justify-content: center;
                align-items: flex-end;
            }
        }
        .cardAlign {
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .orientation {
            position: absolute;
        }
        .largeMargin {
            margin-bottom: 2.5vw;
        }
        .popup {
            display: none;
            position: absolute;
            top: 0;
            left: 120%;
            background: rgba(255, 255, 255, 0.9);
            color: #333;
            height: 3vw;
            font-weight: bold;
            font-size: 0.9vw;
            width: 8vw;
            border-radius: 10px;
            border: 2px solid white;
            z-index: 999;
        }
        // 冷凝风机1电流
        .condensateFan1 {
            top: 47%;
            left: 8%;
        }
        .condensateFan1:hover .popup {
            display: flex;
        }
        //  冷凝风机2电流
        .condensateFan2 {
            top: 47%;
            left: 16.8%;
        }
        .condensateFan2:hover .popup {
            display: flex;
        }
        //  压缩机2
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
        //  压缩机1
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
            top: 3.5%;
            left: 43.5%;
        }
        .system2HighPressure:hover .popup {
            display: flex;
            width: 7vw;
            left: 120%;
        }
        // 系统2吸气
        .system2inhale {
            top: 22.7%;
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
            top: 49%;
            left: 46.5%;
        }
        .system1HighPressure:hover .popup {
            display: flex;
            width: 7vw;
            left: 120%;
        }
        // 系统1吸气
        .system1inhale {
            top: 70%;
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
            width: 7vw;
        }
        // 通风机电流2
        .fanE2 {
            top: 63.5%;
            left: 65.5%;
        }
        .fanE2:hover .popup {
            display: flex;
            left: 110%;
            width: 7vw;
        }
        // 送风温度1
        .supplyAirTemperature1 {
            top: 55.8%;
            left: 95.5%;
        }
        .supplyAirTemperature1:hover .popup {
            display: flex;
            width: 7vw;
            left: -90%;
            top: 70%;
        }
        // 通风机电流1
        .fanE1 {
            top: 63.5%;
            right: 8%;
        }
        .fanE1:hover .popup {
            display: flex;
            width: 7vw;
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
            left: 90%;
        }
        // 空气监测
        .airMonitoring {
            top: 13.5%;
            left: 80%;
            .airMonitoringCard {
                display: grid;
                grid-template-columns: 5.5vw 5.5vw 5.5vw;
                .cardItem {
                    margin-bottom: 0.5vw;
                }
                .cardItem:first-child {
                    z-index: 10;
                }
                .cardItem:nth-child(2) {
                    z-index: 9;
                }
                .cardItem:nth-child(3) {
                    z-index: 8;
                }
                .cardItem:nth-child(4) {
                    z-index: 7;
                }
                .cardItem:nth-child(5) {
                    z-index: 6;
                }
                .cardItem:last-child {
                    z-index: 5;
                }
                .airMonitoringCardT:hover .popup {
                    display: flex;
                    left: 20%;
                }
                .airMonitoringCardRH:hover .popup {
                    display: flex;
                    left: 55%;
                }
                .airMonitoringCardppm:hover .popup {
                    display: flex;
                    left: 67%;
                    top: 42%;
                }
                .airMonitoringCardppb:hover .popup {
                    display: flex;
                    left: 25%;
                    top: 58%;
                    width: 9vw;
                }
                .airMonitoringCardPM2:hover .popup {
                    display: flex;
                    left: 65%;
                    top: 58%;
                }
                .airMonitoringCardPM10:hover .popup {
                    display: flex;
                    left: 62%;
                    top: 98%;
                    width: 9vw;
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
            width: 6vw;
        }
        // 新风阀开度
        .newTrendOpen {
            top: 23%;
            left: 69%;
        }
        .newTrendOpen:hover .popup {
            display: flex;
            width: 6vw;
            left: 90%;
        }
        // 回风温度
        .returnAirTemperature {
            top: 33.5%;
            left: 83%;
        }
        .returnAirTemperature:hover .popup {
            display: flex;
            width: 6vw;
        }
        // 回风阀开度
        .returnAirOpne {
            top: 23%;
            left: 85.5%;
        }
        .returnAirOpne:hover .popup {
            display: flex;
            width: 6vw;
            left: 90%;
        }
        // 滤网
        .percolator {
            top: 69%;
            left: 86%;
        }
        .percolator:hover .popup {
            display: flex;
            left: 110%;
        }
        @media screen and (max-width: 1470px)  {
            .cardItem {
                scale: 0.9;
            }
            .popup {
                left: 110%;
                z-index: 999;
            }
            //  压缩机2
            .compressor2 {
                left: 34%;
                .compressor2V:hover .popup {
                    top: 20%;
                }
                .compressor2Hz:hover .popup {
                    top: 15%;
                }
            }
            //  压缩机1
            .compressor1 {
                left: 34%;
                .compressor1V:hover .popup {
                    top: 20%;
                }
                .compressor1Hz:hover .popup {
                    top: 15%;
                }
            }
            // 系统2吸气
            .system2inhale {
                top: 22.4%;
                .system2inhaleBar:hover .popup {
                    top: 15%;
                    left: 122%;
                }
            }
            // 系统1吸气
            .system1inhale {
                top: 69%;
                .system1inhaleBar:hover .popup {
                    top: 15%;
                }
            }
            // 送风温度2
            .supplyAirTemperature2 {
                right: 38.5%;
            }
            // 通风机电流2
            .fanE2 {
                top: 62.5%;
            }
            // 通风机电流1
            .fanE1 {
                top: 62.5%;
            }
            // 新风温度
            .newTrendTemperature {
                top: 33%;
            }
            // 回风温度
            .returnAirTemperature {
                top: 33%;
            }
            // 系统2膨胀阀开度
            .system2ExpansionValve {
                left: 76.5%;
            }
            // 系统1膨胀阀开度
            .system1ExpansionValve {
                top: 95%;
            }
            // 空气监测
            .airMonitoring {
                .airMonitoringCard {
                    > div {
                        margin-bottom: 0.4vw;
                    }
                    .airMonitoringCardT:hover .popup {
                        left: 60%;
                    }
                    .airMonitoringCardRH:hover .popup {
                        left: 70%;
                    }
                    .airMonitoringCardppm:hover .popup {
                        left: 0%;
                        top: 120%;
                    }
                    .airMonitoringCardppb:hover .popup {
                        left: 75%;
                        top: 20%;
                    }
                    .airMonitoringCardPM2:hover .popup {
                        left: 100%;
                        top: 15%;
                    }
                    .airMonitoringCardPM10:hover .popup {
                        left: 0%;
                        top: 110%;
                    }
                }
            }
            // 送风温度1
            .supplyAirTemperature1 {
                top: 55.5%;
                left: 95.5%;
            }
        }
        @media screen and (max-width: 1400px)  {
            .cardItem {
                scale: 0.9;
            }
            .popup {
                left: 110%;
                z-index: 999;
            }
            //  压缩机2
            .compressor2 {
                left: 34%;
                .compressor2V:hover .popup {
                    top: 40%;
                }
                .compressor2Hz:hover .popup {
                    top: 82%;
                }
            }
            //  压缩机1
            .compressor1 {
                left: 34%;
                .compressor1V:hover .popup {
                    top: 40%;
                }
                .compressor1Hz:hover .popup {
                    top: 82%;
                }
            }
            // 系统2吸气
            .system2inhale {
                top: 22.4%;
                .system2inhaleBar:hover .popup {
                    top: 74%;
                    left: 122%;
                }
            }
            // 系统1吸气
            .system1inhale {
                top: 69%;
                .system1inhaleBar:hover .popup {
                    top: 74%;
                }
            }
            // 送风温度2
            .supplyAirTemperature2 {
                right: 38.5%;
            }
            // 通风机电流2
            .fanE2 {
                top: 62.5%;
            }
            // 通风机电流1
            .fanE1 {
                top: 62.5%;
            }
            // 新风温度
            .newTrendTemperature {
                top: 33%;
            }
            // 回风温度
            .returnAirTemperature {
                top: 33%;
            }
            // 系统2膨胀阀开度
            .system2ExpansionValve {
                left: 76.5%;
            }
            // 系统1膨胀阀开度
            .system1ExpansionValve {
                top: 95%;
                left: 76.5%;
            }
            // 空气监测
            .airMonitoring {
                .airMonitoringCard {
                    > div {
                        margin-bottom: 0.4vw;
                    }
                    .airMonitoringCardT:hover .popup {
                        left: 20%;
                    }
                    .airMonitoringCardRH:hover .popup {
                        left: 60%;
                    }
                    .airMonitoringCardppm:hover .popup {
                        left: 50%;
                        top: 50%;
                    }
                    .airMonitoringCardppb:hover .popup {
                        left: 25%;
                        top: 55%;
                    }
                    .airMonitoringCardPM2:hover .popup {
                        left: 65%;
                        top: 55%;
                    }
                    .airMonitoringCardPM10:hover .popup {
                        left: 40%;
                        top: 110%;
                    }
                }
            }
            // 送风温度1
            .supplyAirTemperature1 {
                top: 55.5%;
                left: 95.5%;
            }
        }
    }
</style>