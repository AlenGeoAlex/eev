import {AuthService} from "$lib/services/auth.service";

export const ssr = false

import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async () => {
    const authentication = AuthService.getInstance();
    const auth = await authentication.me();
    return {
        auth
    };
};