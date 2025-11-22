// src/config/menu.js
export const menuData = [
    {
        id: 'text',
        title: '文本处理', // 稍微缩短标题，防止换行
        icon: 'mdi-text-box-outline',
        color: 'blue-darken-2',
        children: [
            // { id: 'str-process', title: '字符串处理', icon: 'mdi-format-text' },
            { id: 'str-stats', title: '字符串统计', icon: 'mdi-chart-bar' },
            { id: 'hexdump', title: 'Hexdump', icon: 'mdi-code-json' }, // 名字太长也容易换行，精简一下
            { id: 'text-fmt', title: '文本格式化', icon: 'mdi-format-align-left' },
            { id: 'random-gen', title: '随机生成', icon: 'mdi-shuffle-variant' },
            { id: 'data-fmt', title: '数据转换', icon: 'mdi-swap-horizontal' },
        ]
    },
    {
        id: 'network',
        title: '网络工具',
        icon: 'mdi-ip-network',
        color: 'blue-darken-1',
        children: [
            { id: 'curl', title: 'Curl 模拟', icon: 'mdi-web' },
            { id: 'ping', title: '群 Ping', icon: 'mdi-access-point-network' },
            { id: 'ssh', title: 'SSH 终端', icon: 'mdi-console-line' },
            { id: 'server-mgr', title: '服务器管理', icon: 'mdi-server' },
        ]
    },
    {
        id: 'crypto',
        title: '密码工具',
        icon: 'mdi-shield-key',
        color: 'indigo',
        children: [
            { id: 'asymmetric', title: '非对称运算', icon: 'mdi-key-variant' },
            { id: 'symmetric', title: '对称运算', icon: 'mdi-lock-outline' },
            { id: 'hash', title: '哈希运算', icon: 'mdi-fingerprint' },
            { id: 'quantum', title: '量子运算', icon: 'mdi-atom' },
            { id: 'cert-parse', title: '证书解析', icon: 'mdi-file-certificate-outline' },
            { id: 'key-parse', title: '密钥解析', icon: 'mdi-key-chain' },
            { id: 'cert-mgr', title: '证书管理', icon: 'mdi-certificate' },
            { id: 'der-parse', title: 'DER 解析', icon: 'mdi-file-code-outline' },
        ]
    },
    {
        id: 'vpn',
        title: 'VPN 工具',
        icon: 'mdi-vpn',
        color: 'light-blue-darken-3',
        children: [
            { id: 'ssl-vpn', title: 'SSL VPN', icon: 'mdi-shield-check' },
            { id: 'ipsec-vpn', title: 'IPSec VPN', icon: 'mdi-lock-network' },
            { id: 'agent-check', title: 'Agent 检测', icon: 'mdi-face-agent' },
        ]
    }
];