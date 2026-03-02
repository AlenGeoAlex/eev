import {AuthService} from "$lib/services/auth.service";
import {AppService} from "$lib/services/app-service";

export class ShareableService {
    private static readonly instance: ShareableService = new ShareableService();

    public static getInstance(): ShareableService {
        return this.instance;
    }

    public async create(
        request: ShareableServiceCreateRequest
    ) : Promise<{
        status: 'created' | 'error' | 'encryption' | 'fileUpload',
        id?: string
        error?: string
        uploads?: {
            fileId: string;
            url: string;
            name: string;
            file: File
        }[]
    }>{
        try {
            const shareableResponse = await AppService.instance.apis.shareable.sharePost({
                handlersCreateShareableRequest: {
                    name: request.name || ShareableService.createShareName(),
                    type: request.type,
                    data: request.type === "file" ? "" :
                        request.encrypt ? "" : request.data,
                    allowedEmails: Array.from(request.allowed_emails || []),
                    timeParams: {
                        expiryAt: request.expiry_at.toISOString(),
                        activeFrom: request.active_from?.toISOString()
                    },
                    notificationParams: {
                        emailNotificationOnOpen: request.email_notification_on_open,
                        notifyTargetEmails: request.notify_target_emails
                    },
                    options: {
                        onlyOnce: request.only_once,
                        encrypt: request.encrypt
                    }
                }
            });

            if(!shareableResponse){
                return {
                    status: 'error',
                    error: "Failed to create shareable"
                };
            }

            if(request.encrypt){
                return {
                    status: 'encryption',
                    id: shareableResponse.code!
                }
            }

            if(request.type === "file"){
                return {
                    status: 'fileUpload',
                    uploads: (shareableResponse.uploads ?? []).map(file => {
                        return {
                            fileId: file.fileId!,
                            url: file.uploadUrl!,
                            name: file.fileName!,
                            file: request.data.find(x => x.name === file.fileName)!
                        }
                    }) || []
                }
            }

            return {
                status: 'created',
                id: shareableResponse.code!
            };
        }catch (e){
            console.error("Error creating shareable", e);
            return {
                status: 'error',
                error: "Failed to create shareable"
            };
        }
    }

    private static createShareName() : string{
        const authInstance = AuthService.getInstance();
        const currentIdentity = authInstance.currentIdentity!;
        return `Random Share from ${currentIdentity.email}`
    }
}



export type ShareableServiceCreateRequest =
    | {
    type: "text" | "url";
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
    type: "file";
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