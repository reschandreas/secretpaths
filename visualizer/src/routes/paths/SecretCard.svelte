<script lang="ts">
	import type { AnnotatedSecret } from '../../types';

	export let secret: AnnotatedSecret;

	let splitPath = secret.path.path.split('/');
	let name = splitPath.pop();
	let parts = splitPath.length;
</script>

<div class="card card-hover">
	<header class="card-header mt-2 h4">{name}</header>
	<section class="mt-2 ml-4 mr-4 min-h-20 max-h-20">
		<ol class="breadcrumb mt-1 flex-wrap">
			{#each splitPath as crumb, index}
				{#if crumb === ''}
					<li class="crumb">/</li>
				{:else}
					<li class="crumb">{crumb}</li>
				{/if}
				{#if index < parts - 1}
					<li class="crumb-separator" aria-hidden="true">&#x290D;</li>
				{/if}
			{/each}
		</ol>
	</section>
	<hr class="opacity-50" />
	<footer class="card-footer mt-2">{secret.policies?.length} Policies have access</footer>
</div>
