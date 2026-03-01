import {PUBLIC_API_URL} from "$env/static/public";

export class AppService {

    private static readonly instance: AppService = new AppService();

    public static getInstance(): AppService {
        return this.instance;
    }

    constructor() {
    }

}

export function getApiUrl(){
    const apiUrl = PUBLIC_API_URL;
    if(apiUrl.trim().length === 0){
        return new URL("/api", window.location.origin)
    }

    return new URL("", apiUrl);
}