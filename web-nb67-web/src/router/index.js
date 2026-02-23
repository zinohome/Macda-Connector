import { createRouter, createWebHistory } from "vue-router";

export const router = createRouter({
    history: createWebHistory(),
    routes: [
        {
            path: "/",
            name: "/",
            redirect: '/KT',
            component: () => import("@/views/home/index.vue"),
            meta: {
                keepAlive: false // 不需要被缓存
            },
        },
        {
            path: "/KT",
            name: "home",
            component: () => import("@/views/home/index.vue"),
            meta: {
                keepAlive: false // 不需要被缓存
            },
        },
        {
            path: "/KT/trainDetail/:trainNo",
            name: "trainDetail",
            component: () => import("@/views/trainDetail/index.vue")
        },
        {
            path: "/KT/trainDetail",
            name: "trainDetailQuery",
            component: () => import("@/views/trainDetail/index.vue")
        },
        {
            path: "/KT/airConditioner/:trainCoach/:trainNum/:trainCrew",     //列车号   车厢号   机箱号
            name: "airConditionerParams",
            component: () => import("@/views/airConditioner/index.vue")
        },
        {
            path: "/KT/airConditioner",     //列车号   车厢号   机箱号
            name: "airConditioner",
            component: () => import("@/views/airConditioner/index.vue")
        },
        {
            path: "/KT/carCrewDetail/:trainNo",
            name: "carCrewDetail",
            component: () => import("@/views/carCrewDetail/index.vue")
        },
        {
            path: "/KT/carCrewDetail",
            name: "carCrewDetailQuery",
            component: () => import("@/views/carCrewDetail/index.vue")
        },
        {
            path: "/KT/traininfo",
            name: "trainInfo",
            component: () => import("@/views/trainInfo/index.vue"),
            meta: {
                keepAlive: false // 不需要被缓存
            },
        },
        {
            path: "/historyData",
            name: "historyData",
            component: () => import("@/views/historyData/index.vue")
        },
        {
            path: "/historyAlarm",
            name: "historyAlarm",
            component: () => import("@/views/historyAlarm/index.vue")
        },
        {
            path: "/KT/carrigeinfo",

            name: "carrigeinfo",
            component: () => import("@/views/airConditioner/index.vue")
        }
    ]
})