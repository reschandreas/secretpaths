<script lang="ts">
	/* eslint-disable svelte/prefer-svelte-reactivity */
	/* eslint-disable svelte/infinite-reactive-loop */
	import { Paginator, type PaginationSettings } from '@skeletonlabs/skeleton';
	import SecretCard from './SecretCard.svelte';
	import type { AnnotatedSecret } from '../../types';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';

	let secrets: AnnotatedSecret[] = [];
	let loading = false;
	let error: string | undefined = undefined;
	let input = '';
	let pageParam = 0;
	let limit = 12;
	let hasMore = true;
	let debounceTimer: ReturnType<typeof setTimeout> | undefined;

	let pagination = {
		page: 0,
		limit: 12,
		size: 0,
		amounts: [12, 24, 36, 48]
	} satisfies PaginationSettings;

	async function fetchPage(q: string, p: number, l: number) {
		loading = true;
		error = undefined;
		try {
			const params = new URLSearchParams();
			if (q) params.set('q', q);
			params.set('limit', String(l));
			params.set('offset', String(p * l));
			const res = await fetch(`/v1/annotatedSecrets?${params.toString()}`);
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			const data = await res.json();
			secrets = Array.isArray(data) ? data : [];
			hasMore = secrets.length === l;
			pagination.size = hasMore ? (p + 2) * l : p * l + secrets.length;
		} catch (err) {
			error = String(err);
			secrets = [];
		} finally {
			loading = false;
		}
	}

	function syncUrl(q: string, p: number) {
		if (!browser) return;
		const params = new URLSearchParams();
		if (q) params.set('q', q);
		if (p > 0) params.set('page', String(p));
		const qs = params.toString();
		const url = qs ? `?${qs}` : window.location.pathname;
		history.replaceState({}, '', url);
	}

	function scheduleFetch(q: string, p: number, l: number) {
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(
			() => {
				fetchPage(q, p, l);
				syncUrl(q, p);
			},
			q ? 300 : 0
		);
	}

	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		input = params.get('q') ?? '';
		pageParam = Number(params.get('page') ?? '0');
		pagination.page = pageParam;
		limit = pagination.limit;
		fetchPage(input, pagination.page, limit);
	});

	$: {
		// react to pagination changes
		if (browser && pagination.limit !== limit) {
			limit = pagination.limit;
			fetchPage(input, 0, limit);
			pagination.page = 0;
			syncUrl(input, 0);
		}
	}

	function handleSearch(e: Event) {
		input = (e.target as HTMLInputElement).value;
		pagination.page = 0;
		scheduleFetch(input, 0, limit);
	}

	function handlePage(e: CustomEvent) {
		pagination.page = e.detail?.page ?? pagination.page;
		fetchPage(input, pagination.page, limit);
		syncUrl(input, pagination.page);
	}
</script>

<div class="container h-full w-full mx-auto flex-col justify-center">
	<div class="flex flex-col">
		<h2 class="h2 mt-2 font-thin">search through your paths</h2>
		<input
			class="input autocomplete mt-4 mb-4 border-none min-w-full"
			type="search"
			name="autocomplete-search"
			bind:value={input}
			on:input={handleSearch}
			placeholder="Search..."
			aria-label="Search paths"
		/>
	</div>
	<div>
		{#if loading}
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each [0, 1, 2, 3, 4, 5] as i (i)}
					<div class="card p-4 animate-pulse">
						<div class="h-4 bg-surface-300 rounded w-3/4 mb-2"></div>
						<div class="h-3 bg-surface-300 rounded w-1/2"></div>
					</div>
				{/each}
			</div>
		{:else if error}
			<div class="card p-8 text-center text-error-500">
				<p>Failed to load paths: {error}</p>
				<button
					class="btn variant-filled mt-4"
					on:click={() => fetchPage(input, pagination.page, limit)}>Retry</button
				>
			</div>
		{:else if secrets.length === 0}
			<div class="card p-8 text-center opacity-70">
				<p>No paths match your search.</p>
			</div>
		{:else}
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each secrets as path (path.path?.path ?? path)}
					<SecretCard secret={path} />
				{/each}
			</div>
		{/if}
		<div class="mt-4">
			<Paginator
				bind:settings={pagination}
				showFirstLastButtons={true}
				showPreviousNextButtons={true}
				on:page={handlePage}
			/>
		</div>
	</div>
</div>
EOF echo done
