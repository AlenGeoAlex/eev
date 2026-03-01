<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
    import { ModeWatcher } from "mode-watcher";

	let { children } = $props();
</script>

<ModeWatcher />
{@render children()}
