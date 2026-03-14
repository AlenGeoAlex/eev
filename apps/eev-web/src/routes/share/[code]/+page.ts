import type { PageLoad } from './$types';
import {AppService} from "$lib/services/app-service";
import {ResponseError} from "../../../../../../shared/ts-client";

export const load: PageLoad  = async ({params}) => {
    try {
        const response = await AppService.instance.apis.shareable.shareCodeGetRaw({
            code: params.code
        });

        const shareableResponse = await response.value();
        return {
            pageData: shareableResponse,
        }
    }catch (e){
        if (e instanceof ResponseError){
            return  {
                pageData: e
            }
        }

        return  {
            pageData: undefined
        }
    }
}