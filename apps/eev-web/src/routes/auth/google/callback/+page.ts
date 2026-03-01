import type { PageLoad } from "./$types"

export const ssr = false

export const load: PageLoad = async ({ params, url}) => {
    const code = url.searchParams.get("code");
    const state = url.searchParams.get("state");

    return {
        code,
        state
    }
}