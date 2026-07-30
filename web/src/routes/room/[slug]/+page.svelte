<script lang="ts">
	import { Button, Card, Chip, Dialog, Icon, snackbar } from 'm3-svelte';
	import iconCopy from '@ktibow/iconset-material-symbols/content-copy';
	import iconShare from '@ktibow/iconset-material-symbols/share';
	import iconDownload from '@ktibow/iconset-material-symbols/download';
	import { RoomEvents } from '$lib/sse.svelte';
	import Countdown from '$lib/components/Countdown.svelte';
	import Upload from '$lib/components/Upload.svelte';
	import FileRow from '$lib/components/FileRow.svelte';
	import type { FileInfo, Room } from '$lib/types';
	import { getRoom, deleteFile, listFiles, ApiError, getRoomZipUrl } from '$lib/api';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import QRCode from '$lib/components/QRCode.svelte';
	import { page } from '$app/state';

	let { params } = $props();

	const code = $derived(params.slug.toLowerCase());
	const roomUrl = $derived(page.url.toString());

	let room = $state<Room | null>(null);
	let files = $state<FileInfo[]>([]);
	let loadError = $state('');
	let expired = $state(false);
	let pendingDelete = $state<FileInfo | null>(null);
	let deleteOpen = $state(false);
	let qrOpen = $state(false);

	let events = $state<RoomEvents | null>(null);

	$effect(() => {
		if (!code) return;
		load();
		events = new RoomEvents(code, refetch, () => (expired = true));
		return () => events?.close();
	});

	// TODO: Use sveltekit's error and have a custom error page there?
	async function load() {
		try {
			const r = await getRoom(code);
			room = r;
			files = r.files ?? [];
			loadError = '';
			expired = false;
		} catch (e) {
			if (e instanceof ApiError && e.status === 404) {
				events?.expire();
				expired = true;
			} else {
				loadError = e instanceof ApiError ? e.message : 'could not load room';
			}
		}
	}

	async function refetch() {
		try {
			files = await listFiles(code);
		} catch (e) {
			if (e instanceof ApiError && e.status === 404) events?.expire();
		}
	}

	function copyCode() {
		navigator.clipboard.writeText(code).then(
			() => snackbar('room code copied'),
			() => snackbar('could not copy code')
		);
	}

	function copyRoomLink() {
		navigator.clipboard.writeText(roomUrl).then(
			() => snackbar('room link copied'),
			() => snackbar('could not copy room link')
		);
	}

	function share() {
		if (navigator.share) {
			navigator.share({ title: `debateshare room ${code}`, url: roomUrl }).catch(() => {});
		} else {
			navigator.clipboard.writeText(roomUrl).then(() => snackbar('room link copied'));
		}
	}

	function askDelete(file: FileInfo) {
		pendingDelete = file;
		deleteOpen = true;
	}

	async function confirmDelete() {
		if (!pendingDelete) return;
		try {
			await deleteFile(code, pendingDelete.id);
			refetch();
		} catch (e) {
			snackbar(e instanceof ApiError ? e.message : 'deletion failed');
		} finally {
			deleteOpen = false;
			pendingDelete = null;
		}
	}
</script>

{#if loadError}
	<div class="state">
		<h1>Something went wrong</h1>
		<p>{loadError}</p>
		<Button variant="filled" onclick={() => location.reload()}>Retry</Button>
	</div>
{:else if room}
	<section class="room">
		<header>
			<div class="title">
				<h1>{room.name}</h1>
				<Countdown expiresAt={room.expiresAt} />
			</div>
			<div class="meta">
				<Chip variant="assist" icon={iconCopy} onclick={copyCode}>{code}</Chip>
				<Chip variant="assist" icon={iconShare} onclick={share}>Share</Chip>
				{#if files.length > 0}
					<Button
						variant="outlined"
						iconType="left"
						href={await getRoomZipUrl(code)}
						target="_blank"
						rel="noreferrer"
					>
						<Icon icon={iconDownload} />
						download archive
					</Button>
				{:else}
					<Button
						variant="outlined"
						iconType="left"
						onclick={() => snackbar('no files uploaded yet! why do you need an empty zip? :)')}
					>
						<Icon icon={iconDownload} />
						download archive
					</Button>
				{/if}
				{#if events && !events.connected && !expired}
					<span class="offline">reconnecting…</span>
				{/if}
			</div>
		</header>

		<Upload {code} onUploaded={refetch} />

		<Card variant="filled">
			<div class="files">
				{#if files.length === 0}
					<p class="empty">no files yet | drop the first one above</p>
				{:else}
					{#each files as file (file.id)}
						<FileRow {code} {file} onChanged={refetch} onDelete={askDelete} />
					{/each}
				{/if}
			</div>
		</Card>
	</section>
{:else if !expired}
	<div class="state"><p>loading room…</p></div>
{/if}

<Dialog headline="delete file?" bind:open={deleteOpen}>
	<p>“{pendingDelete?.name}” will be removed for everyone in the room</p>
	{#snippet buttons()}
		<Button variant="text" onclick={() => (deleteOpen = false)}>cancel</Button>
		<Button variant="filled" onclick={confirmDelete}>delete</Button>
	{/snippet}
</Dialog>

<Dialog headline="room expired" open={expired}>
	<p>
		the room you have tried to enter <strong>({room?.code})</strong> has expired.
	</p>
	{#snippet buttons()}
		<Button variant="filled" onclick={() => goto(resolve('/'))}>back to home</Button>
	{/snippet}
</Dialog>

<Dialog headline="show room qrcode" bind:open={qrOpen}>
	<div class="qr-wrap">
		<QRCode url={roomUrl} />
	</div>
	{#snippet buttons()}
		<Button variant="text" onclick={copyRoomLink}>copy link</Button>
		<Button variant="filled" onclick={share}>share</Button>
	{/snippet}
</Dialog>

<style>
	.room {
		width: min(100%, 56rem);
		display: grid;
		gap: 1rem;
	}
	header {
		display: grid;
		gap: 0.75rem;
	}
	.title {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 0.75rem;
		flex-wrap: wrap;
	}
	.title h1 {
		margin: 0;
		font-size: clamp(1.4rem, 2vw, 1.95rem);
	}
	.meta {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-wrap: wrap;
	}
	.offline {
		color: var(--m3c-error);
		font-size: 0.8rem;
	}
	.files {
		padding: 0.25rem;
	}
	.empty {
		margin: 0;
		padding: 1rem;
		text-align: center;
		color: var(--m3c-on-surface-variant);
	}
	.state {
		text-align: center;
		margin-top: 4rem;
	}
</style>
