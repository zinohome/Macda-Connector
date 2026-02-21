import { defineStore } from 'pinia';
import { ref } from 'vue';

/**
 * 实时信号存储
 * 统一管理来自 WebSockets 的实时消息
 */
export const useSignalStore = defineStore('signal', () => {
    const latestParsedSignal = ref<any>(null);
    const latestAlarm = ref<any>(null);
    const isConnected = ref(false);
    let ws: WebSocket | null = null;

    function connect() {
        if (ws) ws.close();

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        // 注意：这里的 /ws/signals 会被 Vite Proxy 转发到 BFF
        const wsUrl = `${protocol}//${window.location.host}/ws/signals`;

        ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            isConnected.value = true;
            console.log('[SignalStore] WebSocket connected');
        };

        ws.onmessage = (event) => {
            try {
                const { topic, data } = JSON.parse(event.data);
                if (topic === 'signal-parsed') {
                    latestParsedSignal.value = data;
                } else if (topic === 'signal-alarm') {
                    latestAlarm.value = data;
                }
            } catch (err) {
                console.error('[SignalStore] Failed to parse ws message', err);
            }
        };

        ws.onclose = () => {
            isConnected.value = false;
            console.log('[SignalStore] WebSocket disconnected, retrying in 5s...');
            setTimeout(connect, 5000);
        };

        ws.onerror = (err) => {
            console.error('[SignalStore] WebSocket error', err);
        };
    }

    return {
        latestParsedSignal,
        latestAlarm,
        isConnected,
        connect,
    };
});
