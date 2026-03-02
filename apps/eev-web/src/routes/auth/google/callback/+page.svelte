<script lang="ts">
    import { onMount } from "svelte";
    import { AuthService } from "$lib/services/auth.service";
    import type { PageProps } from "./$types";

    import { Avatar, AvatarImage, AvatarFallback } from "$lib/components/ui/avatar";
    import {goto} from "$app/navigation";

    let data: PageProps = $props();
    const stateVal = data.data.state;
    const code = data.data.code;

    type Status = "loading" | "success" | "error";
    let status = $state<Status>("loading");
    let message = $state("Signing you in…");
    let avatarUrl = $state<string | undefined>(undefined);
    let email = $state<string | undefined>(undefined);

    let imageFailed = $state(false);

    function handleError() {
        imageFailed = true;
    }

    const initials = $derived(() => {
        if (!email) return "AN";
        return email.slice(0, 2).toUpperCase();
    });

    if (!code || !stateVal) {
        status = "error";
        message = "Invalid login parameters. Please try again.";
    }

    onMount(async () => {
        if (status === "error") return;

        const auth = AuthService.getInstance();

        try {
            if (code && stateVal) {
                const response = await auth.validateCallback(code, stateVal);

                if (response.success) {
                    status = "success";
                    avatarUrl = response.avatar;
                    email = response.email;
                    message = `Welcome ${response.email}! Please wait while we redirect you...`;

                    setTimeout(() => goto("/share/"), 2000);
                    return;
                }

                status = "error";
                message = `Something went wrong (${response.message?.trim() ?? "N/A"}). Please try again.`;
            } else {
                status = "error";
                message = "No authorisation code received.";
            }
        } catch (err) {
            console.error(err);
            status = "error";
            message = "Login failed. Please try again.";
        }
    });
</script>

<div class="flex min-h-screen items-center justify-center">
    <div class="flex w-full max-w-sm flex-col items-center gap-5 rounded-2xl border border-black/5 dark:border-white/10 px-10 py-12 shadow-sm">

        <div class="relative flex h-20 w-20 items-center justify-center">

            {#if status === "loading"}
                <div class="absolute inset-0 rounded-full border-4 border-black/10 dark:border-white/10"></div>
                <div class="absolute inset-0 animate-spin rounded-full border-4 border-transparent border-t-blue-500"></div>
                <div class="h-3 w-3 animate-pulse rounded-full bg-blue-400"></div>

            {:else if status === "success"}
                <div class="relative animate-[scale-in_0.35s_cubic-bezier(0.34,1.56,0.64,1)_both]">
                    <div class="absolute inset-0 rounded-full bg-green-500/20 blur-xl"></div>

                    <Avatar class="relative h-20 w-20 border-2 border-green-500/40 shadow-lg ring-4 ring-green-500/20">
                        {#if avatarUrl}
                            <AvatarImage src={avatarUrl} alt="User avatar" onerror={handleError} />
                        {/if}
                        <AvatarFallback class="text-lg font-semibold">
                            {initials}
                        </AvatarFallback>
                    </Avatar>

                    <div class="absolute -bottom-1 -right-1 flex h-7 w-7 items-center justify-center rounded-full bg-green-500 shadow-md ring-2 ring-white dark:ring-black">
                        <svg
                                class="h-4 w-4 text-white"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="3"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                        >
                            <polyline points="20 6 9 17 4 12" />
                        </svg>
                    </div>
                </div>

            {:else}
                <div class="relative animate-[scale-in_0.35s_cubic-bezier(0.34,1.56,0.64,1)_both]">
                    <div class="absolute inset-0 rounded-full bg-red-500/25 blur-2xl"></div>

                    <div class="relative flex h-20 w-20 items-center justify-center rounded-full
                    bg-red-500/10 backdrop-blur-sm
                    border border-red-500/30
                    shadow-lg ring-4 ring-red-500/20">

                        <span class="text-3xl animate-[emoji-bounce_0.6s_ease]">
                            😕
                        </span>
                    </div>
                </div>
            {/if}
        </div>

        <div class="text-center">
            <p class="text-base font-medium text-gray-700 dark:text-gray-200">{message}</p>
            {#if status === "loading"}
                <p class="mt-1 text-sm text-gray-400 dark:text-gray-500">This won't take long</p>
            {/if}
        </div>

        {#if status === "error"}
            <a
            href="/auth"
            class="mt-1 rounded-lg bg-blue-500 px-5 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-blue-600 active:scale-95"
            >
            Back to login
            </a>
        {/if}
    </div>
</div>

<style>
    @keyframes draw {
        to { stroke-dashoffset: 0; }
    }
    @keyframes scale-in {
        from { opacity: 0; transform: scale(0.6); }
        to   { opacity: 1; transform: scale(1); }
    }

    @keyframes fade-in {
        to { opacity: 1; }
    }

    @keyframes emoji-bounce {
        0%   { transform: scale(0.6); opacity: 0; }
        60%  { transform: scale(1.15); }
        100% { transform: scale(1); opacity: 1; }
    }
</style>