<script>
	import BubbleChart from './BubbleChart.svelte';

	/** @type {import('./$types').PageData} */
	export let data;

	// graph tests

</script>
<div class="container h-full mx-auto flex">
	<div class="flex w-1/4 flex-col">
		<h2 class="h2">Welcome to Secretpaths.</h2>
		{#if data}
			{#each data.policies as policy}
				<div class="card p-2 mt-2">
					<header class="card-header">{policy.name}</header>
					<section class="p-2">{policy.rules.length} Policies</section>
				</div>
			{/each}
		{/if}
	</div>
	<div class="w-full ml-4">
		<h2 class="h2">First steps with graphs</h2>
		{#if data.policies}
			<BubbleChart policies={data.policies}/>
		{/if}
	</div>
</div>
