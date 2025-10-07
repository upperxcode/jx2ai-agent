export namespace api {
	
	export class CommandInfo {
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new CommandInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class FileInfo {
	    name: string;
	    isDir: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.isDir = source["isDir"];
	    }
	}
	export class UIState {
	    currentFile: string;
	    attachedFiles: string[];
	
	    static createFrom(source: any = {}) {
	        return new UIState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentFile = source["currentFile"];
	        this.attachedFiles = source["attachedFiles"];
	    }
	}

}

