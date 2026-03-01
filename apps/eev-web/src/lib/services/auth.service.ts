import {getApiUrl} from "./app-service.ts";
import {get, type Readable, readonly, writable, type Writable} from "svelte/store";
import {string} from "zod";

export class AuthService {

    private static readonly instance: AuthService = new AuthService();

    public static getInstance(): AuthService {
        return this.instance;
    }

    private readonly _authStore : Writable<IIdentity | undefined> = writable(undefined);
    private readonly _readonlyAuthStore : Readable<IIdentity | undefined> = readonly(this._authStore);

    private constructor() {

    }

    public async me(){
        const res = await fetch("/api/me");
        if(!res.ok){
            return undefined;
        }

        const newVar = await res.json() as IIdentity;
        if (!newVar) {
            return undefined;
        }

        this._authStore.set(newVar);
        console.log("Auth store set", newVar);
        return newVar;
    }

    public async getGoogleAuthUrl() : Promise<string | undefined>{
        const res = await fetch('/api/auth/google');
        if(!res.ok){
            return undefined;
        }

        const data = await res.json();
        if (data.url) {
            return data.url;
        }

        return undefined;
    }

    public get authStore() : Readable<IIdentity | undefined> {
        return this._readonlyAuthStore;
    }

    public get currentIdentity() : IIdentity | undefined {
        return get<IIdentity | undefined>(this._readonlyAuthStore);
    }

    public logout(opts?: {
        redirectToLogout?: boolean;
    }) {
        this._authStore.set(undefined);
        if(opts?.redirectToLogout){
            window.location.href = "/logout";
        }
    }

    public async validateCallback(code: string, state: string) : Promise<ValidateCallbackResult>{
        const response = await fetch('/api/auth/google/callback', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ code, state }),
        });
        if(!response.ok){
            return {
                success: false,
                message: await response.text()
            }
        }

        const meResponse = await this.me();
        if(!meResponse){
            return {
                success: false,
                message: "Unable to validate callback"
            }
        }

        const data = meResponse;
        this._authStore.set(data);
        return {
            success: true,
            id: data.id,
            email: data.email,
            avatar: data.avatar
        }
    }
}

export interface IIdentity {
    id: string;
    email: string;
    avatar: string;
}

type ValidateCallbackResult =
    | { success: true; id: string; email: string; avatar?: string }
    | { success: false; message: string };