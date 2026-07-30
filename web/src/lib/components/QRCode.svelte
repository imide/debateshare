<script lang="ts">
	import { qrcode } from 'etiket';

	let { url }: { url: string } = $props();

	let svg = $derived(
		qrcode(url, {
			shape: 'square',
			dotType: 'square',

			corners: {
				topLeft: { outerShape: 'rounded', innerShape: 'rounded' },

				topRight: { outerShape: 'rounded', innerShape: 'rounded' },

				bottomLeft: { outerShape: 'rounded', innerShape: 'rounded' }
			}
		})
	);
</script>

<article>
	{svg}
</article>

<style>
	article :global {
		width: min(18rem, 100%);
		border-radius: 0.8rem;
		background: white;
		padding: 0.6rem;
	}
</style>
