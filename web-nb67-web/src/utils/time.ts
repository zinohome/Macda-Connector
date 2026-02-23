import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'

dayjs.extend(utc)
dayjs.extend(timezone)

// 统一设置为上海时区
dayjs.tz.setDefault('Asia/Shanghai')

/**
 * 统一时间格式化工具
 * @param date 原始时间数据 (Date对象, 字符串, 或时间戳)
 * @param format 目标格式，默认为 YYYY-MM-DD HH:mm:ss
 * @returns 格式化后的字符串
 */
export const formatTime = (date: any, format = 'YYYY-MM-DD HH:mm:ss') => {
    if (!date) return '-'

    // 如果 val 已经是 YYYY-MM-DD-HH:mm:ss 这种带横杠的奇怪格式，先尝试修复
    if (typeof date === 'string' && /^\d{4}-\d{2}-\d{2}-\d{2}:\d{2}:\d{2}$/.test(date)) {
        // 将第 10 个字符（日期和小时之间的横杠）替换为空格
        const fixed = date.slice(0, 10) + ' ' + date.slice(11)
        return dayjs.tz(fixed).format(format)
    }

    return dayjs.tz(date).format(format)
}

export { dayjs }
export default dayjs
