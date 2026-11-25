export namespace dpi {
	
	export class DpiList {
	    name: string;
	    description: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DpiList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.enabled = source["enabled"];
	    }
	}
	export class DpiListContent {
	    name: string;
	    path: string;
	    entries: string[];
	    total: number;
	    customized: boolean;
	    tooLarge: boolean;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new DpiListContent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.entries = source["entries"];
	        this.total = source["total"];
	        this.customized = source["customized"];
	        this.tooLarge = source["tooLarge"];
	        this.description = source["description"];
	    }
	}

}

export namespace main {
	
	export class ActiveConnectionSnapshot {
	    id: string;
	    process: string;
	    remote: string;
	    upload: number;
	    download: number;
	
	    static createFrom(source: any = {}) {
	        return new ActiveConnectionSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.process = source["process"];
	        this.remote = source["remote"];
	        this.upload = source["upload"];
	        this.download = source["download"];
	    }
	}
	export class DpiPresetTestResult {
	    label: string;
	    sourceRel: string;
	    profileName: string;
	    ok: boolean;
	    detail: string;
	
	    static createFrom(source: any = {}) {
	        return new DpiPresetTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.sourceRel = source["sourceRel"];
	        this.profileName = source["profileName"];
	        this.ok = source["ok"];
	        this.detail = source["detail"];
	    }
	}
	export class DpiTuneCandidateResult {
	    label: string;
	    sourceRel: string;
	    profileName: string;
	    presetAlive: boolean;
	    probeHits: number;
	    probeTotal: number;
	    detail: string;
	
	    static createFrom(source: any = {}) {
	        return new DpiTuneCandidateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.sourceRel = source["sourceRel"];
	        this.profileName = source["profileName"];
	        this.presetAlive = source["presetAlive"];
	        this.probeHits = source["probeHits"];
	        this.probeTotal = source["probeTotal"];
	        this.detail = source["detail"];
	    }
	}
	export class DpiTuneResult {
	    candidates: DpiTuneCandidateResult[];
	    bestSourceRel: string;
	    newProfileName: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new DpiTuneResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.candidates = this.convertValues(source["candidates"], DpiTuneCandidateResult);
	        this.bestSourceRel = source["bestSourceRel"];
	        this.newProfileName = source["newProfileName"];
	        this.message = source["message"];
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
	export class ImportedFile {
	    path: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportedFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.content = source["content"];
	    }
	}
	export class LogsState {
	    coreLogPath: string;
	    dpiLogPath: string;
	    coreLogTail: string;
	    dpiLogTail: string;
	    coreErrors: string;
	    dpiErrors: string;
	    coreConnLines: string;
	    dpiConnLines: string;
	
	    static createFrom(source: any = {}) {
	        return new LogsState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coreLogPath = source["coreLogPath"];
	        this.dpiLogPath = source["dpiLogPath"];
	        this.coreLogTail = source["coreLogTail"];
	        this.dpiLogTail = source["dpiLogTail"];
	        this.coreErrors = source["coreErrors"];
	        this.dpiErrors = source["dpiErrors"];
	        this.coreConnLines = source["coreConnLines"];
	        this.dpiConnLines = source["dpiConnLines"];
	    }
	}
	export class ProxyProcessDetail {
	    name: string;
	    pid: string;
	    connCount: number;
	    downloadBps: number;
	    uploadBps: number;
	    estimateNote: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyProcessDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.pid = source["pid"];
	        this.connCount = source["connCount"];
	        this.downloadBps = source["downloadBps"];
	        this.uploadBps = source["uploadBps"];
	        this.estimateNote = source["estimateNote"];
	    }
	}
	export class RealtimeStats {
	    connected: boolean;
	    status: string;
	    downloadBps: number;
	    uploadBps: number;
	    proxyProcessDetails: ProxyProcessDetail[];
	    activeConnections: ActiveConnectionSnapshot[];
	    sessionDownloadBytes: number;
	    sessionUploadBytes: number;
	    sessionElapsedSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new RealtimeStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.status = source["status"];
	        this.downloadBps = source["downloadBps"];
	        this.uploadBps = source["uploadBps"];
	        this.proxyProcessDetails = this.convertValues(source["proxyProcessDetails"], ProxyProcessDetail);
	        this.activeConnections = this.convertValues(source["activeConnections"], ActiveConnectionSnapshot);
	        this.sessionDownloadBytes = source["sessionDownloadBytes"];
	        this.sessionUploadBytes = source["sessionUploadBytes"];
	        this.sessionElapsedSeconds = source["sessionElapsedSeconds"];
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
	export class RouteDecision {
	    query: string;
	    normalized: string;
	    kind: string;
	    target: string;
	    source: string;
	    ruleLabel: string;
	    final: boolean;
	    note: string;
	    dpiOn: boolean;
	    proxyOn: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RouteDecision(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.normalized = source["normalized"];
	        this.kind = source["kind"];
	        this.target = source["target"];
	        this.source = source["source"];
	        this.ruleLabel = source["ruleLabel"];
	        this.final = source["final"];
	        this.note = source["note"];
	        this.dpiOn = source["dpiOn"];
	        this.proxyOn = source["proxyOn"];
	    }
	}
	export class RoutingSplitEntry {
	    pattern: string;
	    target: string;
	    system: boolean;
	    label?: string;
	    source: string;
	    userIndex: number;
	
	    static createFrom(source: any = {}) {
	        return new RoutingSplitEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pattern = source["pattern"];
	        this.target = source["target"];
	        this.system = source["system"];
	        this.label = source["label"];
	        this.source = source["source"];
	        this.userIndex = source["userIndex"];
	    }
	}
	export class ServiceHealth {
	    proxyLocalOk: boolean;
	    proxyListenPort: number;
	    vlessRemote: string;
	    dpiRunning: boolean;
	    dpiEnabled: boolean;
	    proxyEnabled: boolean;
	    activeServer: string;
	    activeDpiProfile: string;
	    dpiTool: string;
	    dpiConfigPath: string;
	    proxyPortOccupied: boolean;
	    dpiConnectivityWarn: boolean;
	    dpiConnectivityDetail?: string;
	    dpiConnectivityCheckedUnix: number;
	
	    static createFrom(source: any = {}) {
	        return new ServiceHealth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyLocalOk = source["proxyLocalOk"];
	        this.proxyListenPort = source["proxyListenPort"];
	        this.vlessRemote = source["vlessRemote"];
	        this.dpiRunning = source["dpiRunning"];
	        this.dpiEnabled = source["dpiEnabled"];
	        this.proxyEnabled = source["proxyEnabled"];
	        this.activeServer = source["activeServer"];
	        this.activeDpiProfile = source["activeDpiProfile"];
	        this.dpiTool = source["dpiTool"];
	        this.dpiConfigPath = source["dpiConfigPath"];
	        this.proxyPortOccupied = source["proxyPortOccupied"];
	        this.dpiConnectivityWarn = source["dpiConnectivityWarn"];
	        this.dpiConnectivityDetail = source["dpiConnectivityDetail"];
	        this.dpiConnectivityCheckedUnix = source["dpiConnectivityCheckedUnix"];
	    }
	}
	export class SubscriptionPreview {
	    ok: boolean;
	    url: string;
	    format: string;
	    providerLabel: string;
	    name: string;
	    total: number;
	    vless: number;
	    vmess: number;
	    trojan: number;
	    shadowsocks: number;
	    hysteria2: number;
	    hysteria: number;
	    tuic: number;
	    anytls: number;
	    socks: number;
	    http: number;
	    wireguard: number;
	    other: number;
	    expiresUnix: number;
	    totalBytes: number;
	    usedBytes: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new SubscriptionPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.url = source["url"];
	        this.format = source["format"];
	        this.providerLabel = source["providerLabel"];
	        this.name = source["name"];
	        this.total = source["total"];
	        this.vless = source["vless"];
	        this.vmess = source["vmess"];
	        this.trojan = source["trojan"];
	        this.shadowsocks = source["shadowsocks"];
	        this.hysteria2 = source["hysteria2"];
	        this.hysteria = source["hysteria"];
	        this.tuic = source["tuic"];
	        this.anytls = source["anytls"];
	        this.socks = source["socks"];
	        this.http = source["http"];
	        this.wireguard = source["wireguard"];
	        this.other = source["other"];
	        this.expiresUnix = source["expiresUnix"];
	        this.totalBytes = source["totalBytes"];
	        this.usedBytes = source["usedBytes"];
	        this.error = source["error"];
	    }
	}
	export class UIServer {
	    id: string;
	    name: string;
	    host: string;
	    port: string;
	    protocol: string;
	    source: string;
	    fromSubscription: boolean;
	    lastCheckOk: boolean;
	    lastCheckLatencyMs: number;
	    lastCheckDownloadMbps: number;
	    recommended: boolean;
	    premium: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UIServer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.source = source["source"];
	        this.fromSubscription = source["fromSubscription"];
	        this.lastCheckOk = source["lastCheckOk"];
	        this.lastCheckLatencyMs = source["lastCheckLatencyMs"];
	        this.lastCheckDownloadMbps = source["lastCheckDownloadMbps"];
	        this.recommended = source["recommended"];
	        this.premium = source["premium"];
	    }
	}
	export class UIState {
	    proxyEnabled: boolean;
	    proxyPort: number;
	    dpiEnabled: boolean;
	    elevated: boolean;
	    proxyMode: string;
	    activeDpiProfile: string;
	    routingProfile: string;
	    activeServerId: string;
	    coreStatus: string;
	    dpiRunning: boolean;
	    dnsPreset: string;
	    dnsPrimary: string;
	    dnsSecondary: string;
	    subscriptionUrl: string;
	    subscriptionExpiresUnix: number;
	    subscriptionProviderLabel: string;
	    telegramInfoUrl: string;
	    telegramChannelName: string;
	    licenseTier: string;
	    licenseExpiresUnix: number;
	    licenseActive: boolean;
	    deviceIdShort: string;
	    deviceIdFull: string;
	    licenseApiBase: string;
	    premiumServerMarkers: string[];
	    servers: UIServer[];
	    dpiProfiles: models.DPIProfile[];
	    routingProfiles: models.RoutingProfile[];
	    strictRoutingMode: boolean;
	    autoUpdateEnabled: boolean;
	    autostartEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UIState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyEnabled = source["proxyEnabled"];
	        this.proxyPort = source["proxyPort"];
	        this.dpiEnabled = source["dpiEnabled"];
	        this.elevated = source["elevated"];
	        this.proxyMode = source["proxyMode"];
	        this.activeDpiProfile = source["activeDpiProfile"];
	        this.routingProfile = source["routingProfile"];
	        this.activeServerId = source["activeServerId"];
	        this.coreStatus = source["coreStatus"];
	        this.dpiRunning = source["dpiRunning"];
	        this.dnsPreset = source["dnsPreset"];
	        this.dnsPrimary = source["dnsPrimary"];
	        this.dnsSecondary = source["dnsSecondary"];
	        this.subscriptionUrl = source["subscriptionUrl"];
	        this.subscriptionExpiresUnix = source["subscriptionExpiresUnix"];
	        this.subscriptionProviderLabel = source["subscriptionProviderLabel"];
	        this.telegramInfoUrl = source["telegramInfoUrl"];
	        this.telegramChannelName = source["telegramChannelName"];
	        this.licenseTier = source["licenseTier"];
	        this.licenseExpiresUnix = source["licenseExpiresUnix"];
	        this.licenseActive = source["licenseActive"];
	        this.deviceIdShort = source["deviceIdShort"];
	        this.deviceIdFull = source["deviceIdFull"];
	        this.licenseApiBase = source["licenseApiBase"];
	        this.premiumServerMarkers = source["premiumServerMarkers"];
	        this.servers = this.convertValues(source["servers"], UIServer);
	        this.dpiProfiles = this.convertValues(source["dpiProfiles"], models.DPIProfile);
	        this.routingProfiles = this.convertValues(source["routingProfiles"], models.RoutingProfile);
	        this.strictRoutingMode = source["strictRoutingMode"];
	        this.autoUpdateEnabled = source["autoUpdateEnabled"];
	        this.autostartEnabled = source["autostartEnabled"];
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
	export class UpdateInfo {
	    ok: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    latestTag: string;
	    upToDate: boolean;
	    releaseUrl: string;
	    downloadUrl: string;
	    installerDownloadUrl: string;
	    message: string;
	    isPrerelease: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.latestTag = source["latestTag"];
	        this.upToDate = source["upToDate"];
	        this.releaseUrl = source["releaseUrl"];
	        this.downloadUrl = source["downloadUrl"];
	        this.installerDownloadUrl = source["installerDownloadUrl"];
	        this.message = source["message"];
	        this.isPrerelease = source["isPrerelease"];
	    }
	}

}

export namespace models {
	
	export class DPIProfile {
	    name: string;
	    tool: string;
	    configPath: string;
	    args: string[];
	
	    static createFrom(source: any = {}) {
	        return new DPIProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.tool = source["tool"];
	        this.configPath = source["configPath"];
	        this.args = source["args"];
	    }
	}
	export class RouteRule {
	    name: string;
	    domains?: string[];
	    domainSuffixes?: string[];
	    cidrs?: string[];
	    processes?: string[];
	    action: string;
	
	    static createFrom(source: any = {}) {
	        return new RouteRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.domains = source["domains"];
	        this.domainSuffixes = source["domainSuffixes"];
	        this.cidrs = source["cidrs"];
	        this.processes = source["processes"];
	        this.action = source["action"];
	    }
	}
	export class RoutingProfile {
	    name: string;
	    rules: RouteRule[];
	    final?: string;
	
	    static createFrom(source: any = {}) {
	        return new RoutingProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.rules = this.convertValues(source["rules"], RouteRule);
	        this.final = source["final"];
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

