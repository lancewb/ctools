export namespace network {
	
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

}

