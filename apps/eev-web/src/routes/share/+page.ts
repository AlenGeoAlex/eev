import {redirect} from "@sveltejs/kit";
import type { PageLoad } from "./$types";


export const load: PageLoad = async ({parent}) => {
    const data = await parent();
    if(!data.auth){
        redirect(302, "/auth");
        return {
            auth: undefined
        }
    }

    return {
        auth: data.auth
    }
}