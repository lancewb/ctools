import { createRouter, createWebHashHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import ToolTemplate from '../views/ToolTemplate.vue'
// 文本处理工具
import StringStats from '../views/text/StringStats.vue'
import Hexdump from '../views/text/Hexdump.vue'
import TextFormat from '../views/text/TextFormat.vue'
import RandomGen from '../views/text/RandomGen.vue'
import DataConvert from '../views/text/DataConvert.vue'
// 网络工具
import GroupPing from '../views/network/GroupPing.vue'
import LitePost from '../views/network/LitePost.vue'
import ServerManager from '../views/network/ServerManager.vue'
// 密码工具
import KeyManager from '../views/crypto/KeyManager.vue'
import AsymmetricOps from '../views/crypto/AsymmetricOps.vue'
import SymmetricOps from '../views/crypto/SymmetricOps.vue'
import HashHmac from '../views/crypto/HashHmac.vue'
import CertManager from '../views/crypto/CertManager.vue'
import CertParser from '../views/crypto/CertParser.vue'
import DerParser from '../views/crypto/DerParser.vue'
import QuantumPlaceholder from '../views/crypto/QuantumPlaceholder.vue'
// 其他工具
import SocksProxy from '../views/other/SocksProxy.vue'
import GMSSLTester from '../views/other/GMSSLTester.vue'
import MermaidWorkbench from '../views/other/MermaidWorkbench.vue'

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
    {
        // 对应 menu.js 中的 id: 'hexdump', catId: 'text'
        path: '/tool/text/hexdump',
        name: 'Hexdump',
        component: Hexdump
    },
    {
        // 对应 menu.js 中的 id: 'text-fmt', catId: 'text'
        path: '/tool/text/text-fmt',
        name: 'TextFormat',
        component: TextFormat
    },
    {
        // 对应 menu.js 中的 id: 'random-gen', catId: 'text'
        path: '/tool/text/random-gen',
        name: 'RandomGen',
        component: RandomGen
    },
    {
        // 对应 menu.js 中的 id: 'data-fmt', catId: 'text'
        path: '/tool/text/data-fmt',
        name: 'DataConvert',
        component: DataConvert
    },
    // 网络工具
    {
        // 对应 menu.js 中的 id: 'ping', catId: 'network'
        path: '/tool/network/ping',
        name: 'GroupPing',
        component: GroupPing
    },
    {
        path: '/tool/network/curl',
        name: 'LitePost',
        component: LitePost
    },
    {
        path: '/tool/network/server-mgr',
        name: 'ServerManager',
        component: ServerManager
    },
    // 密码工具
    {
        path: '/tool/crypto/key-mgr',
        name: 'KeyManager',
        component: KeyManager
    },
    {
        path: '/tool/crypto/asymmetric',
        name: 'AsymmetricOps',
        component: AsymmetricOps
    },
    {
        path: '/tool/crypto/symmetric',
        name: 'SymmetricOps',
        component: SymmetricOps
    },
    {
        path: '/tool/crypto/hash',
        name: 'HashHmac',
        component: HashHmac
    },
    {
        path: '/tool/crypto/cert-mgr',
        name: 'CertManager',
        component: CertManager
    },
    {
        path: '/tool/crypto/cert-parse',
        name: 'CertParser',
        component: CertParser
    },
    {
        path: '/tool/crypto/der-parse',
        name: 'DerParser',
        component: DerParser
    },
    {
        path: '/tool/crypto/quantum',
        name: 'QuantumPlaceholder',
        component: QuantumPlaceholder
    },
    // 其他工具
    {
        path: '/tool/other/socks',
        name: 'SocksProxy',
        component: SocksProxy
    },
    {
        path: '/tool/other/gmssl',
        name: 'GMSSLTester',
        component: GMSSLTester
    },
    {
        path: '/tool/other/mermaid',
        name: 'MermaidWorkbench',
        component: MermaidWorkbench
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
