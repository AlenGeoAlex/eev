<script lang="ts">
	import * as Dialog from "$lib/components/ui/dialog";
	import { Progress } from "$lib/components/ui/progress";
	import { ShareableService } from "$lib/services/shareable.service";
	import { LoaderCircle } from "@lucide/svelte";

	let {
		open = $bindable(false),
		uploads = [],
		onComplete
	}: {
		open: boolean;
		uploads: {
			fileId: string;
			url: string;
			name: string;
			file: File;
		}[];
		onComplete: () => void;
	} = $props();

	let progressMap = $state<{ [key: string]: number }>({});
	let completedCount = $state(0);
	let errorCount = $state(0);
	let totalCount = $derived(uploads.length);

	let started = $state(false);

	async function startUploads() {
		if (started || uploads.length === 0) return;
		started = true;
		completedCount = 0;
		errorCount = 0;

		const uploadPromises = uploads.map(async (upload) => {
			progressMap[upload.fileId] = 0;
			try {
				await ShareableService.getInstance().uploadWithPreSignedUrl(upload.url, upload.file, {
					onProgress: (percent) => {
						progressMap[upload.fileId] = percent;
					},
					onError: (error) => {
						console.error(`Failed to upload ${upload.name}`, error);
						errorCount++;
					},
					onSuccess: () => {
						completedCount++;
					},
				});
			} catch (e) {
				console.error(`Failed to upload ${upload.name}`, e);
				errorCount++;
			}
		});

		await Promise.all(uploadPromises);
		if (completedCount + errorCount === totalCount) {
			onComplete();
		}
	}

	$effect(() => {
		if (open && uploads.length > 0 && !started) {
			startUploads();
		}
	});
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-md" showCloseButton={false}>
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<LoaderCircle class="h-5 w-5 animate-spin" />
				Uploading Files
			</Dialog.Title>
			<Dialog.Description>
				Please wait while we upload your files. Do not close this window.
			</Dialog.Description>
		</Dialog.Header>

		<div class="flex flex-col gap-4 py-4 max-h-[40vh] overflow-y-auto pr-2">
			{#each uploads as upload}
				<div class="space-y-1.5">
					<div class="flex justify-between text-xs font-medium">
						<span class="truncate max-w-62.5">{upload.name}</span>
						<span class="text-muted-foreground">{progressMap[upload.fileId] || 0}%</span>
					</div>
					<Progress value={progressMap[upload.fileId] || 0} class="h-1.5" />
				</div>
			{/each}
		</div>

		<div class="flex flex-col items-center gap-1 text-sm">
			<p class="font-medium">
				{completedCount} of {totalCount} files uploaded
			</p>
			{#if errorCount > 0}
				<p class="text-destructive text-xs">{errorCount} uploads failed</p>
			{/if}
		</div>
	</Dialog.Content>
</Dialog.Root>
