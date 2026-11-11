<script lang="ts">
	/* eslint-disable svelte/prefer-svelte-reactivity */
	/* eslint-disable @typescript-eslint/no-unused-vars */
	import * as d3 from 'd3';
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
				bundle.x =
					bundle.parent.x + node_width + (levelBundles.length - 1 - bundle.i) * bundle_width;
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
		const id: string = target.getAttribute('data-id') ?? '';
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
		const id: string = (e.currentTarget as unknown as Element).getAttribute('data-id') ?? '';
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
		const id: string = target.getAttribute('data-id') ?? '';
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

	let canvas: HTMLCanvasElement;
	let hoveredId: string | undefined = undefined;
	let scale = 1;
	let tx = 0;
	let ty = 0;
	let isPanning = false;
	let rafPending = false;
	let panStart = { x: 0, y: 0, tx: 0, ty: 0 };
	let mouseScreen = { x: 0, y: 0 };
	let nodePositions: { id: string; x: number; y: number; height: number; name: string }[] = [];

	function computeFocus(id: string | undefined): { focused: Set<string>; faded: Set<string> } {
		const focused = new Set<string>();
		const faded = new Set<string>();
		if (!id) return { focused, faded };
		const parents = paths.get(id);
		if (parents) parents.forEach((p) => focused.add(p));
		focused.add(id);
		allIds.forEach((aid) => {
			if (!focused.has(aid)) faded.add(aid);
		});
		return { focused, faded };
	}

	function resetView() {
		scale = 1;
		tx = 0;
		ty = 0;
		hoveredId = undefined;
		drawChart();
	}

	function scheduleDraw() {
		if (rafPending) return;
		rafPending = true;
		requestAnimationFrame(() => {
			rafPending = false;
			drawChart();
		});
	}

	function drawChart() {
		if (!canvas || !tangleLayout) return;
		const ctx = canvas.getContext('2d');
		if (!ctx) return;
		const dpr = window.devicePixelRatio || 1;
		const w = tangleLayout.layout.width;
		const h = tangleLayout.layout.height;
		canvas.width = w * dpr;
		canvas.height = h * dpr;
		canvas.style.width = w + 'px';
		canvas.style.height = h + 'px';
		ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
		ctx.lineCap = 'round';
		ctx.lineJoin = 'round';
		ctx.clearRect(0, 0, canvas.width / dpr, canvas.height / dpr);
		ctx.setTransform(dpr * scale, 0, 0, dpr * scale, dpr * tx, dpr * ty);

		const activeId = hoveredId ?? highlightedNode;
		const isDark = document.documentElement.classList.contains('dark');
		const fg = isDark ? '#fff' : '#000';
		const bg = isDark ? '#000' : '#fff';
		const { focused, faded } = computeFocus(activeId);
		const hasFocus = focused.size > 0;

		nodePositions = [];

		tangleLayout.bundles.forEach((b: Bundle, bi: number) => {
			const col = color(b, bi);
			b.links.forEach((link: Link) => {
				const fadedLink =
					hasFocus && (faded.has(link.source.data.id) || faded.has(link.target.data.id));
				ctx.globalAlpha = fadedLink ? 0.15 : 1;
				const d = `M${link.x_target} ${link.y_target + 2} L${link.x_bundle - c} ${link.y_target + 2} A${c} ${c} 90 0 1 ${link.x_bundle} ${link.y_target + 2 + c} L${link.x_bundle} ${link.y_source - c} A${c} ${c} 90 0 0 ${link.x_bundle + c} ${link.y_source} L${link.x_source} ${link.y_source}`;
				const p = new Path2D(d);
				ctx.strokeStyle = isDark ? '#000' : '#9ca3af';
				ctx.lineWidth = 5;
				ctx.stroke(p);
				ctx.strokeStyle = col;
				ctx.lineWidth = 2;
				ctx.stroke(p);
			});
		});

		tangleLayout.nodes.forEach((node: Node) => {
			const id = node.data.id;
			const fadedNode = hasFocus && faded.has(id);
			ctx.globalAlpha = fadedNode ? 0.15 : 1;
			ctx.fillStyle = fg;
			ctx.beginPath();
			ctx.arc(node.x, node.y, focused.has(id) ? 7 : 5, 0, Math.PI * 2);
			ctx.fill();
			ctx.fillStyle = bg;
			ctx.beginPath();
			ctx.arc(node.x, node.y, focused.has(id) ? 4 : 3, 0, Math.PI * 2);
			ctx.fill();
			ctx.globalAlpha = fadedNode ? 0.15 : 1;
			ctx.fillStyle = fg;
			ctx.font = '300 15px "Roboto Mono Thin", monospace';
			ctx.textAlign = 'start';
			ctx.fillText(node.data.name, node.x, node.y - node.height / 2 - 8);
			nodePositions.push({ id, x: node.x, y: node.y, height: node.height, name: node.data.name });
		});
		// tooltip
		ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
		if (hoveredId) {
			const node = nodePositions.find((n) => n.id === hoveredId);
			if (node) {
				ctx.font = '12px sans-serif';
				const tw = ctx.measureText(node.name).width;
				const x = mouseScreen.x + 12;
				const y = mouseScreen.y - 24;
				ctx.fillStyle = isDark ? 'rgba(255,255,255,0.9)' : 'rgba(0,0,0,0.75)';
				ctx.fillRect(x, y, tw + 8, 18);
				ctx.fillStyle = isDark ? '#000' : '#fff';
				ctx.textAlign = 'start';
				ctx.fillText(node.name, x + 4, y + 13);
			}
		}
		ctx.globalAlpha = 1;
	}

	function findNode(mx: number, my: number): string | undefined {
		for (let i = nodePositions.length - 1; i >= 0; i--) {
			const n = nodePositions[i];
			if (
				Math.abs(mx - n.x) <= 12 &&
				my >= n.y - n.height / 2 - 16 &&
				my <= n.y + n.height / 2 + 4
			) {
				return n.id;
			}
		}
		return undefined;
	}

	function toWorld(mx: number, my: number) {
		return { x: (mx - tx) / scale, y: (my - ty) / scale };
	}

	function handleMove(e: MouseEvent) {
		if (!canvas) return;
		const rect = canvas.getBoundingClientRect();
		const mx = e.clientX - rect.left;
		const my = e.clientY - rect.top;
		mouseScreen = { x: mx, y: my };
		if (isPanning) {
			tx = panStart.tx + (e.clientX - panStart.x);
			ty = panStart.ty + (e.clientY - panStart.y);
			scheduleDraw();
			return;
		}
		const w = toWorld(mx, my);
		const id = findNode(w.x, w.y);
		if (id !== hoveredId) {
			hoveredId = id;
			drawChart();
		}
	}

	function handleMouseDown(e: MouseEvent) {
		if (!canvas) return;
		isPanning = true;
		panStart = { x: e.clientX, y: e.clientY, tx, ty };
	}

	function handleMouseUp() {
		isPanning = false;
	}

	function handleClick(e: MouseEvent) {
		if (!canvas || isPanning) return;
		const rect = canvas.getBoundingClientRect();
		const w = toWorld(e.clientX - rect.left, e.clientY - rect.top);
		const id = findNode(w.x, w.y);
		if (!id) return;
		highlightedNode = highlightedNode === id ? undefined : id;
		drawChart();
	}

	function handleWheel(e: WheelEvent) {
		if (!canvas) return;
		e.preventDefault();
		const rect = canvas.getBoundingClientRect();
		const mx = e.clientX - rect.left;
		const my = e.clientY - rect.top;
		const factor = e.deltaY < 0 ? 1.1 : 0.9;
		const newScale = Math.min(Math.max(scale * factor, 0.2), 6);
		tx = mx - (mx - tx) * (newScale / scale);
		ty = my - (my - ty) * (newScale / scale);
		scale = newScale;
		drawChart();
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (!canvas) return;
		const step = 40;
		switch (e.key) {
			case 'ArrowLeft':
				tx += step;
				break;
			case 'ArrowRight':
				tx -= step;
				break;
			case 'ArrowUp':
				ty += step;
				break;
			case 'ArrowDown':
				ty -= step;
				break;
			case '+':
			case '=':
				scale = Math.min(scale * 1.1, 6);
				break;
			case '-':
				scale = Math.max(scale * 0.9, 0.2);
				break;
			case 'Enter':
				if (hoveredId) {
					highlightedNode = highlightedNode === hoveredId ? undefined : hoveredId;
					drawChart();
				}
				return;
			case 'Escape':
				scale = 1;
				tx = 0;
				ty = 0;
				hoveredId = undefined;
				highlightedNode = undefined;
				drawChart();
				return;
			default:
				return;
		}
		e.preventDefault();
		drawChart();
	}

	$: if (canvas && tangleLayout) drawChart();
</script>

{#if show}
	<div class="container hide-scrollbar content-center overflow-auto">
		{#if highlightedNode}
			<div class="absolute end-60" transition:fade={{ duration: 50 }}>
				<SecretInfoBox secret={highlightedNode} {information} />
			</div>
		{/if}
		<div class="relative">
			<canvas
				bind:this={canvas}
				class="z-0 cursor-grab active:cursor-grabbing"
				tabindex="0"
				aria-label="Subway chart of secret paths. Use arrow keys to pan, plus and minus to zoom, Enter to select, Escape to reset."
				on:mousemove={handleMove}
				on:mousedown={handleMouseDown}
				on:mouseup={handleMouseUp}
				on:mouseleave={handleMouseUp}
				on:click={handleClick}
				on:wheel={handleWheel}
				on:keydown={handleKeyDown}
			></canvas>
			<div class="absolute top-2 right-2">
				<button type="button" class="btn btn-sm variant-ghost-surface" on:click={resetView}
					>Reset view</button
				>
			</div>
		</div>
	</div>
{:else}
	<div class="container justify-center items-center h-full mx-auto flex">
		<ProgressBar class="w-1/2 content-center" />
	</div>
{/if}
