<script lang="ts">
	import * as Dialog from "$lib/components/ui/dialog";
	import { Button } from "$lib/components/ui/button";
	import { CircleCheck, Copy, Plus, Check } from "@lucide/svelte";
	import { cn } from "$lib/utils";
	import {goto} from "$app/navigation";

	let {
		open = $bindable(false),
		shareId = "",
		onCreateNew
	}: {
		open: boolean;
		shareId: string;
		onCreateNew: () => void;
	} = $props();

	let copied = $state(false);
	const shareUrl = $derived(`${window.location.origin}/share/${shareId}`);

	function copyToClipboard() {
		navigator.clipboard.writeText(shareUrl);
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-md">
		<div class="flex flex-col items-center gap-4 py-6 text-center">
			<div
				class="bg-green-100 dark:bg-green-900/30 flex h-16 w-16 items-center justify-center rounded-full"
			>
				<CircleCheck class="text-green-600 dark:text-green-400 h-10 w-10" />
			</div>

			<div class="space-y-2">
				<Dialog.Title class="text-2xl font-bold text-green-700 dark:text-green-400">
					Share is Ready!
				</Dialog.Title>
				<Dialog.Description class="text-base text-muted-foreground">
					Your content has been successfully published. Ask your peer to go to {window.location.origin} and enter
				</Dialog.Description>
			</div>

			<div class="mt-4 flex w-56 items-center gap-2 rounded-xl border bg-muted/30 p-3">
				<div class="flex-1 overflow-hidden">
					<p class="truncate text-center text-sm font-mono text-muted-foreground">
						{shareId}
					</p>
				</div>
				<Button
					size="icon"
					variant="ghost"
					class="shrink-0 hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/20"
					onclick={copyToClipboard}
				>
					{#if copied}
						<Check class="size-4 text-green-600" />
					{:else}
						<Copy class="size-4" />
					{/if}
					<span class="sr-only">Copy link</span>
				</Button>
			</div>

			<div class="mt-4 flex w-full flex-col gap-2 sm:flex-row">
				<Button
					variant="outline"
					class="flex-1"
					onclick={() => {
						open = false;
					}}
				>
					Close
				</Button>
				<Button
					class="flex-1 bg-green-600 hover:bg-green-700 dark:bg-green-600 dark:text-white"
					onclick={() => {
						open = false;
						goto('/share');
					}}
				>
					<Plus class="mr-2 size-4" />
					Create New
				</Button>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>
