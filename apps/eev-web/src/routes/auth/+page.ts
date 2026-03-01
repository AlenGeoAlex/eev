import {AuthService} from "$lib/services/auth.service";
import {redirect} from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
    const instance = AuthService.getInstance();
    const redirectUrl = await instance.getGoogleAuthUrl();
    if(!redirectUrl){
        return;
    }
    redirect(302, redirectUrl);
}