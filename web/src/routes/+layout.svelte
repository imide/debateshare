<script lang="ts">
	import googleSans from '@fontsource-variable/google-sans-flex/files/google-sans-flex-latin-opsz-normal.woff2';
	import favicon from '$lib/assets/favicon.svg';
	import '../app.css';
	import Navbar from '$lib/components/Navbar.svelte';
	import { ModeWatcher } from 'mode-watcher';
	import { Snackbar } from 'm3-svelte';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<link rel="preload" href={googleSans} as="font" type="font/woff2" crossorigin="" />
</svelte:head>

<ModeWatcher defaultMode="dark" />

<div class="app">
	<Navbar />
	<main>
		{@render children()}
	</main>
</div>

<Snackbar />

<style>
	.app {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
	}
	main {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 1.5rem 1rem 3rem;
	}
</style>
