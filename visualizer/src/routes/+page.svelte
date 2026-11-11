<script lang="ts">
	let input = '';
	let policies: string[] | undefined = undefined;
	let loading = false;
	let error: string | undefined = undefined;
	let debounceTimer: ReturnType<typeof setTimeout> | undefined;

	function lookup(value: string) {
		if (!value) {
			policies = undefined;
			error = undefined;
			return;
		}
		loading = true;
		error = undefined;
		const path = value.startsWith('/') ? value : '/' + value;
		fetch(`/v1/annotated?path=${encodeURIComponent(path)}`)
			.then((res) => {
				if (!res.ok) throw new Error(`HTTP ${res.status}`);
				return res.json();
			})
			.then((data) => {
				policies = Array.isArray(data) ? data : (data?.policies ?? []);
			})
			.catch((err) => {
				error = String(err);
				policies = undefined;
			})
			.finally(() => {
				loading = false;
			});
	}

	$: {
		clearTimeout(debounceTimer);
		const v = input;
		// eslint-disable-next-line no-useless-assignment
		debounceTimer = setTimeout(() => lookup(v), 300);
	}

	function clearInput() {
		input = '';
		policies = undefined;
		error = undefined;
	}
</script>

<div class="container p-10 h-full mx-auto flex-row">
	<h2 class="h2 font-thin">welcome to secretpaths.</h2>
	<div class="pt-10 flex flex-col">
		<div class="w-full text-token grid grid-cols-1 md:grid-cols-1 gap-4">
			<div class="card card-hover overflow-hidden p-4">
				<h6 class="h3 mb-5 font-thin" data-toc-ignore>Quick lookup</h6>
				<hr class="opacity-50" />
				<div class="relative">
					<input
						class="input pr-10"
						type="search"
						bind:value={input}
						placeholder="type your secretpath here..."
						aria-label="Secret path lookup"
					/>
					{#if input}
						<button
							type="button"
							class="absolute right-2 top-1/2 -translate-y-1/2 btn btn-sm variant-ghost-surface"
							on:click={clearInput}
							aria-label="Clear"
						>
							×
						</button>
					{/if}
				</div>
				<hr class="opacity-50" />
				<div class="mt-2 ml-2 min-h-[2rem]">
					{#if loading}
						<p class="opacity-60">Looking up…</p>
					{:else if error}
						<p class="text-error-500">Failed to look up: {error}</p>
					{:else if input && policies === undefined}
						<p>Lookup your secretpath to see which policies can access it.</p>
					{:else if policies && policies.length === 0}
						<p>Looks like no policies have access to this secret.</p>
					{:else if policies && policies.length > 0}
						<p>The following policies can access this secret:</p>
						<ul class="list-disc ml-5">
							{#each policies as policy (policy)}
								<li>{policy}</li>
							{/each}
						</ul>
					{:else}
						<p>Lookup your secretpath to see which policies can access it.</p>
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
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
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
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
					<a href="/discover">
						<button type="button" class="btn variant-filled">Go</button>
					</a>
				</footer>
			</div>
		</div>
	</div>
</div>
EOF echo done
