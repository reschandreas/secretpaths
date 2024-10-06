<script lang="ts">
	import { Paginator, type PaginationSettings } from '@skeletonlabs/skeleton';
	import SecretCard from './SecretCard.svelte';
	import type { AnnotatedSecret } from '../../types';

	/** @type {import('../$types').PageData} */
	export let data;

	let secrets: AnnotatedSecret[] = data.annotatedSecrets;

	let pagination = {
		page: 0,
		limit: 12,
		size: secrets.length,
		amounts: [12, 24, 36, 48]
	} satisfies PaginationSettings;

	let input = '';

	$: paginatedSource = secrets
		.filter((p) => p.path.path.includes(input))
		.slice(
			pagination.page * pagination.limit,
			pagination.page * pagination.limit + pagination.limit
		);
</script>

<div class="container h-full w-full mx-auto flex-col justify-center">
	<div class="flex flex-col">
		<h2 class="h2 mt-2 font-thin">search through your paths</h2>
		<input
			class="input autocomplete mt-4 mb-4 border-none min-w-full"
			type="search"
			name="autocomplete-search"
			bind:value={input}
			placeholder="Search..."
		/>
	</div>
	<div>
		{#key paginatedSource}
			<div class="grid grid-cols-3 grid-rows-4 grid-flow-col gap-4">
				{#each paginatedSource as path}
					<SecretCard secret={path} />
				{/each}
			</div>
		{/key}
		<div class="mt-4">
			<Paginator
				bind:settings={pagination}
				showFirstLastButtons={true}
				showPreviousNextButtons={true}
			/>
		</div>
	</div>
</div>
