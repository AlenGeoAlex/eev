import {redirect} from "@sveltejs/kit";
import type { PageLoad } from "./$types";
import {AuthService} from "$lib/services/auth.service";
import {goto} from "$app/navigation";


export const load: PageLoad = async ({parent}) => {
    const data = await parent();
    const refreshedData = AuthService.getInstance().currentIdentity
    if(!data.auth && !refreshedData){
        await goto("/auth");
        return {
            auth: undefined
        }
    }

    return {
        auth: data.auth
    }
}