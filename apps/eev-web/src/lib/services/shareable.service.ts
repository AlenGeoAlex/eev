import {AuthService} from "$lib/services/auth.service";
import {AppService} from "$lib/services/app-service";
import {
    type HandlersShareableFileRequest,
    type HandlersShareableFileUpload,
    ResponseError
} from "../../../../../shared/ts-client";

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
                    },
                    files: request.type === "file" ? Array.from((request.data as File[]).map(x => {
                        return {
                            fileName: x.name,
                            fileSize: x.size,
                            contentType: x.type
                        } as HandlersShareableFileRequest
                    })) : []
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
                    id: shareableResponse.code!,
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
            if (e instanceof ResponseError){
                if (e.response.status === 413){
                    return {
                        status: 'error',
                        error: "The combined file size is exceeding the limit. Please split the file and try again."
                    };
                }
            }
            console.error("Error creating shareable", e);
            return {
                status: 'error',
                error: "Failed to create shareable"
            };
        }
    }

    public async uploadWithPreSignedUrl(
        presignedUrl: string,
        file: File,
        { onSuccess, onError, onProgress }: UploadCallbacks = {}
    ): Promise<void> {
        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();

            xhr.upload.addEventListener('progress', (event: ProgressEvent) => {
                if (event.lengthComputable) {
                    const percent = Math.round((event.loaded / event.total) * 100);
                    onProgress?.(percent, event.loaded, event.total);
                }
            });

            xhr.addEventListener('load', () => {
                if (xhr.status >= 200 && xhr.status < 300) {
                    onSuccess?.();
                    resolve();
                } else {
                    const err = new Error(`Upload failed: ${xhr.status}`);
                    onError?.(err);
                    reject(err);
                }
            });

            xhr.addEventListener('error', () => {
                const err = new Error('Upload failed: network error');
                onError?.(err);
                reject(err);
            });

            xhr.open('PUT', presignedUrl);
            xhr.setRequestHeader('Content-Type', file.type);
            xhr.send(file);
        });
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

interface UploadCallbacks {
    onSuccess?: () => void;
    onError?: (error: Error) => void;
    onProgress?: (percent: number, loaded: number, total: number) => void;
}