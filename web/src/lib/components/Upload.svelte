<script lang="ts">
	import { Button, LinearProgress, snackbar } from 'm3-svelte';
	import { uploadFile } from '$lib/api';

	const MAX_SIZE = 10 * 1024 * 1024; // 10mb

	let { code, onUploaded }: { code: string; onUploaded: () => void } = $props();

	let dragging = $state(false);
	let uploads = $state<{ name: string; percent: number }[]>([]);
	let input: HTMLInputElement;

	async function send(files: FileList | globalThis.File[] | null) {
		if (!files) return;
		for (const file of Array.from(files)) {
			if (file.size > MAX_SIZE) {
				snackbar(`${file.name} is over the 10 MB limit`);
				continue;
			}
			const entry = $state({ name: file.name, percent: 0 });
			uploads.push(entry);
			try {
				await uploadFile(code, file, (p) => (entry.percent = p));
				onUploaded();
			} catch (e) {
				snackbar(e instanceof Error ? e.message : `Upload of ${file.name} failed`);
			} finally {
				uploads = uploads.filter((u) => u !== entry);
			}
		}
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		send(e.dataTransfer?.files ?? null);
	}
</script>

<div
	class="zone"
	class:dragging
	role="button"
	tabindex="0"
	aria-label="Upload files"
	ondragover={(e) => {
		e.preventDefault();
		dragging = true;
	}}
	ondragleave={() => (dragging = false)}
	ondrop={onDrop}
	onclick={() => input.click()}
	onkeydown={(e) => e.key === 'Enter' && input.click()}
>
	<input
		bind:this={input}
		type="file"
		multiple
		hidden
		onchange={(e) => {
			send(e.currentTarget.files);
			e.currentTarget.value = '';
		}}
	/>
	<p>Drag &amp; drop files here</p>
	<Button
		variant="filled"
		onclick={(e: MouseEvent) => {
			e.stopPropagation();
			input.click();
		}}
	>
		Choose files
	</Button>
	<p class="limit">Up to 10 MB per file</p>
</div>

{#each uploads as u (u)}
	<div class="progress">
		<span>{u.name}</span>
		<LinearProgress percent={u.percent} aria-label="pload progress for {u.name}" />
	</div>
{/each}

<style>
	.zone {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
		padding: 2rem 1rem;
		border: 2px dashed var(--m3c-outline-variant);
		border-radius: 1.5rem;
		cursor: pointer;
		transition:
			border-color 0.15s,
			background-color 0.15s;
	}
	.zone.dragging {
		border-color: var(--m3c-primary);
		background-color: var(--m3c-primary-container);
	}
	.zone p {
		margin: 0;
		color: var(--m3c-on-surface-variant);
	}
	.zone .limit {
		font-size: 0.75rem;
	}
	.progress {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		margin-top: 0.75rem;
		font-size: 0.85rem;
	}
</style>
