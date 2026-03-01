import {AuthService} from "$lib/services/auth.service";
import {goto} from "$app/navigation";

export class ShareableService {
    private static readonly instance: ShareableService = new ShareableService();

    public static getInstance(): ShareableService {
        return this.instance;
    }

    public async create(
        request: ShareableServiceCreateRequest
    ){
        const authInstance = AuthService.getInstance();
        if(!authInstance.currentIdentity){
            await goto("/auth");
            return;
        }

        const apiRequest : CreateShareableRequest = {
            name: request.name || ShareableService.createShareName(),
            type: request.type,
            data: request.type === "FILE" ? request.d
        }
    }

    private static createShareName() : string{
        const authInstance = AuthService.getInstance();
        const currentIdentity = authInstance.currentIdentity!;
        return `Random Share from ${currentIdentity.email}`
    }
}

export type CreateShareableRequest = {
    name: string;
    type: ShareableType;
    data: string;
    allowed_emails?: string[];
    time_params: {
        expiry_at: string;       // ISO date string
        active_from?: string;    // ISO date string
    };
    notification_params?: {
        email_notification_on_open: boolean;
        notify_target_emails: boolean;
    };
    options: {
        only_once: boolean;
        encrypt: boolean;
    };
};

export type ShareableType =
    | "TEXT"
    | "FILE"
    | "LINK";

export type CreateShareableResponse = {
    code: string;
    url?: string;
};

export type ShareableServiceCreateRequest =
    | {
    type: "TEXT" | "LINK";
    name?: string;
    data: string;
    allowed_emails?: string[];
    expiry_at: Date;
    active_from?: Date;
    email_notification_on_open?: boolean;
    notify_target_emails?: boolean;
    only_once?: boolean;
    encrypt?: boolean;
}
    | {
    type: "FILE";
    name?: string;
    data: File[];
    allowed_emails?: string[];
    expiry_at: Date;
    active_from?: Date;
    email_notification_on_open?: boolean;
    notify_target_emails?: boolean;
    only_once?: boolean;
    encrypt?: boolean;
};