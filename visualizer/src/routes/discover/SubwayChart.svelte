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
		parents?: string[];
	}

	interface Node {
		id: string;
		data: SubwayStation;
		parents: Node[];
		level: number;
		bundle?: Bundle;
		bundles: Bundles;
		children_bundles?: Map<string, ChildrenBundles>;
		bundlesKeys?: string[];
		links?: Link[];
		height: number;
		width?: number;
		x: number;
		y: number;
		xt: number;
		yt: number;
		ys: number;
		c1: number;
		c2: number;
		xb: number;
		xs: number;
	}

	interface ChildrenBundles {
		bundles: Bundle[];
		i: number;
	}

	interface Bundle {
		id: string;
		parents: Node[];
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
		c1: number;
		c2: number;
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
	const min_family_height = 22;
	let show = false;
	let highlightedNode: string | undefined = undefined;

	const paths: Map<string, string[]> = new Map();
	let allIds: Set<string> = new Set();

	const addDepthInformation = (levels: Level[]) => {
		levels.forEach((level: Level, i: number) => {
			level.nodes.forEach((node: Node) => {
				node.level = i;
			});
		});
	};

	const populateNodeMap = (nodes: Node[]) => {
		const nodes_index: Map<string, Node> = new Map();
		nodes.forEach((node: Node) => {
			nodes_index.set(node.data.id, node);
		});
		return nodes_index;
	};

	const getNodes = (levels: Level[]): Node[] => {
		const nodes: Node[] = [];
		levels.forEach((level: Level) => {
			nodes.push(...level.nodes);
		});
		return nodes;
	};

	const fix_parents = (nodes: Node[], nodes_index: Map<string, Node>) => {
		nodes.forEach((node: Node) => {
			let parents: Node[] = [];
			if (node.data.parents !== undefined) {
				node.data.parents.forEach((parent: string) => {
					const value = nodes_index.get(parent);
					if (value) {
						parents.push(value);
					}
				});
			}
			node.parents = parents;
		});
	};

	const constructTangleLayout = (levels: Level[]) => {
		//precompute level depth
		addDepthInformation(levels);
		const nodes: Node[] = getNodes(levels);

		const nodesMap: Map<string, Node> = populateNodeMap(nodes);
		const bundlesMap: Map<string, Bundle> = new Map();

		fix_parents(nodes, nodesMap);

		// precompute bundles
		levels.forEach((level: Level, i: number) => {
			const map: Map<string, Bundle> = new Map();
			level.nodes.forEach((node: Node) => {
				if (node.parents?.length == 0) {
					return;
				}

				const id = node.parents
					.map((d: Node) => d.data.id)
					.sort()
					.join('-X-');
				const value = map.get(id);
				if (value) {
					value.parents = value.parents.concat(node.parents);
					map.set(id, value);
				} else {
					const min_level = d3.min(node.parents, (p) => p.level);
					if (min_level === undefined) {
						return;
					}
					const bundle: Bundle = {
						i: 0,
						links: [],
						x: 0,
						y: 0,
						id: id,
						parents: node.parents,
						level: i,
						span: i - min_level
					};
					map.set(id, bundle);
					bundlesMap.set(id, bundle);
				}
				node.bundle = map.get(id);
				return node;
			});

			let ids = 0;
			map.forEach((b: Bundle) => (b.i = ids++));
		});

		const links: Link[] = [];

		nodes.forEach((node: Node) => {
			node.parents.forEach((parent: Node) => {
				if (node.bundle === undefined) {
					return;
				}
				const link: Link = {
					c1: 0,
					c2: 0,
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
				const pathsToHere = paths.get(node.data.id) || [];
				pathsToHere.push(parent.data.id);
				paths.set(node.data.id, pathsToHere);
				links.push(link);
			});
		});

		paths.forEach((value: string[], key: string) => {
			const newPaths: Set<string> = new Set(value);
			value.forEach((v: string) => {
				const pathsToParent = paths.get(v) || [];
				pathsToParent.forEach((p: string) => newPaths.add(p));
			});
			paths.set(key, Array.from(newPaths));
		});

		Array.from(paths.keys())
			.map((k: string) => {
				const tmp: string[] = paths.get(k) || [];
				tmp.push(k);
				return tmp;
			})
			.flat()
			.forEach((id: string) => allIds.add(id));

		const bundles: Bundle[] = Array.from(bundlesMap.values());

		// reverse pointer from parent to bundles
		bundles.forEach((b: Bundle) => {
			b.parents.forEach((node: Node) => {
				if (node.children_bundles === undefined) {
					node.children_bundles = new Map<string, ChildrenBundles>();
				}
				let value = node.children_bundles.get(b.id);
				if (value === undefined) {
					value = {
						bundles: [b],
						i: 0
					};
				} else {
					value.bundles.push(b);
				}
				node.children_bundles.set(b.id, value);
			});
		});

		nodes.forEach((node: Node) => {
			if (node.children_bundles !== undefined) {
				const children: Bundle[][] = Array.from(node.children_bundles.keys()).map(
					(key: string) => node.children_bundles?.get(key)?.bundles || []
				);
				node.bundles = { bundles: children, i: 0 };
			} else {
				node.children_bundles = new Map<string, ChildrenBundles>();
				node.bundles = { bundles: [], i: 0 };
			}

			node.bundles.bundles.sort((a: Bundle[], b: Bundle[]) =>
				d3.descending(
					d3.max(a, (d) => d.span),
					d3.max(b, (d) => d.span)
				)
			);
			node.bundles.bundles.forEach((b: Bundle[], i) => (b.i = i));
		});

		links.forEach((link: Link) => {
			const linkBundle: Bundle | undefined = bundlesMap.get(link.bundle.id);
			if (linkBundle) {
				linkBundle.links.push(link);
			}
		});

		options.c ||= 16;
		const c = options.c;
		options.bigc ||= node_width + c;

		nodes.forEach((node: Node) => {
			node.height = (Math.max(1, node.bundles.bundles.length) - 1) * metro_d;
		});

		let x_offset = padding;
		let y_offset = padding;

		levels.forEach((level: Level) => {
			const levelBundles: Bundle[] = bundles.filter((b) => b.level === level.nodes[0].level);
			x_offset += levelBundles.length * bundle_width;
			y_offset += level_y_padding;
			level.nodes.forEach((node: Node) => {
				node.x = node.level * node_width + x_offset;
				node.y = node_height + y_offset + node.height / 2;

				y_offset += node_height + node.height;
			});
		});

		let i = 0;
		levels.forEach((level: Level) => {
			const levelBundles: Bundle[] = bundles.filter((b) => b.level === level.nodes[0].level);
			levelBundles.forEach((bundle: Bundle) => {
				const parent_x: number[] = bundle.parents.map((d) => d.x);
				const max_x: number = d3.max(parent_x) || 0;
				bundle.x = max_x + node_width + (levelBundles.length - 1 - bundle.i) * bundle_width;
				bundle.y = i * node_height;
			});
			i += level.length;
		});

		links.forEach((link: Link) => {
			if (link.target.children_bundles === undefined || !bundlesMap.has(link.bundle.id)) {
				return;
			}
			const linkBundle: Bundle = bundlesMap.get(link.bundle.id) as Bundle;

			link.x_target = link.target.x;
			link.y_target =
				link.target.y +
				linkBundle.i * metro_d -
				(link.target.bundles.bundles.length * metro_d) / 2 +
				metro_d / 2;
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
			if (link.target.children_bundles === undefined) {
				return;
			}
			const targetBundleOffset: number = link.target.children_bundles.get(link.bundle.id)?.i || 0;

			const c1 =
				link.source.level - link.target.level > 1
					? Math.min(options.bigc, link.x_bundle - link.x_target, link.y_bundle - link.y_target) - c
					: c;
			link.y_target =
				link.target.y +
				targetBundleOffset * metro_d -
				(link.target.bundles.bundles.length * metro_d) / 2 +
				metro_d / 2;
			link.y_source = link.source.y;
			link.c1 = c1;
			link.c2 = c;
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

		return { levels, nodes, nodes_index: nodesMap, links, bundles, layout };
	};
	let d3Color = d3.scaleOrdinal(d3.schemeDark2);
	const color = (_: Bundle, i: number) => d3Color(i.toString());
	const renderChart = (data: SubwayStation[][]) => {
		const nodes: Node[][] = data.map((stations: SubwayStation[]) =>
			stations.map((station: SubwayStation) => ({ data: station }) as Node)
		);
		const levels: Level[] = nodes.map((n: Node[]) => ({ nodes: n, bundles: [], length: n.length }));

		return constructTangleLayout(levels);
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
					{#if !isNaN(link.c1)}
						<path
							class="fill-none belongs-to-{link.source.data.id} stroke-gray dark:stroke-black"
							in:draw|global={{ duration: 250, delay: i * 50 }}
							d="
							M{link.x_target} {link.y_target}
							L{link.x_bundle - link.c1} {link.y_target}
							A{link.c1} {link.c1} 90 0 1 {link.x_bundle} {link.y_target + link.c1}
							L{link.x_bundle} {link.y_source - link.c2}
							A{link.c2} {link.c2} 90 0 0 {link.x_bundle + link.c2} {link.y_source}
							L{link.x_source} {link.y_source}"
							stroke-width="5"
						/>
						<path
							class="fill-none belongs-to-{link.source.data.id} stroke-2"
							in:draw|global={{ duration: 250, delay: i * 50 }}
							d="
							M{link.x_target} {link.y_target}
							L{link.x_bundle - link.c1} {link.y_target}
							A{link.c1} {link.c1} 90 0 1 {link.x_bundle} {link.y_target + link.c1}
							L{link.x_bundle} {link.y_source - link.c2}
							A{link.c2} {link.c2} 90 0 0 {link.x_bundle + link.c2} {link.y_source}
							L{link.x_source} {link.y_source}"
							stroke={color(b, i)}
						/>
					{/if}
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
