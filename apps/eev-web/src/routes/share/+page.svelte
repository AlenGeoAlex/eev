<script lang="ts" context="module">
    import {z} from "zod";

    export const shareSchema = z
        .object({
            targetEmails: z
                .array(z.email({
                    error: "Please enter a valid email address.",
                })),

            activeFrom: z.string().nullable(),
            expiresAt: z.string().min(1, "Expiry date is required.")
                .default(new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString().slice(0, 16)),

            notifyOnOpen: z.boolean().default(false),
            notifyTargetUsers: z.boolean().default(false),

            // Options
            allowOnce: z.boolean().default(true),
            encrypt: z.boolean().default(false),
        })
        .refine(
            (data) =>
                !data.activeFrom ||
                !data.expiresAt ||
                new Date(data.expiresAt) > new Date(data.activeFrom),
            {
                message: "Expiry must be after the active from date.",
                path: ["expiresAt"],
            }
        );

    export type ShareSchema = typeof shareSchema;
    export type ShareFormData = z.infer<typeof shareSchema>;
</script>
<script lang="ts">
    import * as Card from "$lib/components/ui/card";
    import * as Avatar from "$lib/components/ui/avatar";
    import { Textarea } from "$lib/components/ui/textarea";
    import { Label } from "$lib/components/ui/label";
    import { Button } from "$lib/components/ui/button";
    import { Checkbox } from "$lib/components/ui/checkbox";
    import { Switch } from "$lib/components/ui/switch";
    import { Badge } from "$lib/components/ui/badge";
    import { Input } from "$lib/components/ui/input";
    import { Upload, FileText, Mail, X, CalendarClock, Clock4, Check, Loader2, Plus } from "@lucide/svelte";
    import { superForm, defaults } from "sveltekit-superforms";
    import type { PageProps } from "./$types";
    import { format, formatDistance, formatRelative, subDays } from 'date-fns'
    import {zod4} from "sveltekit-superforms/adapters";
    import * as Tooltip from "$lib/components/ui/tooltip";
    import {ShareableService} from "$lib/services/shareable.service";
    import FileUploadDialog from "$lib/components/share/FileUploadDialog.svelte";
    import SuccessDialog from "$lib/components/share/SuccessDialog.svelte";

    let { data }: PageProps = $props();
    let dragOver = $state(false);
    let uploadedFiles: File[] = $state([]);
    let textValue = $state("");
    const { form, errors, validate, submitting } = superForm(
        defaults(zod4(shareSchema)),
        {
            SPA: true,
            validators: zod4(shareSchema),
            dataType: "json",
        }
    );

    let submittingToBackend = $state(false);
    let submitError = $state("");
    let submitSuccess = $state(false);

    let showFileUpload = $state(false);
    let showSuccess = $state(false);
    let uploadsData = $state<any[]>([]);
    let currentShareId = $state("");

    let hasFiles = $derived(uploadedFiles.length > 0);
    let hasText = $derived(textValue.trim() !== "");
    let hasErrors = $derived(Object.keys($errors).length > 0);
    let canSubmit = $derived(hasFiles || hasText);

    function handleDragOver(e: DragEvent) {
        e.preventDefault();
        dragOver = true;
    }
    function handleDragLeave() {
        dragOver = false;
    }
    function handleDrop(e: DragEvent) {
        e.preventDefault();
        dragOver = false;
        const files = e.dataTransfer?.files;
        if (files && files.length > 0) {
            uploadedFiles = [...uploadedFiles, ...Array.from(files)];
        }
    }
    function handleFileInput(e: Event) {
        const input = e.target as HTMLInputElement;
        const files = input.files;
        if (files && files.length > 0) {
            uploadedFiles = [...uploadedFiles, ...Array.from(files)];
        }
        input.value = "";
    }
    function removeFile(index: number) {
        uploadedFiles = uploadedFiles.filter((_, i) => i !== index);
    }
    function clearFiles() {
        uploadedFiles = [];
    }

    function formatFileSize(bytes: number): string {
        if (bytes < 1024) return `${bytes} B`;
        if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
        return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    }

    interface UserSuggestion {
        email: string;
        is_starred?: boolean;
    }

    let emailInput = $state("");
    let emailError = $state("");
    let suggestions = $state<UserSuggestion[]>([]);
    let isLoadingSuggestions = $state(false);
    let isDropdownOpen = $state(false);
    let searchDebounce: ReturnType<typeof setTimeout>;

    async function searchUsers(query: string): Promise<UserSuggestion[]> {
        const response = await ShareableService.getInstance().getTargetUsers(query);
        return response.map((user: any) => ({
            email: user.email,
            is_starred: user.is_starred ?? false
        }));
    }

    function isValidEmail(email: string) {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
    }

    async function handleEmailInput() {
        clearTimeout(searchDebounce);
        emailError = "";

        if (emailInput.length < 2) {
            suggestions = [];
            isDropdownOpen = false;
            return;
        }

        searchDebounce = setTimeout(async () => {
            isLoadingSuggestions = true;
            try {
                suggestions = await searchUsers(emailInput)
                isDropdownOpen = true;
            } catch {
                suggestions = [];
            } finally {
                isLoadingSuggestions = false;
            }
        }, 300);
    }

    function selectSuggestion(user: UserSuggestion) {
        if ($form.targetEmails.includes(user.email)) return;
        $form.targetEmails = [...$form.targetEmails, user.email];
        emailInput = "";
        suggestions = [];
        isDropdownOpen = false;
        emailError = "";
    }

    function addEmailTag() {
        const email = emailInput.trim().replace(/,$/, "");
        if (!email) return;
        if (!isValidEmail(email)) { emailError = "Please enter a valid email address."; return; }
        if ($form.targetEmails.includes(email)) { emailError = "Already added."; return; }
        $form.targetEmails = [...$form.targetEmails, email];
        emailInput = "";
        emailError = "";
        suggestions = [];
        isDropdownOpen = false;
    }

    function removeEmailTag(email: string) {
        $form.targetEmails = $form.targetEmails.filter((t) => t !== email);
    }

    function handleEmailKeydown(e: KeyboardEvent) {
        if (e.key === "Enter" || e.key === ",") {
            e.preventDefault();
            addEmailTag();
        }
        if (e.key === "Backspace" && emailInput === "" && $form.targetEmails.length > 0) {
            $form.targetEmails = $form.targetEmails.slice(0, -1);
        }
        if (e.key === "Escape") {
            isDropdownOpen = false;
        }
    }

    function handleEmailFocus() {
        if (emailInput.length >= 2 && suggestions.length > 0) {
            isDropdownOpen = true;
        }
    }

    // Click-outside action to close dropdown
    function clickOutside(node: HTMLElement, handler: () => void) {
        const listener = (e: MouseEvent) => {
            if (!node.contains(e.target as Node)) handler();
        };
        document.addEventListener("mousedown", listener);
        return { destroy: () => document.removeEventListener("mousedown", listener) };
    }

    function inferType() : "text" | "url" | "file" {
        if (uploadedFiles.length > 0) return "file";
        const text = textValue.trim();
        try {
            new URL(text);
            return "url";
        } catch(e) {
            return "text";
        }
    }

    async function onFormSubmit(e: SubmitEvent) {
        e.preventDefault();
        if (hasErrors) return;
        const type = inferType();
        const formData = $form;
        submittingToBackend = true;
        submitError = "";

        const response = await ShareableService.getInstance().create({
            name: undefined,
            type: type,
            data: type === 'file' ? uploadedFiles : textValue,
            expiry_at: new Date(formData.expiresAt),
            active_from: formData.activeFrom ? new Date(formData.activeFrom) : new Date(),
            email_notification_on_open: formData.notifyOnOpen,
            encrypt: formData.encrypt,
            allowed_emails: formData.targetEmails,
            only_once: formData.allowOnce,
            notify_target_emails: formData.notifyTargetUsers
        } as any);

        submittingToBackend = false;

        if (response.status === 'fileUpload') {
            uploadsData = response.uploads || [];
            currentShareId = response.id!;
            showFileUpload = true;
        } else if (response.status === 'created') {
            currentShareId = response.id!;
            showSuccess = true;
        } else if (response.status === 'error') {
            console.error(response);
            submitError = response.error || "Failed to create shareable";
        }
    }

    function handleFileUploadComplete() {
        showFileUpload = false;
        showSuccess = true;
    }

    function createNew() {
        uploadedFiles = [];
        textValue = "";
        $form.targetEmails = [];
        $form.activeFrom = null;
        $form.expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString().slice(0, 16);
        $form.notifyOnOpen = false;
        $form.notifyTargetUsers = false;
        $form.allowOnce = true;
        $form.encrypt = false;
        showSuccess = false;
        currentShareId = "";
        submitError = "";
        submitSuccess = false;
    }
</script>

<form onsubmit={(event) => onFormSubmit(event)}>
    <div class="min-h-screen bg-background p-8 flex flex-col items-center justify-center gap-6">

        <Card.Root class="w-full max-w-4xl">
            <Card.Content class="flex items-center gap-4 pt-6 pb-6">
                <Avatar.Root class="h-14 w-14">
                    {#if data.auth?.avatar}
                        <Avatar.Image src={data.auth.avatar} alt="User avatar" />
                    {:else}
                        <Avatar.Fallback class="text-lg font-semibold">AN</Avatar.Fallback>
                    {/if}
                </Avatar.Root>
                <div class="flex flex-col gap-0.5">
                    <p class="text-sm font-medium leading-none">Hey!</p>
                    <div class="flex items-center gap-1.5 text-muted-foreground">
                        <Mail class="h-3.5 w-3.5" />
                        <p class="text-sm">{data.auth?.email}</p>
                    </div>
                </div>
            </Card.Content>
        </Card.Root>

        <div class="w-full max-w-4xl grid grid-cols-1 md:grid-cols-[1fr_auto_1fr] items-stretch">

            <Tooltip.Provider>
                <Tooltip.Root disabled={!hasText}>
                    <Tooltip.Trigger class="flex">
                        <Card.Root class="flex flex-col flex-1 transition-opacity {hasText ? 'opacity-50 pointer-events-none' : ''}">
                            <Card.Header>
                                <Card.Title class="text-base">Upload Files</Card.Title>
                                <Card.Description>Drag & drop or click to select one or more files.</Card.Description>
                            </Card.Header>
                            <Card.Content class="flex-1 flex flex-col gap-3">
                                <label
                                        class="flex flex-col items-center justify-center w-full h-28 border-2 border-dashed rounded-lg cursor-pointer transition-colors
                                        {dragOver ? 'border-primary bg-primary/5' : 'border-border bg-muted/30 hover:bg-muted/50 hover:border-muted-foreground/40'}"
                                        ondragover={handleDragOver}
                                        ondragleave={handleDragLeave}
                                        ondrop={handleDrop}
                                >
                                    <div class="flex flex-col items-center gap-2 text-muted-foreground pointer-events-none select-none">
                                        <Upload class="h-6 w-6" />
                                        <p class="text-sm font-medium">
                                            {hasFiles ? "Drop more files here" : "Drop files here"}
                                        </p>
                                        <p class="text-xs">or click to browse</p>
                                    </div>
                                    <input type="file" multiple disabled={hasText} class="hidden" onchange={handleFileInput} />
                                </label>

                                {#if hasFiles}
                                    <div class="flex flex-col gap-1.5">
                                        {#each uploadedFiles as file, index}
                                            <div class="flex items-center gap-2 rounded-md border border-border bg-muted/30 px-3 py-2">
                                                <FileText class="h-4 w-4 shrink-0 text-primary" />
                                                <div class="flex-1 min-w-0">
                                                    <p class="text-sm font-medium truncate">{file.name}</p>
                                                    <p class="text-xs text-muted-foreground">{formatFileSize(file.size)}</p>
                                                </div>
                                                <button
                                                        type="button"
                                                        onclick={() => removeFile(index)}
                                                        class="shrink-0 rounded-full hover:bg-muted-foreground/20 p-1 transition-colors text-muted-foreground hover:text-foreground"
                                                        aria-label="Remove {file.name}"
                                                >
                                                    <X class="h-3.5 w-3.5" />
                                                </button>
                                            </div>
                                        {/each}
                                        <button
                                                type="button"
                                                onclick={clearFiles}
                                                class="text-xs text-muted-foreground hover:text-foreground text-right transition-colors"
                                        >
                                            Remove all
                                        </button>
                                    </div>
                                {/if}
                            </Card.Content>
                        </Card.Root>
                    </Tooltip.Trigger>
                    <Tooltip.Content>
                        <p>Only one share can be made at a time. Please remove the text to share files.</p>
                    </Tooltip.Content>
                </Tooltip.Root>
            </Tooltip.Provider>

            <div class="hidden md:flex flex-col items-center justify-center gap-3 px-4">
                <div class="flex-1 w-px bg-border"></div>
                <span class="text-xs font-medium text-muted-foreground uppercase tracking-widest">or</span>
                <div class="flex-1 w-px bg-border"></div>
            </div>

            <div class="flex md:hidden items-center gap-4 my-2">
                <div class="flex-1 h-px bg-border"></div>
                <span class="text-xs font-medium text-muted-foreground uppercase tracking-widest">or</span>
                <div class="flex-1 h-px bg-border"></div>
            </div>

            <Tooltip.Provider>
                <Tooltip.Root disabled={!hasFiles}>
                    <Tooltip.Trigger class="flex">
                        <Card.Root class="flex flex-col flex-1 transition-opacity {hasFiles ? 'opacity-50 pointer-events-none' : ''}">
                            <Card.Header>
                                <Card.Title class="text-base">Share a text</Card.Title>
                                <Card.Description>Write the text you want to share.</Card.Description>
                            </Card.Header>
                            <Card.Content class="flex-1 flex flex-col gap-3">
                                <Label for="note">Your text</Label>
                                <Textarea
                                        id="note"
                                        disabled={hasFiles}
                                        placeholder="Type your text content here..."
                                        class="resize-none flex-1 min-h-30"
                                        bind:value={textValue}
                                />
                                <p class="text-xs text-muted-foreground text-right">{textValue.length} characters</p>
                            </Card.Content>
                        </Card.Root>
                    </Tooltip.Trigger>
                    <Tooltip.Content>
                        <p>Only one share can be made at a time. Please remove the files to share text.</p>
                    </Tooltip.Content>
                </Tooltip.Root>
            </Tooltip.Provider>

        </div>

        <Card.Root class="w-full max-w-4xl">
            <Card.Header>
                <Card.Title class="text-base">Preferences</Card.Title>
                <Card.Description>Configure how your content is shared and handled.</Card.Description>
            </Card.Header>

            <Card.Content class="flex flex-col gap-6">

                <div class="flex flex-col gap-2">
                    <Label>Target Users</Label>
                    <p class="text-xs text-muted-foreground">
                        Search for a user or type an email and press
                        <kbd class="rounded border border-border px-1 py-0.5 font-mono text-[10px]">Enter</kbd> to add.
                        Backspace removes the last one.
                    </p>

                    <!-- Anchor wrapper for the dropdown -->
                    <div class="relative" use:clickOutside={() => { isDropdownOpen = false; }}>

                        <!-- Tag + input row -->
                        <div
                                class="min-h-10 w-full flex flex-wrap gap-1.5 items-center rounded-md border bg-background px-3 py-2 text-sm
                                focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2 transition-shadow
                                {$errors.targetEmails ? 'border-destructive' : 'border-input'}"
                        >
                            {#each $form.targetEmails as email}
                                <Badge variant="secondary" class="flex items-center gap-1 pr-1 text-xs font-normal">
                                    {email}
                                    <button
                                            type="button"
                                            onclick={() => removeEmailTag(email)}
                                            class="ml-0.5 rounded-full hover:bg-muted-foreground/20 p-0.5 transition-colors"
                                            aria-label="Remove {email}"
                                    >
                                        <X class="h-3 w-3" />
                                    </button>
                                </Badge>
                            {/each}

                            <input
                                    class="flex-1 min-w-40 bg-transparent outline-none placeholder:text-muted-foreground text-sm"
                                    placeholder={$form.targetEmails.length === 0 ? "Search users or enter email..." : "Add another..."}
                                    autocomplete="off"
                                    bind:value={emailInput}
                                    onkeydown={handleEmailKeydown}
                                    oninput={handleEmailInput}
                                    onfocus={handleEmailFocus}
                            />

                            {#if isLoadingSuggestions}
                                <Loader2 class="h-4 w-4 animate-spin text-muted-foreground shrink-0" />
                            {/if}
                        </div>

                        {#if isDropdownOpen && (suggestions.length > 0 || (isValidEmail(emailInput) && !$form.targetEmails.includes(emailInput.trim())))}
                            <div class="absolute z-50 mt-1 w-full rounded-md border border-border bg-popover shadow-md text-sm overflow-hidden">

                                {#if suggestions.length > 0}
                                    <ul class="max-h-52 overflow-y-auto py-1" role="listbox">
                                        {#each suggestions as user (user.email)}
                                            {@const alreadyAdded = $form.targetEmails.includes(user.email)}
                                            <li
                                                    role="option"
                                                    aria-selected={alreadyAdded}
                                                    class="flex items-center gap-2.5 px-3 py-2 cursor-pointer transition-colors
                                                    {alreadyAdded
                                                        ? 'opacity-50 pointer-events-none'
                                                        : 'hover:bg-accent hover:text-accent-foreground'}"
                                                    onmousedown={() => {selectSuggestion(user)}}
                                            >
                                                <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-[10px] font-semibold uppercase">
                                                    {(user.email)[0]}
                                                </span>
                                                <div class="flex flex-col min-w-0 flex-1">
                                                    <span class="text-muted-foreground truncate text-xs">{user.email}</span>
                                                </div>
                                                {#if alreadyAdded}
                                                    <Check class="ml-auto h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                                                {/if}
                                            </li>
                                        {/each}
                                    </ul>
                                {/if}

                                {#if isValidEmail(emailInput) && !suggestions.some(s => s.email === emailInput.trim()) && !$form.targetEmails.includes(emailInput.trim())}
                                    {#if suggestions.length > 0}
                                        <div class="border-t border-border"></div>
                                    {/if}
                                    <button
                                            type="button"
                                            class="flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-accent hover:text-accent-foreground transition-colors"
                                            onmousedown={addEmailTag}
                                    >
                                        <Plus class="h-3.5 w-3.5 shrink-0" />
                                        Add <span class="font-medium ml-1">{emailInput.trim()}</span>
                                    </button>
                                {/if}

                            </div>
                        {/if}
                    </div>

                    {#if emailError}
                        <p class="text-xs text-destructive">{emailError}</p>
                    {:else if $errors.targetEmails}
                        <p class="text-xs text-destructive">{$errors.targetEmails}</p>
                    {/if}
                </div>

                <div class="h-px bg-border"></div>

                <div class="flex flex-col gap-4">
                    <p class="text-sm font-medium">Time Controls</p>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">

                        <div class="flex flex-col gap-2">
                            <Label for="active-from" class="flex items-center gap-1.5">
                                <CalendarClock class="h-3.5 w-3.5 text-muted-foreground" />
                                Active From
                            </Label>
                            <Input
                                    id="active-from"
                                    type="datetime-local"
                                    bind:value={$form.activeFrom}
                                    class="{$errors.activeFrom ? 'border-destructive' : ''}"
                            />
                            {#if $errors.activeFrom}
                                <p class="text-xs text-destructive">{$errors.activeFrom}</p>
                            {:else if $form.activeFrom}
                                <p class="text-xs text-muted-foreground">
                                    Will be accessible only from {formatRelative(new Date($form.activeFrom), new Date())}
                                </p>
                            {:else}
                                <p class="text-xs text-muted-foreground">
                                    Will be accessible immediately.
                                </p>
                            {/if}
                        </div>

                        <div class="flex flex-col gap-2">
                            <Label for="expires-at" class="flex items-center gap-1.5">
                                <Clock4 class="h-3.5 w-3.5 text-muted-foreground" />
                                Expires At
                            </Label>
                            <Input
                                    id="expires-at"
                                    type="datetime-local"
                                    bind:value={$form.expiresAt}
                                    min={$form.activeFrom || undefined}
                                    class="{$errors.expiresAt ? 'border-destructive' : ''}"
                            />
                            {#if $errors.expiresAt}
                                <p class="text-xs text-destructive">{$errors.expiresAt}</p>
                            {:else if $form.expiresAt}
                                <p class="text-xs text-muted-foreground">
                                    Ends at {formatRelative(new Date($form.expiresAt), new Date())} — {formatDistance(new Date($form.expiresAt), $form.activeFrom ? new Date($form.activeFrom) : new Date())} after activation
                                </p>
                            {/if}
                        </div>

                    </div>
                </div>

                <div class="h-px bg-border"></div>

                <div class="flex flex-col gap-3">
                    <p class="text-sm font-medium">Notifications</p>
                    <div class="flex items-center gap-3">
                        <Checkbox id="notify-on-open" bind:checked={$form.notifyOnOpen} />
                        <Label for="notify-on-open" class="font-normal cursor-pointer">
                            Email notification when opened
                        </Label>
                    </div>
                    <div class="flex items-center gap-3">
                        <Checkbox id="notify-target-users" bind:checked={$form.notifyTargetUsers} disabled={$form.targetEmails.length === 0} />
                        <Label for="notify-target-users" class="font-normal cursor-pointer">
                            Notify target users by email
                        </Label>
                    </div>
                </div>

                <div class="h-px bg-border"></div>

                <div class="flex flex-col gap-4">
                    <p class="text-sm font-medium">Options</p>
                    <div class="flex items-center justify-between">
                        <div class="flex flex-col gap-0.5">
                            <Label for="auto-save" class="font-normal">Open only once</Label>
                            <p class="text-xs text-muted-foreground">Allow the share to be only viewed/downloaded once. (Refreshing the page will count as a new view)</p>
                        </div>
                        <Switch id="auto-save" bind:checked={$form.allowOnce} />
                    </div>
                    <div class="flex items-center justify-between">
                        <div class="flex flex-col gap-0.5">
                            <Label for="public-listing" class="font-normal">Encrypt content</Label>
                            <p class="text-xs text-muted-foreground">Encrypt the content before uploading to the server</p>
                        </div>
                        <Switch id="public-listing" bind:checked={$form.encrypt} />
                    </div>
                </div>

            </Card.Content>

            <Card.Footer class="flex flex-col gap-3 items-stretch">
                {#if submitError}
                    <p class="text-xs text-destructive text-center">{submitError}</p>
                {/if}
                {#if submitSuccess}
                    <p class="text-xs text-green-600 text-center">Shared successfully!</p>
                {/if}
                <div class="flex justify-end gap-2">
                    <Button
                            variant="outline"
                            type="button"
                            onclick={() => {
                            $form.targetEmails = [];
                            $form.activeFrom = "";
                            $form.expiresAt = "";
                            $form.notifyOnOpen = false;
                            $form.notifyTargetUsers = false;
                            $form.allowOnce = true;
                            $form.encrypt = false;
                            uploadedFiles = [];
                            textValue = "";
                            submitSuccess = false;
                            submitError = "";
                        }}
                    >
                        Reset
                    </Button>
                    <Button type="submit" disabled={submittingToBackend || hasErrors || canSubmit === false}>
                        {submittingToBackend ? "Sharing..." : "Share"}
                    </Button>
                </div>
            </Card.Footer>
        </Card.Root>

    </div>
</form>

<FileUploadDialog
        bind:open={showFileUpload}
        uploads={uploadsData}
        onComplete={handleFileUploadComplete}
/>

<SuccessDialog
        bind:open={showSuccess}
        shareId={currentShareId}
        onCreateNew={createNew}
/>