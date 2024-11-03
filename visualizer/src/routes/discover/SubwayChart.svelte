<script lang="ts">
	import * as d3 from 'd3';
	import type { Element } from 'svelte/types/compiler/interfaces';
	import { draw, fade } from 'svelte/transition';
	import { onMount } from 'svelte';
	import { ProgressBar } from '@skeletonlabs/skeleton';
	import SecretInfoBox from './SecretInfoBox.svelte';
	import type { Information } from '../../types';

	interface SubwayStation {
		id: string;
		level: number;
		name: string;
		parent?: string;
	}

	interface Node {
		id: string;
		data: SubwayStation;
		parent: Node;
		level: number;
		bundle?: Bundle;
		bundles: Bundles;
		bundlesKeys?: string[];
		links?: Link[];
		height: number;
		width?: number;
		x: number;
		y: number;
		xt: number;
		yt: number;
		ys: number;
		xb: number;
		xs: number;
	}

	interface Bundle {
		id: string;
		parent: Node;
		level: number;
		span: number;
		links: Link[];
		i: number;
		x: number;
		y: number;
	}

	interface Bundles {
		bundles: Bundle[][];
		i: number;
	}

	interface Level {
		nodes: Node[];
		length: number;
	}

	interface Link {
		source: Node;
		bundle: Bundle;
		target: Node;
		x_target: number;
		y_target: number;
		y_bundle: number;
		y_source: number;
		x_bundle: number;
		x_source: number;
	}

	export let data: SubwayStation[][];

	export let information: Information;

	let options = {
		c: 16,
		bigc: 70
	};

	// layout
	const padding = 8;
	const node_height = 30;
	const node_width = 70;
	const bundle_width = 14;
	const level_y_padding = 16;
	const metro_d = 4;
	const half_metro_d = metro_d / 2;
	const min_family_height = 22;
	let show = data.length < 4;
	let highlightedNode: string | undefined = undefined;

	const d3Color = d3.scaleOrdinal(d3.schemeDark2);

	options.c ||= 16;
	const c = options.c;
	options.bigc ||= node_width + c;

	const paths: Map<string, string[]> = new Map();
	let allIds: Set<string> = new Set();

	const constructTangleLayout = (levels: Level[], nodes: Node[]) => {
		const bundlesMap: Map<string, Bundle> = new Map();

		// precompute bundles
		levels.forEach((level: Level, i: number) => {
			let index = 0;
			const map: Map<string, Bundle> = new Map();
			level.nodes.forEach((node: Node) => {
				if (node.parent === undefined) {
					return;
				}

				const id = node.parent.data.id;
				const value = map.get(id);
				if (value) {
					value.parent = node.parent;
					map.set(id, value);
				} else {
					const bundle: Bundle = {
						i: index++,
						links: [],
						x: 0,
						y: 0,
						id: id,
						parent: node.parent,
						level: i,
						span: i - 1
					};
					map.set(id, bundle);
					bundlesMap.set(id, bundle);
				}
				node.bundle = map.get(id);
				return node;
			});
		});

		const links: Link[] = [];

		nodes.forEach((node: Node) => {
			if (node.bundle === undefined) {
				return;
			}
			let parent = node.parent;
			const link: Link = {
				x_bundle: 0,
				x_source: 0,
				x_target: 0,
				y_bundle: 0,
				y_source: 0,
				y_target: 0,
				source: node,
				bundle: node.bundle,
				target: parent
			};
			const pathsToHere = paths.get(node.data.id) || [parent.data.id];
			paths.set(node.data.id, pathsToHere);
			const linkBundle: Bundle | undefined = bundlesMap.get(link.bundle.id);
			if (linkBundle) {
				linkBundle.links.push(link);
			}
			links.push(link);
		});

		paths.forEach((value: string[], key: string) => {
			const newPaths: Set<string> = new Set(value);
			value.forEach((v: string) => {
				const pathsToParent = paths.get(v) || [];
				pathsToParent.forEach((p: string) => newPaths.add(p));
			});
			paths.set(key, Array.from(newPaths));
		});

		paths.forEach((value, key) => {
			allIds.add(key);
			value.forEach((id: string) => allIds.add(id));
		});

		const bundles: Bundle[] = Array.from(bundlesMap.values());

		let x_offset = padding;
		let y_offset = padding;

		let i = 0;

		levels.forEach((level: Level) => {
			const levelBundles: Bundle[] = bundles.filter((b) => b.level === level.nodes[0].level);
			x_offset += levelBundles.length * bundle_width;
			y_offset += level_y_padding;
			level.nodes.forEach((node: Node) => {
				node.x = node.level * node_width + x_offset;
				node.y = node_height + y_offset;
				y_offset += node_height;
			});
			levelBundles.forEach((bundle: Bundle) => {
				bundle.x = bundle.parent.x + node_width + (levelBundles.length - 1 - bundle.i) * bundle_width;
				bundle.y = i * node_height;
			});
			i += level.length;
		});

		links.forEach((link: Link) => {
			const linkBundle: Bundle = bundlesMap.get(link.bundle.id) as Bundle;

			link.x_target = link.target.x;
			link.y_target = link.target.y + linkBundle.i * metro_d - half_metro_d;
			link.x_bundle = linkBundle.x;
			link.y_bundle = linkBundle.y;
			link.x_source = link.source.x;
			link.y_source = link.source.y;
		});

		// compress vertical space
		let y_negative_offset = 0;

		levels.forEach((level: Level) => {
			const levelBundles: Bundle[] = bundles.filter((b) => b.level === level.nodes[0].level);
			const minimum: number =
				d3.min(levelBundles, (bundle) =>
					d3.min(bundle.links, (link) => link.y_source - 2 * c - (link.y_target + c))
				) || 0;
			y_negative_offset += -min_family_height + minimum;
			level.nodes.forEach((node: Node) => (node.y -= y_negative_offset));
		});

		links.forEach((link: Link) => {
			link.y_target = link.target.y - half_metro_d;
			link.y_source = link.source.y;
		});

		const maxWidth: number = d3.max(nodes, (node) => node.x) || 0;
		const maxHeight: number = d3.max(nodes, (node) => node.y) || 0;
		const layout = {
			width: maxWidth + node_width + 2 * padding,
			height: maxHeight + node_height / 2 + 2 * padding,
			node_height,
			node_width,
			bundle_width,
			level_y_padding,
			metro_d
		};

		return { nodes, bundles, layout };
	};
	const color = (_: Bundle, i: number) => d3Color(i.toString());
	const renderChart = (data: SubwayStation[][]) => {
		const flatNodes: Node[] = [];
		const nodes: Node[][] = [];

		data.forEach((stations: SubwayStation[]) => {
			let previousLevel: Node[] = [];
			if (nodes.length > 0) {
				previousLevel = nodes[nodes.length - 1];
			}
			nodes.push(
				stations.map((station: SubwayStation) => {
					let parent: Node | undefined = undefined;
					if (previousLevel) {
						parent = previousLevel.find((node: Node) => node.data.id === station.parent);
					}
					return { data: station, level: station.level, parent: parent, height: 0 } as Node;
				})
			);
			flatNodes.push(...nodes[nodes.length - 1]);
		});
		const levels: Level[] = nodes.map((n: Node[]) => ({ nodes: n, bundles: [], length: n.length }));

		return constructTangleLayout(levels, flatNodes);
	};

	$: tangleLayout = renderChart(data);

	onMount(() => {
		show = true;
	});

	function getFocusedElements(id: string): HTMLCollection[] {
		const parents: string[] | undefined = paths.get(id);
		if (!parents) {
			return [];
		}
		parents.push(id);
		const pathsToHighlight = parents.map((p: string) =>
			document.getElementsByClassName('belongs-to-' + p)
		);
		if (!pathsToHighlight) {
			return [];
		}
		return pathsToHighlight;
	}

	function getOutOfFocusElements(id: string): HTMLCollection[] {
		const hide: string[] = [];
		const parents: string[] | undefined = paths.get(id);
		if (!parents) {
			return [];
		}
		parents.push(id);
		allIds.forEach((id: string) => {
			if (!parents.includes(id)) {
				hide.push(id);
			}
		});
		const pathsToHide = hide.map((p: string) => document.getElementsByClassName('belongs-to-' + p));
		if (!pathsToHide) {
			return [];
		}
		return pathsToHide;
	}

	function entryMouseOver(e: MouseEvent) {
		if (!e.currentTarget) {
			return;
		}
		const target: Element = e.currentTarget as unknown as Element;
		const id: string = target.getAttribute('data-id');
		getFocusedElements(id).forEach((p: HTMLCollection) => {
			for (let element of p) {
				if (element.tagName === 'text') {
					continue;
				}
				if (element.classList.contains('outer-dot')) {
					element.classList.add('outer-dot-highlight');
				} else if (element.classList.contains('inner-dot')) {
					element.classList.add('inner-dot-highlight');
				}
				element.classList.add('highlight');
			}
		});
		getOutOfFocusElements(id).forEach((p: HTMLCollection) => {
			for (let element of p) {
				element.classList.add('opacity-50');
			}
		});
	}

	async function showInfoBox(e: MouseEvent) {
		if (!e.target) {
			return;
		}
		const id: string = (e.currentTarget as unknown as Element).getAttribute('data-id');
		if (!id) {
			return;
		}
		if (highlightedNode === id) {
			highlightedNode = undefined;
			return;
		}
		highlightedNode = id;
	}

	function entryMouseOut(e: MouseEvent) {
		if (!e.currentTarget) {
			return;
		}
		const target: Element = e.currentTarget as unknown as Element;
		const id: string = target.getAttribute('data-id');
		getFocusedElements(id).forEach((p: HTMLCollection) => {
			for (let element of p) {
				if (element.tagName === 'text') {
					continue;
				}
				if (element.classList.contains('outer-dot')) {
					element.classList.remove('outer-dot-highlight');
				} else if (element.classList.contains('inner-dot')) {
					element.classList.remove('inner-dot-highlight');
				}
				element.classList.remove('highlight');
			}
		});
		getOutOfFocusElements(id).forEach((p: HTMLCollection) => {
			for (let element of p) {
				element.classList.remove('opacity-50');
			}
		});
	}
</script>

{#if show}
	<div class="container hide-scrollbar content-center overflow-auto">
		{#if highlightedNode}
			<div class="absolute end-60" transition:fade={{ duration: 50 }}>
				<SecretInfoBox secret={highlightedNode} {information} />
			</div>
		{/if}
		<svg width={tangleLayout.layout.width} height={tangleLayout.layout.height} class="z-0">
			<style>
				text {
					font-family: 'Roboto Mono Thin', sans-serif;
					font-weight: lighter;
					font-size: 15px;
				}

				.node {
					stroke-linecap: round;
				}

				.highlight {
					stroke-width: 7px;
				}

				.inner-dot-highlight {
					stroke-width: 7px;
				}

				.outer-dot-highlight {
					stroke-width: 15px;
				}
			</style>
			{#each tangleLayout.bundles as b, i}
				{#each b.links as link}
						<path
							class="fill-none belongs-to-{link.source.data.id} stroke-gray dark:stroke-black"
							in:draw|global={{ duration: 250, delay: i * 50 }}
							d="M{link.x_target} {link.y_target + 2}
							L{link.x_bundle - c} {link.y_target + 2}
							A{c} {c} 90 0 1 {link.x_bundle} {link.y_target + 2 + c}
							L{link.x_bundle} {link.y_source - c}
							A{c} {c} 90 0 0 {link.x_bundle + c} {link.y_source}
							L{link.x_source} {link.y_source}"
							stroke-width="5"
						/>
						<path
							class="fill-none belongs-to-{link.source.data.id} stroke-2"
							in:draw|global={{ duration: 250, delay: i * 50 }}
							d="M{link.x_target} {link.y_target + 2}
							L{link.x_bundle - c} {link.y_target + 2}
							A{c} {c} 90 0 1 {link.x_bundle} {link.y_target + 2 + c}
							L{link.x_bundle} {link.y_source - c}
							A{c} {c} 90 0 0 {link.x_bundle + c} {link.y_source}
							L{link.x_source} {link.y_source}"
							stroke={color(b, i)}
						/>
				{/each}
			{/each}
			{#each tangleLayout.nodes as node, i}
				<!-- little dot-->
				<path
					class="selectable node outer-dot belongs-to-{node.data.id} stroke-black dark:stroke-white"
					data-id={node.data.id}
					stroke-width="10"
					in:fade|global={{ duration: 200, delay: i * 5 }}
					d="M{node.x} {node.y - node.height / 2} L{node.x} {node.y + node.height / 2}"
				/>
				<!-- inner circle -->
				<path
					in:fade|global={{ duration: 200, delay: i * 5.1 }}
					class="node inner-dot belongs-to-{node.data.id} stroke-white dark:stroke-black"
					stroke-width="6"
					d="M{node.x} {node.y - node.height / 2} L{node.x} {node.y + node.height / 2}"
				/>

				<text
					class="fill-black dark:fill-white belongs-to-{node.data.id}"
					data-id={node.data.id}
					transition:fade|global={{ duration: 250, delay: i * 5 }}
					x={node.x}
					role="term"
					y={node.y - node.height / 2 - 8}
					on:mouseover|preventDefault={entryMouseOver}
					on:mouseout|preventDefault={entryMouseOut}
					on:click|preventDefault={showInfoBox}
					on:focus|preventDefault={() => {}}
					on:blur|preventDefault={() => {}}
				>
					{node.data.name}
				</text>
			{/each}
		</svg>
	</div>
{:else}
	<div class="container justify-center items-center h-full mx-auto flex">
		<ProgressBar class="w-1/2 content-center" />
	</div>
{/if}
