export namespace api {
	
	export class AttackTypeInfo {
	    id: string;
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new AttackTypeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class DatabaseConfig {
	    path: string;
	    name: string;
	    isNew: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.isNew = source["isNew"];
	    }
	}
	export class DictEntry {
	    id: number;
	    category: string;
	    type: string;
	    content: string;
	    description: string;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new DictEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.type = source["type"];
	        this.content = source["content"];
	        this.description = source["description"];
	        this.source = source["source"];
	    }
	}
	export class HostCount {
	    host: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new HostCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.count = source["count"];
	    }
	}
	export class InitResult {
	    success: boolean;
	    message: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new InitResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class NetworkInterface {
	    name: string;
	    ips: string[];
	    mac: string;
	
	    static createFrom(source: any = {}) {
	        return new NetworkInterface(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.ips = source["ips"];
	        this.mac = source["mac"];
	    }
	}
	export class ProxyConfig {
	    proxyPort: number;
	    bindAddress: string;
	    apiPort: number;
	    enableHttps: boolean;
	    enableHttp2: boolean;
	    enableWebSocket: boolean;
	    upstreamEnabled: boolean;
	    upstreamAddress: string;
	    upstreamUsername: string;
	    upstreamPassword: string;
	    interceptMode: boolean;
	    certDir: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyPort = source["proxyPort"];
	        this.bindAddress = source["bindAddress"];
	        this.apiPort = source["apiPort"];
	        this.enableHttps = source["enableHttps"];
	        this.enableHttp2 = source["enableHttp2"];
	        this.enableWebSocket = source["enableWebSocket"];
	        this.upstreamEnabled = source["upstreamEnabled"];
	        this.upstreamAddress = source["upstreamAddress"];
	        this.upstreamUsername = source["upstreamUsername"];
	        this.upstreamPassword = source["upstreamPassword"];
	        this.interceptMode = source["interceptMode"];
	        this.certDir = source["certDir"];
	    }
	}
	export class VulnerabilityReport {
	    id: string;
	    name: string;
	    severity: string;
	    confidence: number;
	    url: string;
	    parameter: string;
	    status: string;
	    evidence: string;
	    remediation: string;
	    first_seen: string;
	    last_seen: string;
	    occurrences: number;
	
	    static createFrom(source: any = {}) {
	        return new VulnerabilityReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.severity = source["severity"];
	        this.confidence = source["confidence"];
	        this.url = source["url"];
	        this.parameter = source["parameter"];
	        this.status = source["status"];
	        this.evidence = source["evidence"];
	        this.remediation = source["remediation"];
	        this.first_seen = source["first_seen"];
	        this.last_seen = source["last_seen"];
	        this.occurrences = source["occurrences"];
	    }
	}
	export class VulnTypeCount {
	    type: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new VulnTypeCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.count = source["count"];
	    }
	}
	export class ReportSummary {
	    total: number;
	    by_severity: Record<string, number>;
	    by_status: Record<string, number>;
	    top_vuln_types: VulnTypeCount[];
	    top_hosts: HostCount[];
	
	    static createFrom(source: any = {}) {
	        return new ReportSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.by_severity = source["by_severity"];
	        this.by_status = source["by_status"];
	        this.top_vuln_types = this.convertValues(source["top_vuln_types"], VulnTypeCount);
	        this.top_hosts = this.convertValues(source["top_hosts"], HostCount);
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
	export class ReportData {
	    title: string;
	    generated_at: string;
	    session_id: string;
	    summary: ReportSummary;
	    vulns: VulnerabilityReport[];
	    metadata: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ReportData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.generated_at = source["generated_at"];
	        this.session_id = source["session_id"];
	        this.summary = this.convertValues(source["summary"], ReportSummary);
	        this.vulns = this.convertValues(source["vulns"], VulnerabilityReport);
	        this.metadata = source["metadata"];
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
	export class ReportOptions {
	    session_id: string;
	    title: string;
	    format: string;
	    severity: string[];
	    status: string[];
	
	    static createFrom(source: any = {}) {
	        return new ReportOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.session_id = source["session_id"];
	        this.title = source["title"];
	        this.format = source["format"];
	        this.severity = source["severity"];
	        this.status = source["status"];
	    }
	}
	
	export class Rule {
	    id: string;
	    name: string;
	    description: string;
	    severity: string;
	    enabled: boolean;
	    priority: number;
	    tags: string[];
	    remediation?: string;
	
	    static createFrom(source: any = {}) {
	        return new Rule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.severity = source["severity"];
	        this.enabled = source["enabled"];
	        this.priority = source["priority"];
	        this.tags = source["tags"];
	        this.remediation = source["remediation"];
	    }
	}
	export class ScanResult {
	    id: number;
	    pluginName: string;
	    pluginId: string;
	    severity: string;
	    title: string;
	    description: string;
	    url: string;
	    method: string;
	    evidence: string;
	    request: string;
	    response: string;
	    timestamp: string;
	    falsePositive: boolean;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.pluginName = source["pluginName"];
	        this.pluginId = source["pluginId"];
	        this.severity = source["severity"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.url = source["url"];
	        this.method = source["method"];
	        this.evidence = source["evidence"];
	        this.request = source["request"];
	        this.response = source["response"];
	        this.timestamp = source["timestamp"];
	        this.falsePositive = source["falsePositive"];
	        this.tags = source["tags"];
	    }
	}
	
	export class Vulnerability {
	    id: number;
	    title: string;
	    severity: string;
	    type: string;
	    url: string;
	    method: string;
	    request: string;
	    response: string;
	    description: string;
	    remediation: string;
	    references: string[];
	    status: string;
	    createdAt: string;
	    updatedAt: string;
	    source: string;
	    cwe: string;
	    cvss: number;
	
	    static createFrom(source: any = {}) {
	        return new Vulnerability(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.severity = source["severity"];
	        this.type = source["type"];
	        this.url = source["url"];
	        this.method = source["method"];
	        this.request = source["request"];
	        this.response = source["response"];
	        this.description = source["description"];
	        this.remediation = source["remediation"];
	        this.references = source["references"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.source = source["source"];
	        this.cwe = source["cwe"];
	        this.cvss = source["cvss"];
	    }
	}
	
	export class WebSocketMessage {
	    id: number;
	    timestamp: string;
	    direction: string;
	    type: string;
	    url: string;
	    host: string;
	    size: number;
	    content: string;
	    contentType: string;
	    connectionId: string;
	
	    static createFrom(source: any = {}) {
	        return new WebSocketMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.direction = source["direction"];
	        this.type = source["type"];
	        this.url = source["url"];
	        this.host = source["host"];
	        this.size = source["size"];
	        this.content = source["content"];
	        this.contentType = source["contentType"];
	        this.connectionId = source["connectionId"];
	    }
	}

}

export namespace intruder {
	
	export class PayloadPosition {
	    start: number;
	    end: number;
	
	    static createFrom(source: any = {}) {
	        return new PayloadPosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.start = source["start"];
	        this.end = source["end"];
	    }
	}
	export class AttackConfig {
	    id: string;
	    name: string;
	    attackType: string;
	    baseRequest: string;
	    method: string;
	    url: string;
	    headers: Record<string, string>;
	    body: string;
	    payloadSets: string[][];
	    positions: PayloadPosition[];
	    concurrency: number;
	    rateLimit: number;
	    followRedirects: boolean;
	    timeout: number;
	
	    static createFrom(source: any = {}) {
	        return new AttackConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.attackType = source["attackType"];
	        this.baseRequest = source["baseRequest"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.payloadSets = source["payloadSets"];
	        this.positions = this.convertValues(source["positions"], PayloadPosition);
	        this.concurrency = source["concurrency"];
	        this.rateLimit = source["rateLimit"];
	        this.followRedirects = source["followRedirects"];
	        this.timeout = source["timeout"];
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

}

export namespace models {
	
	export class DashboardMetrics {
	    qps: number;
	    avgResponseTime: number;
	    activeConnections: number;
	    totalRequests: number;
	    totalBytesIn: number;
	    totalBytesOut: number;
	    errorRate: number;
	    uptime: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.qps = source["qps"];
	        this.avgResponseTime = source["avgResponseTime"];
	        this.activeConnections = source["activeConnections"];
	        this.totalRequests = source["totalRequests"];
	        this.totalBytesIn = source["totalBytesIn"];
	        this.totalBytesOut = source["totalBytesOut"];
	        this.errorRate = source["errorRate"];
	        this.uptime = source["uptime"];
	    }
	}
	export class FingerprintResult {
	    url: string;
	    fingerprints: string[];
	    confidence: number;
	    processTime: number;
	    title: string;
	    statusCode: number;
	    timestamp: string;
	
	    static createFrom(source: any = {}) {
	        return new FingerprintResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.fingerprints = source["fingerprints"];
	        this.confidence = source["confidence"];
	        this.processTime = source["processTime"];
	        this.title = source["title"];
	        this.statusCode = source["statusCode"];
	        this.timestamp = source["timestamp"];
	    }
	}
	export class ProxyStatus {
	    running: boolean;
	    port: number;
	    interceptMode: boolean;
	    activeConnections: number;
	    totalRequests: number;
	    uptime: number;
	
	    static createFrom(source: any = {}) {
	        return new ProxyStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.port = source["port"];
	        this.interceptMode = source["interceptMode"];
	        this.activeConnections = source["activeConnections"];
	        this.totalRequests = source["totalRequests"];
	        this.uptime = source["uptime"];
	    }
	}
	export class RepeaterRequest {
	    id: string;
	    name: string;
	    method: string;
	    url: string;
	    headers: Record<string, string>;
	    body: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new RepeaterRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class RepeaterResponse {
	    statusCode: number;
	    statusText: string;
	    headers: Record<string, string>;
	    body: string;
	    responseTime: number;
	    contentLength: number;
	
	    static createFrom(source: any = {}) {
	        return new RepeaterResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.statusText = source["statusText"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.responseTime = source["responseTime"];
	        this.contentLength = source["contentLength"];
	    }
	}
	export class TrafficItem {
	    id: string;
	    timestamp: string;
	    method: string;
	    url: string;
	    host: string;
	    path: string;
	    statusCode: number;
	    contentType: string;
	    requestSize: number;
	    responseSize: number;
	    duration: number;
	    requestHeaders: Record<string, string>;
	    responseHeaders: Record<string, string>;
	    requestBody: string;
	    responseBody: string;
	    clientIP: string;
	    protocol: string;
	    intercepted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TrafficItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.host = source["host"];
	        this.path = source["path"];
	        this.statusCode = source["statusCode"];
	        this.contentType = source["contentType"];
	        this.requestSize = source["requestSize"];
	        this.responseSize = source["responseSize"];
	        this.duration = source["duration"];
	        this.requestHeaders = source["requestHeaders"];
	        this.responseHeaders = source["responseHeaders"];
	        this.requestBody = source["requestBody"];
	        this.responseBody = source["responseBody"];
	        this.clientIP = source["clientIP"];
	        this.protocol = source["protocol"];
	        this.intercepted = source["intercepted"];
	    }
	}

}

export namespace service {
	
	export class AppConfig {
	    lastConnectionMode: string;
	    localDataDir?: string;
	    localApiPort?: number;
	    localProxyPort?: number;
	    remoteHost?: string;
	    remotePort?: number;
	    remoteApiKey?: string;
	    theme?: string;
	    windowWidth?: number;
	    windowHeight?: number;
	    windowX?: number;
	    windowY?: number;
	    interceptEnabled: boolean;
	    defaultMethodFilter?: string;
	    defaultStatusFilter?: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lastConnectionMode = source["lastConnectionMode"];
	        this.localDataDir = source["localDataDir"];
	        this.localApiPort = source["localApiPort"];
	        this.localProxyPort = source["localProxyPort"];
	        this.remoteHost = source["remoteHost"];
	        this.remotePort = source["remotePort"];
	        this.remoteApiKey = source["remoteApiKey"];
	        this.theme = source["theme"];
	        this.windowWidth = source["windowWidth"];
	        this.windowHeight = source["windowHeight"];
	        this.windowX = source["windowX"];
	        this.windowY = source["windowY"];
	        this.interceptEnabled = source["interceptEnabled"];
	        this.defaultMethodFilter = source["defaultMethodFilter"];
	        this.defaultStatusFilter = source["defaultStatusFilter"];
	    }
	}
	export class LocalConfig {
	    dataDir: string;
	    apiPort: number;
	    proxyPort: number;
	
	    static createFrom(source: any = {}) {
	        return new LocalConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dataDir = source["dataDir"];
	        this.apiPort = source["apiPort"];
	        this.proxyPort = source["proxyPort"];
	    }
	}
	export class ModifiedRequestResult {
	    statusCode: number;
	    statusText: string;
	    headers: Record<string, string>;
	    body: string;
	    responseTime: number;
	    contentType: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ModifiedRequestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.statusText = source["statusText"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.responseTime = source["responseTime"];
	        this.contentType = source["contentType"];
	        this.error = source["error"];
	    }
	}
	export class RemoteConfig {
	    host: string;
	    port: number;
	    apiKey?: string;
	
	    static createFrom(source: any = {}) {
	        return new RemoteConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.apiKey = source["apiKey"];
	    }
	}

}

