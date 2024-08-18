<script lang="ts">

	import { ProgressBar } from '@skeletonlabs/skeleton';
	import type { Information } from '../../types';

	export let secret: string;
	export let information: Information;
	// $: policies = fetchPolicies();
	let policies: string[] | undefined = undefined;

	$: if (secret) {
		loadPolicies();
	}

	async function loadPolicies() {
		try {
			const res = await fetch(`/analyzed?path=${secret}`);
			policies = await res.json();
		} catch (err) {
			console.error(err);
		}
	}

	$: splitPath = secret.split('/');
	$: parts = splitPath.length;
</script>
<div class="card card-hover overflow-hidden z-10 w-96">
	<div class="p-4 space-y-4">
		<section class="mt-2 ml-4 mr-4 min-h-20 max-h-20">
			<ol class="breadcrumb mt-1 flex-wrap">
				{#each splitPath as crumb, index }
					{#if crumb === ''}
						<li class="crumb">/</li>
					{:else}
						<li class="crumb">{crumb}</li>
					{/if}
					{#if index < parts - 1 }
						<li class="crumb-separator" aria-hidden="true">&#x290D;</li>
					{/if}
				{/each}
			</ol>
			<button>close</button>
		</section>
		<hr class="opacity-50" />
		<!--		<h6 class="h4 mb-10 font-thin" data-toc-ignore>{secret}</h6>-->
		{#if policies === undefined }
			<ProgressBar class="w-1/2 content-center" />
		{:else }
			{#if policies.length === 0}
				<p>Looks like no policies have access to this secret.</p>
			{:else}
				<p>The following policies can access this secret:</p>
				<ul>
					{#each policies as policy}
						<li>{policy}</li>
					{/each}
				</ul>
			{/if}
		{/if}
	</div>
	<hr class="opacity-50" />
	<footer class="p-4 flex justify-start items-center space-x-4">
		<a href="{information.vaultAddress}/ui/vault/secrets/{information.kvEngine}/show{secret}" target="_blank">
			<button type="button" class="btn variant-filled">Bring me there!</button>
		</a>
	</footer>
</div>