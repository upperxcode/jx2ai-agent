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
	    viewContent: string;
	
	    static createFrom(source: any = {}) {
	        return new UIState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentFile = source["currentFile"];
	        this.attachedFiles = source["attachedFiles"];
	        this.viewContent = source["viewContent"];
	    }
	}

}

export namespace csync {
	
	export class Map_string__github_com_upperxcode_jx2ai_agent_api_internal_lsp_Client_ {
	
	
	    static createFrom(source: any = {}) {
	        return new Map_string__github_com_upperxcode_jx2ai_agent_api_internal_lsp_Client_(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

