<script lang="ts">
	import { Button, Icon, ListItem, snackbar } from 'm3-svelte';
	import iconDescription from '@ktibow/iconset-material-symbols/description';
	import iconDownload from '@ktibow/iconset-material-symbols/download';
	import iconEdit from '@ktibow/iconset-material-symbols/edit';
	import iconPreview from '@ktibow/iconset-material-symbols/preview';
	import iconDelete from '@ktibow/iconset-material-symbols/delete';
	import type { FileInfo } from '$lib/types';
	import { getDownloadUrl, uploadFile } from '$lib/api';
	import { relativeTime } from '$lib/format';
	import { filesize } from 'filesize';

	let {
		code,
		file,
		onChanged,
		onDelete
	}: {
		code: string;
		file: FileInfo;
		onChanged: () => void;
		onDelete: (file: FileInfo) => void;
	} = $props();

	let replaceInput: HTMLInputElement;
	let replacing = $state(false);

	async function replace(files: FileList | null) {
		const chosen = files?.[0];
		if (!chosen) return;
		if (chosen.size > 10 * 1024 * 1024) {
			snackbar(`${chosen.name} is over the 10mb limit.`);
			return;
		}
		replacing = true;
		try {
			await uploadFile(code, chosen, () => {}, file.id);
			onChanged();
		} catch (e) {
			snackbar(e instanceof Error ? e.message : 'Replace failed');
		} finally {
			replacing = false;
		}
	}
</script>

<ListItem
	headline={file.name}
	supporting="{filesize(file.size)} | updated {relativeTime(file.updatedAt)}"
>
	{#snippet leading()}
		<Icon icon={iconDescription} size={24} />
	{/snippet}
	{#snippet trailing()}
		<div class="actions">
			<Button
				variant="text"
				iconType="full"
				href={`https://docs.google.com/gview?url=${await getDownloadUrl(code, file.id)}`}
				/* is there a better way to do this? */
				target="_blank"
				title="Preview"
			>
				<Icon icon={iconPreview} />
			</Button>
			<Button
				variant="text"
				iconType="full"
				href={await getDownloadUrl(code, file.id)}
				target="_blank"
				title="Download"
			>
				<Icon icon={iconDownload} />
			</Button>
			<Button
				variant="text"
				iconType="full"
				onclick={() => replaceInput.click()}
				disabled={replacing}
				title="Replace with a new version"
			>
				<Icon icon={iconEdit} />
			</Button>
			<Button variant="text" iconType="full" onclick={() => onDelete(file)} title="Delete">
				<Icon icon={iconDelete} />
			</Button>
		</div>
	{/snippet}
</ListItem>

<input
	bind:this={replaceInput}
	type="file"
	hidden
	onchange={(e) => {
		replace(e.currentTarget.files);
		e.currentTarget.value = '';
	}}
/>

<style>
	.actions {
		display: flex;
		gap: 0.25rem;
	}
</style>
