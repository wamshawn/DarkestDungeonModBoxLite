export namespace box {
	
	export class ImportEntry {
	    chosen: boolean;
	    key: string;
	    title: string;
	    children: ImportEntry[];
	
	    static createFrom(source: any = {}) {
	        return new ImportEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.chosen = source["chosen"];
	        this.key = source["key"];
	        this.title = source["title"];
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
	export class ImportMod {
	    dst: string;
	    title: string;
	    kind: string;
	    entries: ImportEntry[];
	
	    static createFrom(source: any = {}) {
	        return new ImportMod(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dst = source["dst"];
	        this.title = source["title"];
	        this.kind = source["kind"];
	        this.entries = this.convertValues(source["entries"], ImportEntry);
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
	    steam: string;
	    mods: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game = source["game"];
	        this.steam = source["steam"];
	        this.mods = source["mods"];
	    }
	}

}

