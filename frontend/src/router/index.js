import { createRouter, createWebHashHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import ToolTemplate from '../views/ToolTemplate.vue'

import StringStats from '../views/text/StringStats.vue'
import Hexdump from '../views/text/Hexdump.vue'
import TextFormat from '../views/text/TextFormat.vue'
import RandomGen from '../views/text/RandomGen.vue'
import DataConvert from '../views/text/DataConvert.vue'

import LitePost from '../views/network/LitePost.vue'
import GroupPing from '../views/network/GroupPing.vue'
import ServerManager from '../views/network/ServerManager.vue'
import DNSLookup from '../views/network/DNSLookup.vue'
import PortScan from '../views/network/PortScan.vue'
import TCPClient from '../views/network/TCPClient.vue'
import PrometheusMonitor from '../views/monitoring/PrometheusMonitor.vue'

import KeyManager from '../views/crypto/KeyManager.vue'
import AsymmetricOps from '../views/crypto/AsymmetricOps.vue'
import SymmetricOps from '../views/crypto/SymmetricOps.vue'
import HashHmac from '../views/crypto/HashHmac.vue'
import CertManager from '../views/crypto/CertManager.vue'
import CertParser from '../views/crypto/CertParser.vue'
import DerParser from '../views/crypto/DerParser.vue'
import QuantumPlaceholder from '../views/crypto/QuantumPlaceholder.vue'

import SocksProxy from '../views/other/SocksProxy.vue'
import GMSSLTester from '../views/other/GMSSLTester.vue'

import JsonFormatter from '../views/dev/JsonFormatter.vue'
import RegexTester from '../views/dev/RegexTester.vue'
import DiffTool from '../views/dev/DiffTool.vue'
import CodecTool from '../views/encode/CodecTool.vue'
import JWTDecoder from '../views/encode/JWTDecoder.vue'
import TimeUuid from '../views/encode/TimeUuid.vue'
import ColorTool from '../views/encode/ColorTool.vue'

const routes = [
  { path: '/', name: 'Dashboard', component: Dashboard },

  { path: '/tool/text/str-stats', name: 'StringStats', component: StringStats },
  { path: '/tool/text/hexdump', name: 'Hexdump', component: Hexdump },
  { path: '/tool/text/text-fmt', name: 'TextFormat', component: TextFormat },
  { path: '/tool/text/random-gen', name: 'RandomGen', component: RandomGen },
  { path: '/tool/text/diff', name: 'DiffTool', component: DiffTool },

  { path: '/tool/codec/data-fmt', name: 'DataConvert', component: DataConvert },
  { path: '/tool/codec/codec', name: 'CodecTool', component: CodecTool },
  { path: '/tool/codec/jwt', name: 'JWTDecoder', component: JWTDecoder },
  { path: '/tool/codec/time-uuid', name: 'TimeUuid', component: TimeUuid },
  { path: '/tool/codec/color', name: 'ColorTool', component: ColorTool },

  { path: '/tool/network/curl', name: 'LitePost', component: LitePost },
  { path: '/tool/network/tcp-client', name: 'TCPClient', component: TCPClient },
  { path: '/tool/network/port-scan', name: 'PortScan', component: PortScan },
  { path: '/tool/network/dns', name: 'DNSLookup', component: DNSLookup },
  { path: '/tool/network/ping', name: 'GroupPing', component: GroupPing },

  { path: '/tool/monitoring/prometheus', name: 'PrometheusMonitor', component: PrometheusMonitor },
  { path: '/tool/monitoring/server-mgr', name: 'ServerManager', component: ServerManager },

  { path: '/tool/crypto/key-mgr', name: 'KeyManager', component: KeyManager },
  { path: '/tool/crypto/asymmetric', name: 'AsymmetricOps', component: AsymmetricOps },
  { path: '/tool/crypto/symmetric', name: 'SymmetricOps', component: SymmetricOps },
  { path: '/tool/crypto/hash', name: 'HashHmac', component: HashHmac },
  { path: '/tool/crypto/quantum', name: 'QuantumPlaceholder', component: QuantumPlaceholder },

  { path: '/tool/cert/cert-mgr', name: 'CertManager', component: CertManager },
  { path: '/tool/cert/cert-parse', name: 'CertParser', component: CertParser },
  { path: '/tool/cert/der-parse', name: 'DerParser', component: DerParser },
  { path: '/tool/cert/gmssl', name: 'GMSSLTester', component: GMSSLTester },

  { path: '/tool/dev/json', name: 'JsonFormatter', component: JsonFormatter },
  { path: '/tool/dev/regex', name: 'RegexTester', component: RegexTester },

  { path: '/tool/other/socks', name: 'SocksProxy', component: SocksProxy },
  { path: '/tool/:catId/:id', name: 'Tool', component: ToolTemplate }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
