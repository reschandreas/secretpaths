<script lang="ts">

	let input = '';

	let policies: string[] | undefined = undefined;

	$: if (input) {
		fetch(`/v1/annotated?path=${input.startsWith('/') ? input : '/' + input}`)
			.then((res) => res.json())
			.then((data) => {
				policies = data;
				console.log(data);
			})
			.catch((err) => {
				console.error(err);
			});
	}
</script>

<div class="container p-10 h-full mx-auto flex-row">
	<h2 class="h2 font-thin">welcome to secretpaths.</h2>
	<div class="pt-10 flex flex-col">
		<div class="w-full text-token grid grid-cols-1 md:grid-cols-1 gap-4">
			<div class="card card-hover overflow-hidden p-4">
				<h6 class="h3 mb-5 font-thin" data-toc-ignore>Quick lookup</h6>
				<hr class="opacity-50" />
				<input class="input" type="search" bind:value={input} placeholder="type your secretpath here...">
				<hr class="opacity-50" />
				<div class="mt-2 ml-2">
					{#if policies === undefined}
						<p>Lookup your secretpath to see which policies can access it.</p>
					{:else if policies.length === 0}
						<p>Looks like no policies have access to this secret.</p>
					{:else}
						<p>The following policies can access this secret:</p>
						<ul>
							{#each policies as policy}
								<li>{policy}</li>
							{/each}
						</ul>
					{/if}
				</div>
			</div>
		</div>
		<div class="w-full text-token grid grid-cols-1 md:grid-cols-2 gap-4 mt-20">
			<div class="card card-hover overflow-hidden">
				<div class="p-4 space-y-4">
					<h6 class="h3 mb-10 font-thin" data-toc-ignore>Inspect policies</h6>
					Quickly inspect your paths and see which policies can access them. No more guessing, just inspect.
				</div>
				<hr class="opacity-50" />
				<footer class="p-4 flex justify-start items-center space-x-4">
					<a href="/paths">
						<button type="button" class="btn variant-filled">Go</button>
					</a>
				</footer>
			</div>
			<div class="card card-hover overflow-hidden">
				<div class="p-4 space-y-4">
					<h6 class="h3 mb-10 font-thin" data-toc-ignore>Discover paths</h6>
					Simply look for your secrets without any hassle. No clicking, no typing, just look for your
					secrets.
				</div>
				<hr class="opacity-50" />
				<footer class="p-4 flex justify-start items-center space-x-4">
					<a href="/discover">
						<button type="button" class="btn variant-filled">Go</button>
					</a>
				</footer>
			</div>
		</div>
	</div>
</div>
