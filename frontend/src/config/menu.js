export const menuData = [
  {
    id: 'text',
    title: '文本工具',
    icon: 'mdi-text-box-outline',
    color: 'blue-darken-2',
    children: [
      { id: 'str-stats', title: '字符串统计', icon: 'mdi-chart-bar' },
      { id: 'hexdump', title: 'Hexdump', icon: 'mdi-code-json' },
      { id: 'text-fmt', title: '文本格式化', icon: 'mdi-format-align-left' },
      { id: 'random-gen', title: '随机生成', icon: 'mdi-shuffle-variant' },
      { id: 'diff', title: '文本 Diff', icon: 'mdi-compare-horizontal' }
    ]
  },
  {
    id: 'codec',
    title: '编码转换',
    icon: 'mdi-swap-horizontal-bold',
    color: 'teal-darken-2',
    children: [
      { id: 'data-fmt', title: '单位换算', icon: 'mdi-calculator-variant' },
      { id: 'codec', title: 'URL/Base64/Hex', icon: 'mdi-code-braces' },
      { id: 'jwt', title: 'JWT 解析', icon: 'mdi-card-account-details-outline' },
      { id: 'time-uuid', title: '时间戳/UUID', icon: 'mdi-clock-outline' },
      { id: 'color', title: '颜色转换', icon: 'mdi-palette-outline' }
    ]
  },
  {
    id: 'network',
    title: '网络调试',
    icon: 'mdi-ip-network',
    color: 'blue-darken-1',
    children: [
      { id: 'curl', title: 'HTTP 客户端', icon: 'mdi-web' },
      { id: 'tcp-client', title: 'TCP 客户端', icon: 'mdi-console-network' },
      { id: 'port-scan', title: '端口扫描', icon: 'mdi-radar' },
      { id: 'dns', title: 'DNS 查询', icon: 'mdi-dns-outline' },
      { id: 'ping', title: '群 Ping', icon: 'mdi-access-point-network' }
    ]
  },
  {
    id: 'monitoring',
    title: '监控观测',
    icon: 'mdi-monitor-dashboard',
    color: 'green-darken-2',
    children: [
      { id: 'prometheus', title: 'Prometheus 监控', icon: 'mdi-chart-line' },
      { id: 'server-mgr', title: '服务器管理', icon: 'mdi-server' }
    ]
  },
  {
    id: 'crypto',
    title: '密码运算',
    icon: 'mdi-shield-key',
    color: 'indigo',
    children: [
      { id: 'asymmetric', title: '非对称运算', icon: 'mdi-key-variant' },
      { id: 'symmetric', title: '对称运算', icon: 'mdi-lock-outline' },
      { id: 'hash', title: '哈希/HMAC', icon: 'mdi-fingerprint' },
      { id: 'key-mgr', title: '密钥管理', icon: 'mdi-key-chain' },
      { id: 'quantum', title: '量子占位', icon: 'mdi-atom' }
    ]
  },
  {
    id: 'cert',
    title: '证书 ASN.1',
    icon: 'mdi-certificate-outline',
    color: 'deep-purple-darken-1',
    children: [
      { id: 'cert-parse', title: '证书解析', icon: 'mdi-file-certificate-outline' },
      { id: 'cert-mgr', title: '证书管理', icon: 'mdi-certificate' },
      { id: 'der-parse', title: 'DER 解析', icon: 'mdi-file-code-outline' },
      { id: 'gmssl', title: '国密 SSL 检测', icon: 'mdi-shield-lock' }
    ]
  },
  {
    id: 'dev',
    title: '开发辅助',
    icon: 'mdi-code-tags',
    color: 'cyan-darken-3',
    children: [
      { id: 'json', title: 'JSON 格式化', icon: 'mdi-code-json' },
      { id: 'regex', title: '正则测试', icon: 'mdi-regex' }
    ]
  },
  {
    id: 'other',
    title: '其他小工具',
    icon: 'mdi-tools',
    color: 'light-blue-darken-3',
    children: [
      { id: 'socks', title: 'SOCKS5 代理', icon: 'mdi-access-point' }
    ]
  }
]
