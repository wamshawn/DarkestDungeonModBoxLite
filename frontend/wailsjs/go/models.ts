export namespace box {
	
	export class ImportArchiveFilePassword {
	    path: string;
	    password: string;
	    invalid: boolean;
	    children: ImportArchiveFilePassword[];
	
	    static createFrom(source: any = {}) {
	        return new ImportArchiveFilePassword(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.password = source["password"];
	        this.invalid = source["invalid"];
	        this.children = this.convertValues(source["children"], ImportArchiveFilePassword);
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
	export class ImportArchiveFileStats {
	    password: ImportArchiveFilePassword;
	
	    static createFrom(source: any = {}) {
	        return new ImportArchiveFileStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.password = this.convertValues(source["password"], ImportArchiveFilePassword);
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
	export class ImportEntry {
	    chosen: boolean;
	    key: string;
	    filename: string;
	    children: ImportEntry[];
	
	    static createFrom(source: any = {}) {
	        return new ImportEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.chosen = source["chosen"];
	        this.key = source["key"];
	        this.filename = source["filename"];
	        this.children = this.convertValues(source["children"], ImportEntry);
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
	export class VersionedModule {
	    version: Version;
	    previewIconFile: string;
	    updateDetails: string;
	    itemDescriptionShort: string;
	    itemDescription: string;
	
	    static createFrom(source: any = {}) {
	        return new VersionedModule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = this.convertValues(source["version"], Version);
	        this.previewIconFile = source["previewIconFile"];
	        this.updateDetails = source["updateDetails"];
	        this.itemDescriptionShort = source["itemDescriptionShort"];
	        this.itemDescription = source["itemDescription"];
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
	export class Module {
	    id: string;
	    publishId: string;
	    kind: string;
	    title: string;
	    remark: string;
	    // Go type: time
	    modifyAT: any;
	    previewIconFile: string;
	    version: Version;
	    versions: VersionedModule[];
	
	    static createFrom(source: any = {}) {
	        return new Module(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.publishId = source["publishId"];
	        this.kind = source["kind"];
	        this.title = source["title"];
	        this.remark = source["remark"];
	        this.modifyAT = this.convertValues(source["modifyAT"], null);
	        this.previewIconFile = source["previewIconFile"];
	        this.version = this.convertValues(source["version"], Version);
	        this.versions = this.convertValues(source["versions"], VersionedModule);
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
	export class Version {
	    major: number;
	    minor: number;
	    patch: number;
	
	    static createFrom(source: any = {}) {
	        return new Version(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.major = source["major"];
	        this.minor = source["minor"];
	        this.patch = source["patch"];
	    }
	}
	export class ModulePlan {
	    existed: boolean;
	    kind: string;
	    publishFileId: string;
	    version: Version;
	    title: string;
	    iconBase64: string;
	    filename: string;
	    isDir: boolean;
	    entries: ImportEntry[];
	    similar: Module[];
	
	    static createFrom(source: any = {}) {
	        return new ModulePlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.existed = source["existed"];
	        this.kind = source["kind"];
	        this.publishFileId = source["publishFileId"];
	        this.version = this.convertValues(source["version"], Version);
	        this.title = source["title"];
	        this.iconBase64 = source["iconBase64"];
	        this.filename = source["filename"];
	        this.isDir = source["isDir"];
	        this.entries = this.convertValues(source["entries"], ImportEntry);
	        this.similar = this.convertValues(source["similar"], Module);
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
	export class ImportPlan {
	    source: string;
	    archived?: ImportArchiveFileStats;
	    invalid: boolean;
	    modules: ModulePlan[];
	
	    static createFrom(source: any = {}) {
	        return new ImportPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.archived = this.convertValues(source["archived"], ImportArchiveFileStats);
	        this.invalid = source["invalid"];
	        this.modules = this.convertValues(source["modules"], ModulePlan);
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
	export class MakeModuleImportPlanParam {
	    filename: string;
	    archiveFilePasswords: ImportArchiveFilePassword;
	
	    static createFrom(source: any = {}) {
	        return new MakeModuleImportPlanParam(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.archiveFilePasswords = this.convertValues(source["archiveFilePasswords"], ImportArchiveFilePassword);
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
	
	
	export class Settings {
	    game: string;
	    workshop: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game = source["game"];
	        this.workshop = source["workshop"];
	    }
	}
	
	
	export class WorkshopModule {
	    id: string;
	    title: string;
	    icon: string;
	    synced: boolean;
	    version: Version;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new WorkshopModule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.icon = source["icon"];
	        this.synced = source["synced"];
	        this.version = this.convertValues(source["version"], Version);
	        this.tags = source["tags"];
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

