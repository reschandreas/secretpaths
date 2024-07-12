<script lang="ts">
	import * as d3 from 'd3';
	import { blur } from 'svelte/transition';
	import type { Policy } from '../types.js';
	import { onMount } from 'svelte';

	export let policies: Policy[];
	let width: number;
	const margin = 1;

	const name = (d: Policy) => d.name;
	const names = (d: Policy) => name(d).split(/(?=[A-Z][a-z])|\s+/g);

	const format = d3.format(',d');
	const color = d3.scaleOrdinal(d3.schemeTableau10);

	const getSize = (d: Policy | undefined) => {
		if (d === undefined || !d.rules || !d.rules.length) {
			return 0;
		}
		return d.rules.length;
	};

	let show = false;

	$: pack = d3.pack()
		.size([width - margin * 2, width - margin * 2])
		.padding(3);

	$: root = pack(d3.hierarchy({ children: policies })
		.sum((d) => getSize(d)));

	onMount(() => {
		show = true;
	});

</script>

<div class="container" bind:clientWidth={width}>
	<svg width="{width}" height="{width}">
		{#if show}
			<g transform={`translate(${margin},${margin})`}>
				{#each policies as policy, i}
					<g transition:blur|global={{duration: 500, delay: i * 100}}
						 transform={`translate(${root.leaves().find(d => d.data === policy)?.x},${root.leaves().find(d => d.data === policy)?.y})`}>
						<circle fill-opacity="0.7" fill={color(name(policy))} r={root.leaves().find(d => d.data === policy)?.r}
						/>
						<text clip-path={`circle(${root.leaves().find(d => d.data === policy)?.r})`} class="text-white">
							{#each names(policy) as name, index}
								<tspan x="0" y="${index - names(policy).length / 2 + 0.35}em">{name}</tspan>
							{/each}
							<tspan x="0" y="1em">{format(policy.rules.length)} Rules</tspan>
							<title>{policy.name} has {format(policy.rules.length)} Policies</title>
						</text>
					</g>
				{/each}
			</g>
		{/if}
	</svg>
</div>
<style>
    svg {
        text-anchor: middle;
        display: block;
        margin: auto;
    }
</style>