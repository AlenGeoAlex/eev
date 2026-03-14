<script lang="ts">
    import * as Card from "$lib/components/ui/card/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import { Badge } from "$lib/components/ui/badge/index.js";
    import {
        FileText,
        Link,
        Download,
        Clock,
        AlertTriangle,
        Timer,
        ArrowLeft,
        ExternalLink,
        File
    } from "@lucide/svelte";
    import { formatDistanceToNow, format } from "date-fns";
    import { onMount } from "svelte";
    import type {PageData} from "./$types";
    import {type HandlersGetShareableResponse, ResponseError} from "../../../../../../shared/ts-client";
    import { invalidateAll } from '$app/navigation';

    let {data} = $props();

    let error: string | null = $state(null);
    let shareableResponse = $state<HandlersGetShareableResponse | null>(null);
    let availableFrom = $state<number | null>(null);
    let timeRemaining = $state<string>("");

    async function loadData() {
        try {
            // @ts-ignore
            const status = data.pageData instanceof ResponseError ? data.pageData.response.status : 200;

            if (status === 418) {
                const availableFromHeader = (data.pageData as ResponseError).response.headers.get('X-S-AvailableFrom');
                if (availableFromHeader) {
                    availableFrom = new Date(availableFromHeader).getTime();
                    startTimer();
                } else {
                    error = "This share is not available yet.";
                }
                return;
            }

            if (status >= 400) {
                if (data.pageData instanceof ResponseError) {
                    const responseError = await data.pageData.response.json()
                    error = `${responseError.message || 'Unknown error'}`;
                } else {
                    error = `Error: The server returned an unexpected response. Status: ${status}`;
                }
                 return;
            }

            shareableResponse = await data.pageData as HandlersGetShareableResponse;
            if (shareableResponse?.type === 'url' && shareableResponse.data) {
                window.open(shareableResponse.data, '_blank');
            }

        } catch (e) {
            console.error(e);
            if (e instanceof ResponseError) {
                error = `API Error: ${e.response.status} ${e.response.statusText}`;
            } else {
                error = "An unexpected error occurred.";
            }
            console.error(e);
        }
    }

    function startTimer() {
        const updateTimer = () => {
            if (!availableFrom) return;
            const now = Date.now();
            const diff = availableFrom - now;

            if (diff <= 0) {
                availableFrom = null;
                loadData();
                return;
            }

            const seconds = Math.floor((diff / 1000) % 60);
            const minutes = Math.floor((diff / 1000 / 60) % 60);
            const hours = Math.floor((diff / (1000 * 60 * 60)));

            timeRemaining = `${hours}h ${minutes}m ${seconds}s`;
            setTimeout(updateTimer, 1000);
        };
        updateTimer();
    }

    onMount(() => {
        loadData();
    });

    function formatSize(bytes?: number) {
        if (!bytes) return "0 B";
        const k = 1024;
        const sizes = ["B", "KB", "MB", "GB", "TB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
    }
</script>

<div class="min-h-screen flex items-center justify-center p-4">
    <Card.Root class="w-full max-w-2xl shadow-lg">
        <Card.Header>
            <div class="flex items-center gap-2 mb-2">
                <Button variant="ghost" size="icon" href="/" aria-label="back">
                    <ArrowLeft class="w-4 h-4" />
                </Button>
                <div>
                    <Card.Title>
                        {#if shareableResponse?.name}
                            {shareableResponse.name}
                        {:else if availableFrom}
                            Share Pending
                        {:else if error}
                            Error
                        {:else}
                            Loading Share...
                        {/if}
                    </Card.Title>
                    <Card.Description>
                        {#if shareableResponse?.id}
                            Shared via eev
                        {:else if availableFrom}
                            This share will be available soon
                        {:else if error}
                            Something went wrong
                        {/if}
                    </Card.Description>
                </div>
            </div>
        </Card.Header>

        <Card.Content>
            {#if error}
                <div class="flex flex-col items-center justify-center py-8 text-destructive gap-4">
                    <AlertTriangle class="w-12 h-12" />
                    <p class="text-lg font-medium">{error}</p>
                    {#if !error.includes("revoked")}
                        <Button variant="outline" onclick={() => invalidateAll()}>Try Again</Button>
                    {/if}
                </div>
            {:else if availableFrom}
                <div class="flex flex-col items-center justify-center py-8 gap-4">
                    <Timer class="w-12 h-12 text-primary animate-pulse" />
                    <p class="text-xl font-mono font-bold">{timeRemaining}</p>
                    <p class="text-muted-foreground text-center">
                        This share is scheduled and will be available at:<br/>
                        <span class="font-medium">{format(availableFrom, "PPP pp")}</span>
                    </p>
                </div>
            {:else if shareableResponse}
                {@const code = shareableResponse}
                {@const files = shareableResponse.files}

                {#if code.type === 'text'}
                    <div class="bg-muted p-4 rounded-md relative group">
                        <pre class="whitespace-pre-wrap break-all font-mono text-sm">{code.data}</pre>
                        <Button
                            variant="secondary"
                            size="sm"
                            class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
                            onclick={() => {
                                if (navigator.clipboard && code.data) {
                                    navigator.clipboard.writeText(code.data);
                                }
                            }}
                        >
                            Copy
                        </Button>
                    </div>
                {:else if code.type === 'url'}
                    <div class="flex flex-col items-center justify-center py-6 gap-4">
                        <Link class="w-12 h-12 text-primary" />
                        <div class="text-center">
                            <p class="text-sm text-muted-foreground mb-1">Shared Link:</p>
                            <a href={code.data} target="_blank" rel="noopener noreferrer" class="text-lg font-medium text-primary hover:underline break-all">
                                {code.data}
                            </a>
                        </div>
                        <Button href={code.data} target="_blank" rel="noopener noreferrer">
                            <ExternalLink class="w-4 h-4 mr-2" />
                            Open Link
                        </Button>
                    </div>
                {:else if code.type === 'file'}
                    <div class="space-y-3">
                        {#if files && files.length > 0}
                            <div class="rounded-md border">
                                {#each files as file, i}
                                    <div class="flex items-center justify-between p-3 {i !== files.length - 1 ? 'border-b' : ''}">
                                        <div class="flex items-center gap-3">
                                            <File class="w-8 h-8 text-muted-foreground" />
                                            <div>
                                                <p class="font-medium text-sm leading-none">{file.fileName}</p>
                                                <p class="text-xs text-muted-foreground mt-1">
                                                    {formatSize(0)} • {file.contentType}
                                                </p>
                                            </div>
                                        </div>
                                        <Button variant="ghost" size="icon" href={file.signedUrl} download={file.fileName} target="_blank">
                                            <Download class="w-4 h-4" />
                                        </Button>
                                    </div>
                                {/each}
                            </div>
                        {:else}
                            <p class="text-center text-muted-foreground py-4">No files found.</p>
                        {/if}
                    </div>
                {/if}

                <!-- Metadata section -->
                <div class="mt-6 space-y-4">
                    {#if code.options?.only_once === 'true'}
                        <div class="flex items-center gap-2 p-3 bg-yellow-50 dark:bg-yellow-950/30 text-yellow-800 dark:text-yellow-200 rounded-md border border-yellow-200 dark:border-yellow-800/50">
                            <AlertTriangle class="w-4 h-4 shrink-0" />
                            <p class="text-xs font-medium">This share will be destructed after this view.</p>
                        </div>
                    {/if}

                    {#if code.expiryAt}
                        <div class="flex items-center gap-2 text-xs text-muted-foreground px-1">
                            <Clock class="w-3.5 h-3.5" />
                            <span>
                                Expires {formatDistanceToNow(new Date(code.expiryAt), { addSuffix: true })}
                                <span class="opacity-70">({format(new Date(code.expiryAt), "PPP p")})</span>
                            </span>
                        </div>
                    {/if}
                </div>
            {:else}
                <div class="flex flex-col items-center justify-center py-12 gap-4">
                    <div class="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
                    <p class="text-sm text-muted-foreground">Fetching share details...</p>
                </div>
            {/if}
        </Card.Content>

        <Card.Footer class="justify-center border-t bg-muted/30 py-3">
            <p class="text-[10px] text-muted-foreground uppercase tracking-widest font-semibold">Protected by eev</p>
        </Card.Footer>
    </Card.Root>
</div>
