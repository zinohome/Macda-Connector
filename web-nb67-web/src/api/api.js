import request from '../utils/request.js'

//获取指定列车空调运行状态
export const getAirStartDataApi = (params) => request('/api/airStartData', 'get', params)

//取在线列车
// export const getActiveCarApi = () => request('/api/getActiveCar', 'get')
export const getActiveCarApi = () => request('/api/rest/train', 'get')
// 空调系统
export const getAirSystemApi = () => request('/api/rest/AirSystem', 'get')
// 实时报警
export const getRealtimeAlarm = (params) => request('/api/rest/v2/train/RealtimeAlarm', 'post', params)
// 实时预警
export const getRealtimeWarning = () => request('/api/rest/RealtimeWarning', 'get')
// 空调故障统计
export const getFaultStatistics = (params) => request('/api/rest/v2/FaultStatistics', 'post', params)
// 列车详情-车厢实时报警
export const getRealtimeAlarmDetail = (trainID) => request('/api/rest/train/RealtimeAlarm/' + trainID, 'get')
// 列车详情-车厢状态预警
export const getStatusAlert = (trainID) => request('/api/rest/train/StatusAlert/' + trainID, 'get')
// 列车详情-车厢选择
export const getTrainSelection = (trainId) => request('/api/rest/rain/TrainSelection/' + trainId, 'get')
// 列车详情-报警信息
// export const getAlarmInformation = (trainId) => request('/api/rest/train/AlarmInformation/' + trainId, 'get')
export const getAlarmInformation = (params) => request('/api/rest/train/AlarmInformation', 'post', params)

// export const getAlarmInformation = (params) => request('/v1/graphql', 'post', params)
// 列车详情-所有机组运行状态
export const getRunningState = (trainID) => request('/api/rest/train/RunningState/' + trainID, 'get')
// 列车详情-车厢温度信息
export const getTrainTemperature = (trainID) => request('/api/rest/train/TrainTemperature/' + trainID, 'get')
// 列车详情-车厢详情-机组详情-健康状态
export const getHealthAssessment = (carriageID) => request('/api/rest/carriage/HealthAssessment/' + carriageID, 'get')
// 列车详情-车厢详情-机组详情-系统温度1
export const gettemperature = (carriageID) => request('/api/rest/carriage/SystemInfo/' + carriageID, 'get')
// 列车详情-车厢详情-机组详情-系统温度1
export const getSystemFirst = (carriageID) => request('/api/rest/carriage/SystemInfo/' + carriageID, 'get')
// 列车详情-车厢详情-机组详情-系统温度2
export const getSystemSecond = (carriageID) => request('/api/rest/carriage/SystemSecond/' + carriageID, 'get')
//获取在线列车数据
export const getActiveCarNumApi = () => request('/api/getActiveCarNum', 'get')

//获取在线列车告警数据
export const getCarEaDataApi = () => request('/api/getActiveCarEaData', 'get')

//获取在线列车预警数据
export const getCarPaDataApi = () => request('/api/getActiveCarPaData', 'get')

//获取告警统计数据
export const getEaDataApi = (params) => request('/api/getEaData', 'get', params)

// 获取指定列车的告警或预警数据 
export const getEaDataBydvcAddrApi = (params) => request('/api/getEaDataBydvcAddr', 'get', { restore: true, ...params })

//获取指定列车具体车厢最近10分钟内最新一条数据的健康评估信息数据
export const getHealthDataApi = (id) => request('/api/getHealthData', 'get', { FL_dvcAddr: id })

// 获取指定列车具体车厢最近10分钟内最新一条数据
export const getNewDataApi = (id) => request('/api/getNewData', 'get', { FL_dvcAddr: id })

//获取在线列车指定列车预警数据
export const getPaDataByDvcAddrApi = () => request('/api/getPaDataByDvcAddr', 'get')

// 获取在线列车指定列车实时统计数据
export const getStatisDataByDvcAddrApi = () => request('/api/getStatisDataByDvcAddr', 'get')

// 获取在线列车指定列车实时数据
export const getThDataByDvcAddrApi = (params) => request('/api/rest/carriage/TemperatureTrend', 'post', params)

// 获取指定列车最近10分钟的各节车厢数据
export const getTrainDataApi = (params) => request('/api/getTrainData', 'get', params)

//根据id获取数据
export const getEaOrPaById = (params) => request('/api/getEaOrPaById', 'get', params)