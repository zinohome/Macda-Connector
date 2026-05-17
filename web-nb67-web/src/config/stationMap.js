// 宁波地铁站点编码 → 站名映射
// key: line_id（来自 payload_json.line_id）
// value: { stationId: '站名' }
// stationId 对应协议字段 StartStation / TerminalStation / CurStation / NextStation（uint16）

export const STATION_MAP = {
  // 6 号线站点（来源：6号线站点.xlsx）
  6: {
    1:  '古林',
    2:  '云林西路',
    3:  '集士港',
    4:  '汇士路',
    5:  '卖面桥',
    6:  '宋家漕',
    7:  '望春桥',
    8:  '市中医院',
    9:  '翠柏里',
    10: '新芝路',
    11: '宁波大剧院',
    12: '正大路',
    13: '庆丰桥东',
    14: '明楼',
    15: '明珠路',
    16: '沧海北路',
    17: '院士路',
    18: '聚贤路',
    19: '梅墟',
    20: '谢墅',
    21: '小港西',
    22: '衙前',
    23: '红联',
  },
  // 7 号线站点（来源：7号线站点.xlsx）
  7: {
    1:  '云龙',
    2:  '小洋江',
    3:  '安石路',
    4:  '回龙',
    5:  '尚德路',
    6:  '邱隘',
    7:  '盛莫路',
    8:  '北明程路',
    9:  '民安东路',
    10: '新天路',
    11: '福民公园',
    12: '体育馆',
    13: '曙光',
    14: '外滩大桥',
    15: '宁波大剧院',
    16: '湾头',
    17: '云飞路',
    18: '宁慈东路',
    19: '应嘉路',
    20: '骆兴',
    21: '镇海大道',
    22: '镇海市民广场',
    23: '贵驷',
    24: '明海北路',
    25: '俞范',
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
