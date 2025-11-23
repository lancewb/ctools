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

