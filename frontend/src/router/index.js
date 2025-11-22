import { createRouter, createWebHashHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import ToolTemplate from '../views/ToolTemplate.vue'
// 文本处理工具
import StringStats from '../views/text/StringStats.vue'
// 网络工具
import GroupPing from '../views/network/GroupPing.vue'
// 密码工具
// VPN

const routes = [
    {
        path: '/',
        name: 'Dashboard',
        component: Dashboard
    },
    // 文本处理工具
    {
        path: '/tool/text/str-stats',
        name: 'StringStats',
        component: StringStats
    },
    // 网络工具
    {
        // 对应 menu.js 中的 id: 'ping', catId: 'network'
        path: '/tool/network/ping',
        name: 'GroupPing',
        component: GroupPing
    },
    {
        path: '/tool/:catId/:id',
        name: 'Tool',
        component: ToolTemplate
    }
]

const router = createRouter({
    history: createWebHashHistory(),
    routes
})

export default router