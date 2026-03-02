import {AuthApi, Configuration, ShareableApi, UserApi} from "../../../../../shared/ts-client";


class Apis {

    private readonly _auth: AuthApi;
    private readonly _user: UserApi;
    private readonly _shareable: ShareableApi;

    constructor(private readonly configuration: Configuration) {
        this._auth = new AuthApi(configuration);
        this._user = new UserApi(configuration);
        this._shareable = new ShareableApi(configuration);
    }


    get auth(): AuthApi {
        return this._auth;
    }

    get user(): UserApi {
        return this._user;
    }

    get shareable(): ShareableApi {
        return this._shareable;
    }
}

export class AppService {

    private static readonly _instance: AppService = new AppService();

    public static get instance(): AppService {
        return this._instance;
    }

    private readonly configuration: Configuration;
    private readonly _apis: Apis;
    constructor() {
        this.configuration = new Configuration({
            basePath: "/api"
        })
        this._apis = new Apis(this.configuration);
    }

    public get apis(): Apis {
        return this._apis;
    }

}