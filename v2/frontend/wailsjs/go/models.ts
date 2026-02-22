export namespace main {
	
	export class AccountDTO {
	    id: string;
	    name: string;
	    email: string;
	    hasSession: boolean;
	    sessionCaptured: string;
	
	    static createFrom(source: any = {}) {
	        return new AccountDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.hasSession = source["hasSession"];
	        this.sessionCaptured = source["sessionCaptured"];
	    }
	}
	export class SettingsDTO {
	    launcherPath: string;
	    language: string;
	    streamerMode: boolean;
	    theme: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.launcherPath = source["launcherPath"];
	        this.language = source["language"];
	        this.streamerMode = source["streamerMode"];
	        this.theme = source["theme"];
	    }
	}
	export class SwitchResultDTO {
	    success: boolean;
	    accountName: string;
	    email: string;
	    hasSession: boolean;
	    message: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SwitchResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.accountName = source["accountName"];
	        this.email = source["email"];
	        this.hasSession = source["hasSession"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}

}

