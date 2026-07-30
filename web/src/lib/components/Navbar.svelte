<script lang="ts">
	import { resolve } from '$app/paths';
	import iconHome from '@ktibow/iconset-material-symbols/home';
	import iconInfo from '@ktibow/iconset-material-symbols/info';
	import iconLightMode from '@ktibow/iconset-material-symbols/light-mode';
	import iconDarkMode from '@ktibow/iconset-material-symbols/dark-mode';
	import GithubIcon from '@iconify-svelte/mdi/github';
	import { page } from '$app/state';
	import { Card, Button, Icon } from 'm3-svelte';
	import { mode, toggleMode } from 'mode-watcher';
	import { goto } from '$app/navigation';

	// let currentPath: string = page.url.pathname;

	const paths = [
		{
			path: resolve('/'),
			icon: iconHome,
			label: 'home'
		},
		{
			path: resolve('/about'),
			icon: iconInfo,
			label: 'about'
		}
	];

	// const normalizePath = (path: string) => {
	// 	const u = new URL(path, page.url.href);
	// 	path = u.pathname;
	// 	if (path.endsWith('/')) path = path.slice(0, -1);
	// 	return path || '/';
	// };

  function isActivePath(path: string) {
    if (path === "/") return page.url.pathname === "/";
    return page.url.pathname.startsWith(path);
  }
</script>

<header class="navbar-wrap">
	<Card variant="filled" class="navbar-card">
		<div class="navbar-row">
			<div class="left-group">
				<a
					class="brand"
					href={resolve('/')}
					onclick={(event) => {
						event.preventDefault();
						goto(resolve('/'));
					}}
					aria-label="home"
				>
					<span>debate</span>
					<span class="brand-accent">share</span>
				</a>
				<nav class="navlist">
					{#each paths as path (path.label)}
						<Button
							variant={isActivePath(path.path) ? 'filled' : 'tonal'}
							onclick={() => goto(path.path)}
						>
							<Icon icon={path.icon} />
							{path.label}
						</Button>
					{/each}
				</nav>
			</div>

			<div class="right-group">
				<a
					class="icon-link"
					href="https://github.com/imide/debateshare"
					target="_blank"
					rel="noreferrer"
					aria-label="open github"
				>
					<GithubIcon width=1.5rem height=1.5rem />
				</a>
				<Button
					variant="tonal"
					square
					onclick={toggleMode}
					aria-label={mode.current === 'light' ? 'enable dark mode' : 'enable light mode'}
				>
					<Icon icon={mode.current === 'light' ? iconDarkMode : iconLightMode} />
				</Button>
			</div>
		</div>
	</Card>
</header>

<style>
	.navbar-wrap {
		position: sticky;
		top: 1rem;
		z-index: 100;
		display: flex;
		justify-content: center;
		padding: 0 1rem;
	}
	:global(.navbar-card) {
		width: min(100%, 56rem);
		border-radius: 1rem;
		background: color-mix(in oklab, var(--m3c-surface-container) 78%, transparent);
		backdrop-filter: blur(16px);
	}
	.navbar-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		padding: 0.75rem;
		flex-wrap: nowrap;
	}
	.left-group,
	.right-group {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-wrap: nowrap;
	}
	.brand {
		display: inline-flex;
		align-items: center;
		gap: 0.1rem;
		color: var(--m3c-on-surface);
		text-decoration: none;
		font-size: 1.15rem;
		font-weight: 700;
		letter-spacing: 0.02em;
	}
	.brand-accent {
		color: var(--m3c-primary);
	}
	.navlist {
		display: flex;
		gap: 0.5rem;
		align-items: center;
		flex-wrap: nowrap;
	}
	.navlist :global(.button),
	.right-group :global(.button) {
		--density: -2;
	}
	.icon-link {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 999px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		color: var(--m3c-on-surface);
		text-decoration: none;
		background: transparent;
	}
	.icon-link:hover {
		background: color-mix(in oklab, var(--m3c-secondary-container) 45%, transparent);
	}
	@media (min-width: 48rem) {
		.navbar-wrap {
			top: 2rem;
		}
	}
</style>
