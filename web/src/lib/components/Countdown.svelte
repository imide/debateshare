<script lang="ts">
	import { Icon } from 'm3-svelte';
	import iconSchedule from '@ktibow/iconset-material-symbols/schedule';
	import { formatCountdown } from '$lib/format';

	let { expiresAt }: { expiresAt: string } = $props();

	let now = $state(Date.now());

	$effect(() => {
		const t = setInterval(() => (now = Date.now()), 1000);
		return () => clearInterval(t);
	});

	const remaining = $derived(new Date(expiresAt).getTime() - now);
	const urgent = $derived(remaining < 10 * 60 * 1000);
</script>

<span class="countdown" class:urgent title="Time until this room expires">
	<Icon icon={iconSchedule} size={18} />
	{formatCountdown(remaining)}
</span>

<style>
	.countdown {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		font-variant-numeric: tabular-nums;
		color: var(--m3c-on-surface-variant);
	}
	.countdown.urgent {
		color: var(--m3c-error);
		font-weight: 600;
	}
</style>
