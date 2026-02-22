import axios from 'axios'

// 创建axios实例
const instance = axios.create({
    // 关键点：不再硬编码 IP，由 Vite Proxy 或反向代理处理。由于api.js里手写了/api，此处必须为空才能避免成为/api/api/
    baseURL: '',
    timeout: 15000 // 修正原来的 2 小时延迟，改为 15 秒快速失败
})

// 默认配置
instance.defaults.headers['Content-Type'] = 'application/json'

// 请求拦截器
instance.interceptors.request.use(config => {
    // 已经移除了 Admin Secret 硬编码。
    // 未来在这里增加：config.headers.Authorization = `Bearer ${token}`
    return config
}, e => Promise.reject(e))

// 响应拦截器
instance.interceptors.response.use((response) => {
    // 直接返回业务数据
    return response.data
}, (error) => {
    console.error('[Network Error]', error);
    return Promise.reject(error)
})

const request = (url, method, datas) => {
    return instance({
        url,
        method: method || 'get',
        [method?.toLowerCase() === 'get' ? 'params' : 'data']: datas
    })
}

export default request