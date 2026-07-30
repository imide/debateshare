<script lang="ts">
	import { ApiError, createRoom, joinRoom } from '$lib/api';
	import { pushState } from '$app/navigation';
	import { Card, snackbar, TextFieldOutlined, Button } from 'm3-svelte';
	import { resolve } from '$app/paths';

	let roomName = $state('');
	let joinCode = $state('');
	let busy = $state(false);

	const unused: string = 'undefined'; // for features that are not supported yet

	async function create() {
		busy = true;
		try {
			// roomname can be null but for now email is not supported.
			const room = await createRoom(roomName.trim(), unused);
			pushState(
				resolve('/room/[slug]', {
					slug: room.code
				}),
				{}
			);
		} catch (e) {
			snackbar(e instanceof ApiError ? e.message : 'Could not create room');
		} finally {
			busy = false;
		}
	}

	async function join() {
		const code = joinCode.trim().toLowerCase();
		if (!code) {
			snackbar('Enter a room code');
			return;
		}
		busy = true;
		try {
			const room = await joinRoom(code, unused);
			pushState(
				resolve('/room/[slug]', {
					slug: room.code
				}),
				{}
			);
		} catch (e) {
			snackbar(e instanceof ApiError ? e.message : 'Could not join room');
		} finally {
			busy = false;
		}
	}
</script>

<div class="hero">
	<h1>DebateShare</h1>
	<p class="summary">
		a modern alternative to the amazing <a href="https://speechdrop.net" target="_blank"
			>speechdrop</a
		> with a few quality of life additions.
		rooms last 24 hours. email support is coming soon.
	</p>
</div>

<div class="panels">
	<Card variant="filled">
		<div class="panel">
			<h2>create a room</h2>
			<p class="hint">create once, then share the <strong>two word room code</strong> to others or the qr code.</p>
			<TextFieldOutlined label="room name (optional)" bind:value={roomName} enter={create} />
			<TextFieldOutlined label="email (optional) - coming soon" disabled={true} />
			<Button variant="filled" onclick={create} disabled={busy}>create room</Button>
		</div>
	</Card>

	<Card variant="outlined">
		<div class="panel">
			<h2>join an ongoing room</h2>
			<p class="hint">enter the <strong>two word room code</strong> to join.</p>
			<l></l>
			<TextFieldOutlined
				label="room code"
				bind:value={joinCode}
				enter={join}
				oninput={() => (joinCode = joinCode.toLowerCase())}
			/>
			<TextFieldOutlined label="email (optional) - coming soon" disabled={true} />
			<Button variant="tonal" onclick={join} disabled={busy}>Join room</Button>
		</div>
	</Card>
</div>

<style>
	.hero {
		text-align: center;
		max-width: 34rem;
		margin: 2rem 0 2.5rem;
	}
	.hero h1 {
		font-size: 3rem;
		font-weight: 650;
		margin: 0 0 0.5rem;
		color: var(--m3c-primary);
	}
	.hero p {
		color: var(--m3c-on-surface-variant);
		margin: 0;
	}

	.summary {
		margin: 0.9rem 0 0;
		color: var(--m3c-on-surface-variant);
		max-width: 56ch;
	}
	.panels {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr));
		gap: 1.25rem;
		width: 100%;
		max-width: 44rem;
	}
	.panel {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		padding: 0.5rem;
	}
	.panel h2 {
		margin: 0;
		font-size: 1.3rem;
		font-weight: 550;
	}
	.hint {
		margin: 0;
		font-size: 0.8rem;
		color: var(--m3c-on-surface-variant);
	}
</style>
