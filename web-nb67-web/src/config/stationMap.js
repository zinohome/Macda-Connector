// 宁波地铁站点编码 → 站名映射
// key: line_id（来自 payload_json.line_id）
// value: { stationId: '站名' }
// stationId 对应协议字段 StartStation / TerminalStation / CurStation / NextStation（uint16）

export const STATION_MAP = {
  // 6 号线站点
  6: {
    1:  '高塘桥',
    2:  '李兴',
    3:  '蒋家桥',
    4:  '联丰',
    5:  '集仕港',
    6:  '大通',
    7:  '长丰',
    8:  '金家漕',
    9:  '后庄桥',
    10: '上庄',
    11: '东岙',
    12: '西岙',
    13: '梅墟',
    14: '世纪大道',
    15: '云龙',
    16: '南塘',
    17: '春晓',
    18: '咸祥',
    19: '江六',
  },
  // 7 号线站点
  7: {
    1:  '北仑中心',
    2:  '霞浦',
    3:  '港东路',
    4:  '明港',
    5:  '大碶',
    6:  '凤洋',
    7:  '柴桥',
    8:  '贵驷',
    9:  '镇海',
    10: '九龙湖',
    11: '庄市',
    12: '孔浦',
    13: '庄桥',
    14: '姜山',
    15: '横溪',
    16: '云龙',
  },
}

/**
 * 根据线路号和站点编码返回站名；无映射时原样返回编码
 * @param {number|string} lineId
 * @param {number|string} stationId
 */
export function getStationName(lineId, stationId) {
  if (stationId === '' || stationId === null || stationId === undefined) return '-'
  const lineMap = STATION_MAP[Number(lineId)]
  if (!lineMap) return String(stationId)
  return lineMap[Number(stationId)] || String(stationId)
}
