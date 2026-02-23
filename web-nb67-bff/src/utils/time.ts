import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc.js'
import timezone from 'dayjs/plugin/timezone.js'

dayjs.extend(utc)
dayjs.extend(timezone)

// 统一设置为上海时区
dayjs.tz.setDefault('Asia/Shanghai')

/**
 * 统一时间格式化工具
 * @param date 原始时间数据
 * @param format 目标格式，默认为 YYYY-MM-DD HH:mm:ss
 * @returns 格式化后的字符串
 */
export const formatTime = (date: any, format = 'YYYY-MM-DD HH:mm:ss') => {
    if (!date) return '-'
    // 强制转换为上海时区后再格式化
    return dayjs.tz(date).format(format)
}

export { dayjs }
export default dayjs
