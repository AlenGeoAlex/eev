import { AppService} from "./app-service.ts";
import {get, type Readable, readonly, writable, type Writable} from "svelte/store";
import {ResponseError} from "../../../../../shared/ts-client";


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
        try {
            const meResponse = await AppService.instance.apis.user.meGet();
            if(!meResponse){
                return undefined;
            }

            this._authStore.set({
                id: meResponse.id!,
                email: meResponse.email!,
                avatar: meResponse.avatar!
            });
            console.log("Auth store set");
            return meResponse;
        }catch (e){
            const error = e as ResponseError;
            //error.response.status === 401
            console.error("Error fetching user", e);
            return undefined;
        }
    }

    public async getGoogleAuthUrl() : Promise<string | undefined>{
        try {
            const apiResponse = await AppService.instance.apis.auth.authGoogleLoginGet();
            if(!apiResponse){
                return undefined;
            }

            return apiResponse.url;
        }catch (e){
            console.error("Error fetching auth url", e);
            return undefined;
        }
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

    public async validateCallback(code: string, state: string) : Promise<{
        success: boolean
        message?: string
        id?: string
        email?: string
        avatar?: string
    }> {

        try {
            const validateResponse = await AppService.instance.apis.auth.authGoogleCallbackPost({
                handlersGoogleCallbackRequest: {
                    code: code,
                    state: state
                }
            });

            if (!validateResponse) {
                return {
                    success: false,
                    message: "Unable to validate callback"
                }
            }

            this._authStore.set({
                id: validateResponse.id!,
                email: validateResponse.email!,
                avatar: validateResponse.avatar!
            })
            console.log("Auth store set");

            return {
                success: true,
                id: validateResponse.id,
                email: validateResponse.email,
                avatar: validateResponse.avatar
            }
        } catch (e) {
            console.error("Error validating callback", e);
            return {
                success: false,
                message: "e."
            }
        }
    }
}

export interface IIdentity {
    id: string;
    email: string;
    avatar: string;
}