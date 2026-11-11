<script lang="ts">
	/* eslint-disable svelte/prefer-svelte-reactivity */
	/* eslint-disable @typescript-eslint/no-unused-vars */
	import * as d3 from 'd3';
	import { draw, fade } from 'svelte/transition';
	import { onMount } from 'svelte';
	import { ProgressBar } from '@skeletonlabs/skeleton';
	import SecretInfoBox from './SecretInfoBox.svelte';
	import type { Information } from '../../types';
	import type { Node, Bundle, Link, TangleLayout, SubwayStation } from '../../lib/layout';

	let worker: Worker | undefined;

	export let data: SubwayStation[][];

	export let information: Information;

	let show = data.length < 4;
	let highlightedNode: string | undefined = undefined;
	let tangleLayout: TangleLayout | undefined = undefined;
	let allIds: Set<string> = new Set();
	let colorFn = (_: Bundle, i: number) => '#6366f1';

	const d3Color = d3.scaleOrdinal(d3.schemeDark2);
	const c = 16;
	const color = (_: Bundle, i: number) => d3Color(i.toString());

	onMount(() => {
		worker = new Worker(new URL('../../lib/layout.worker.ts', import.meta.url), { type: 'module' });
		worker.onmessage = (e) => {
			tangleLayout = e.data as TangleLayout;
			allIds = new Set(tangleLayout.nodes.map((n) => n.data.id));
			show = true;
		};
		worker.postMessage(data);
		return () => worker?.terminate();
	});

	$: if (worker && data) {
		worker.postMessage(data);
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
		const node = tangleLayout?.nodes.find((n) => n.data.id === id);
		let cur: Node | undefined = node;
		while (cur) {
			focused.add(cur.data.id);
			cur = cur.parent;
		}
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
