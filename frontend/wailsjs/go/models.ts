export namespace crypto {
	
	export class AsymmetricRequest {
	    algorithm: string;
	    operation: string;
	    payloadIsHash: boolean;
	    keyId: string;
	    peerKeyId: string;
	    keyData: string;
	    keyFormat: string;
	    payload: string;
	    payloadFormat: string;
	    signature: string;
	    signatureFormat: string;
	    uid: string;
	    padding: string;
	    oaepHash: string;
	    mgf1Hash: string;
	    outputFormat: string;
	    kdf: string;
	    symmetricCipher: string;
	    macAlgorithm: string;
	    eccMode: string;
	
	    static createFrom(source: any = {}) {
	        return new AsymmetricRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.algorithm = source["algorithm"];
	        this.operation = source["operation"];
	        this.payloadIsHash = source["payloadIsHash"];
	        this.keyId = source["keyId"];
	        this.peerKeyId = source["peerKeyId"];
	        this.keyData = source["keyData"];
	        this.keyFormat = source["keyFormat"];
	        this.payload = source["payload"];
	        this.payloadFormat = source["payloadFormat"];
	        this.signature = source["signature"];
	        this.signatureFormat = source["signatureFormat"];
	        this.uid = source["uid"];
	        this.padding = source["padding"];
	        this.oaepHash = source["oaepHash"];
	        this.mgf1Hash = source["mgf1Hash"];
	        this.outputFormat = source["outputFormat"];
	        this.kdf = source["kdf"];
	        this.symmetricCipher = source["symmetricCipher"];
	        this.macAlgorithm = source["macAlgorithm"];
	        this.eccMode = source["eccMode"];
	    }
	}
	export class StoredKey {
	    id: string;
	    name: string;
	    algorithm: string;
	    keyType: string;
	    format: string;
	    usage: string[];
	    privatePem?: string;
	    publicPem?: string;
	    extra?: Record<string, string>;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new StoredKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.algorithm = source["algorithm"];
	        this.keyType = source["keyType"];
	        this.format = source["format"];
	        this.usage = source["usage"];
	        this.privatePem = source["privatePem"];
	        this.publicPem = source["publicPem"];
	        this.extra = source["extra"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class CertRecord {
	    id: string;
	    name: string;
	    algorithm: string;
	    usage: string;
	    certPem: string;
	    keyId?: string;
	    serial: string;
	    notBefore: string;
	    notAfter: string;
	    subject: Record<string, string>;
	    issuer: Record<string, string>;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new CertRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.algorithm = source["algorithm"];
	        this.usage = source["usage"];
	        this.certPem = source["certPem"];
	        this.keyId = source["keyId"];
	        this.serial = source["serial"];
	        this.notBefore = source["notBefore"];
	        this.notAfter = source["notAfter"];
	        this.subject = source["subject"];
	        this.issuer = source["issuer"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class CertExport {
	    cert: CertRecord;
	    key?: StoredKey;
	
	    static createFrom(source: any = {}) {
	        return new CertExport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cert = this.convertValues(source["cert"], CertRecord);
	        this.key = this.convertValues(source["key"], StoredKey);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CertIssueRequest {
	    commonName: string;
	    algorithm: string;
	    keySize: number;
	    validDays: number;
	    usage: string;
	    save: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CertIssueRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.commonName = source["commonName"];
	        this.algorithm = source["algorithm"];
	        this.keySize = source["keySize"];
	        this.validDays = source["validDays"];
	        this.usage = source["usage"];
	        this.save = source["save"];
	    }
	}
	export class CertIssueResult {
	    rootCa?: CertRecord;
	    certificates: CertRecord[];
	    keys: StoredKey[];
	
	    static createFrom(source: any = {}) {
	        return new CertIssueResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rootCa = this.convertValues(source["rootCa"], CertRecord);
	        this.certificates = this.convertValues(source["certificates"], CertRecord);
	        this.keys = this.convertValues(source["keys"], StoredKey);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CertParseRequest {
	    pem: string;
	
	    static createFrom(source: any = {}) {
	        return new CertParseRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pem = source["pem"];
	    }
	}
	export class CertParseResult {
	    subject: Record<string, string>;
	    issuer: Record<string, string>;
	    serial: string;
	    notBefore: string;
	    notAfter: string;
	    publicKeyAlgorithm: string;
	    signatureAlgorithm: string;
	    dnsNames: string[];
	    ipAddresses: string[];
	    sans: string[];
	    keyUsage: string[];
	    extKeyUsage: string[];
	    rawHex: string;
	
	    static createFrom(source: any = {}) {
	        return new CertParseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.subject = source["subject"];
	        this.issuer = source["issuer"];
	        this.serial = source["serial"];
	        this.notBefore = source["notBefore"];
	        this.notAfter = source["notAfter"];
	        this.publicKeyAlgorithm = source["publicKeyAlgorithm"];
	        this.signatureAlgorithm = source["signatureAlgorithm"];
	        this.dnsNames = source["dnsNames"];
	        this.ipAddresses = source["ipAddresses"];
	        this.sans = source["sans"];
	        this.keyUsage = source["keyUsage"];
	        this.extKeyUsage = source["extKeyUsage"];
	        this.rawHex = source["rawHex"];
	    }
	}
	
	export class DerNode {
	    tag: number;
	    class: string;
	    label?: string;
	    constructed: boolean;
	    length: number;
	    value?: string;
	    hex: string;
	    children?: DerNode[];
	
	    static createFrom(source: any = {}) {
	        return new DerNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag = source["tag"];
	        this.class = source["class"];
	        this.label = source["label"];
	        this.constructed = source["constructed"];
	        this.length = source["length"];
	        this.value = source["value"];
	        this.hex = source["hex"];
	        this.children = this.convertValues(source["children"], DerNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DerParseRequest {
	    name: string;
	    hexString: string;
	    base64: string;
	
	    static createFrom(source: any = {}) {
	        return new DerParseRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.hexString = source["hexString"];
	        this.base64 = source["base64"];
	    }
	}
	export class DerParseResult {
	    nodes: DerNode[];
	
	    static createFrom(source: any = {}) {
	        return new DerParseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodes = this.convertValues(source["nodes"], DerNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HashRequest {
	    algorithm: string;
	    mode: string;
	    input: string;
	    inputFormat: string;
	    key: string;
	    keyFormat: string;
	    outputFormat: string;
	
	    static createFrom(source: any = {}) {
	        return new HashRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.algorithm = source["algorithm"];
	        this.mode = source["mode"];
	        this.input = source["input"];
	        this.inputFormat = source["inputFormat"];
	        this.key = source["key"];
	        this.keyFormat = source["keyFormat"];
	        this.outputFormat = source["outputFormat"];
	    }
	}
	export class KeyGenRequest {
	    name: string;
	    algorithm: string;
	    keySize: number;
	    curve: string;
	    publicExponent: number;
	    usage: string[];
	    variant: string;
	    save: boolean;
	    uid: string;
	
	    static createFrom(source: any = {}) {
	        return new KeyGenRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.algorithm = source["algorithm"];
	        this.keySize = source["keySize"];
	        this.curve = source["curve"];
	        this.publicExponent = source["publicExponent"];
	        this.usage = source["usage"];
	        this.variant = source["variant"];
	        this.save = source["save"];
	        this.uid = source["uid"];
	    }
	}
	export class KeyParseRequest {
	    name: string;
	    algorithm: string;
	    format: string;
	    data: string;
	    usage: string[];
	    variant: string;
	    save: boolean;
	
	    static createFrom(source: any = {}) {
	        return new KeyParseRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.algorithm = source["algorithm"];
	        this.format = source["format"];
	        this.data = source["data"];
	        this.usage = source["usage"];
	        this.variant = source["variant"];
	        this.save = source["save"];
	    }
	}
	export class KeyParseResult {
	    stored: boolean;
	    key?: StoredKey;
	    privatePem?: string;
	    publicPem?: string;
	    summary: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new KeyParseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stored = source["stored"];
	        this.key = this.convertValues(source["key"], StoredKey);
	        this.privatePem = source["privatePem"];
	        this.publicPem = source["publicPem"];
	        this.summary = source["summary"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OperationResult {
	    output?: string;
	    verified: boolean;
	    details?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new OperationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.output = source["output"];
	        this.verified = source["verified"];
	        this.details = source["details"];
	    }
	}
	
	export class SymmetricRequest {
	    algorithm: string;
	    mode: string;
	    padding: string;
	    operation: string;
	    key: string;
	    keyFormat: string;
	    iv: string;
	    ivFormat: string;
	    nonce: string;
	    nonceFormat: string;
	    input: string;
	    inputFormat: string;
	    additionalData: string;
	    additionalDataFormat: string;
	    outputFormat: string;
	
	    static createFrom(source: any = {}) {
	        return new SymmetricRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.algorithm = source["algorithm"];
	        this.mode = source["mode"];
	        this.padding = source["padding"];
	        this.operation = source["operation"];
	        this.key = source["key"];
	        this.keyFormat = source["keyFormat"];
	        this.iv = source["iv"];
	        this.ivFormat = source["ivFormat"];
	        this.nonce = source["nonce"];
	        this.nonceFormat = source["nonceFormat"];
	        this.input = source["input"];
	        this.inputFormat = source["inputFormat"];
	        this.additionalData = source["additionalData"];
	        this.additionalDataFormat = source["additionalDataFormat"];
	        this.outputFormat = source["outputFormat"];
	    }
	}

}

export namespace network {
	
	export class RequestOption {
	    id: string;
	    method: string;
	    url: string;
	    headers: Record<string, string>;
	    body: string;
	    protocol: string;
	    tlsVersion: string;
	    timeout: number;
	
	    static createFrom(source: any = {}) {
	        return new RequestOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.protocol = source["protocol"];
	        this.tlsVersion = source["tlsVersion"];
	        this.timeout = source["timeout"];
	    }
	}
	export class CollectionItem {
	    id: string;
	    name: string;
	    request: RequestOption;
	
	    static createFrom(source: any = {}) {
	        return new CollectionItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.request = this.convertValues(source["request"], RequestOption);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PingResult {
	    id: number;
	    latency: number;
	
	    static createFrom(source: any = {}) {
	        return new PingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.latency = source["latency"];
	    }
	}
	
	export class ResponseResult {
	    statusCode: number;
	    headers: Record<string, string>;
	    body: string;
	    timeCost: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ResponseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.timeCost = source["timeCost"];
	        this.error = source["error"];
	    }
	}
	export class ServerConfig {
	    id: string;
	    name: string;
	    host: string;
	    port: string;
	    user: string;
	    authType: string;
	    password: string;
	    keyPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ServerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.authType = source["authType"];
	        this.password = source["password"];
	        this.keyPath = source["keyPath"];
	    }
	}
	export class ServerStatus {
	    id: string;
	    isOnline: boolean;
	    error: string;
	    cpuModel: string;
	    cpuUsage: string;
	    ramTotal: string;
	    ramUsed: string;
	    ramPercent: number;
	    diskSize: string;
	    diskUsed: string;
	    diskPercent: number;
	    pciDevices: string[];
	
	    static createFrom(source: any = {}) {
	        return new ServerStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.isOnline = source["isOnline"];
	        this.error = source["error"];
	        this.cpuModel = source["cpuModel"];
	        this.cpuUsage = source["cpuUsage"];
	        this.ramTotal = source["ramTotal"];
	        this.ramUsed = source["ramUsed"];
	        this.ramPercent = source["ramPercent"];
	        this.diskSize = source["diskSize"];
	        this.diskUsed = source["diskUsed"];
	        this.diskPercent = source["diskPercent"];
	        this.pciDevices = source["pciDevices"];
	    }
	}

}

export namespace other {
	
	export class GMSSLClientConfig {
	    serverIp: string;
	    port: number;
	    enableClientAuth: boolean;
	    signCertId: string;
	    signKeyId: string;
	    encCertId: string;
	    encKeyId: string;
	    skipVerify: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GMSSLClientConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serverIp = source["serverIp"];
	        this.port = source["port"];
	        this.enableClientAuth = source["enableClientAuth"];
	        this.signCertId = source["signCertId"];
	        this.signKeyId = source["signKeyId"];
	        this.encCertId = source["encCertId"];
	        this.encKeyId = source["encKeyId"];
	        this.skipVerify = source["skipVerify"];
	    }
	}
	export class GMSSLClientResult {
	    success: boolean;
	    message: string;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new GMSSLClientResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class GMSSLServerConfig {
	    listenIp: string;
	    port: number;
	    signCertId: string;
	    signKeyId: string;
	    encCertId: string;
	    encKeyId: string;
	    clientAuth: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GMSSLServerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.listenIp = source["listenIp"];
	        this.port = source["port"];
	        this.signCertId = source["signCertId"];
	        this.signKeyId = source["signKeyId"];
	        this.encCertId = source["encCertId"];
	        this.encKeyId = source["encKeyId"];
	        this.clientAuth = source["clientAuth"];
	    }
	}
	export class GMSSLServerStatus {
	    running: boolean;
	    address: string;
	    error: string;
	    startedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new GMSSLServerStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.address = source["address"];
	        this.error = source["error"];
	        this.startedAt = source["startedAt"];
	    }
	}
	export class Socks5Config {
	    listenIp: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new Socks5Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.listenIp = source["listenIp"];
	        this.port = source["port"];
	    }
	}
	export class Socks5Status {
	    running: boolean;
	    address: string;
	    activeConnections: number;
	    error: string;
	    lastControlMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new Socks5Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.address = source["address"];
	        this.activeConnections = source["activeConnections"];
	        this.error = source["error"];
	        this.lastControlMessage = source["lastControlMessage"];
	    }
	}

}

