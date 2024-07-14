<script lang="ts">
	import * as d3 from 'd3';

	const background_color = 'white';

	interface GreekGod {
		id: string;
		parents?: string[];
	}

	interface Node {
		id: string;
		data: GreekGod;
		parents: Node[];
		level: number;
		bundle?: Bundle;
		bundles: Bundle[];
		bundles_index?: Map<string, Bundle[]>;
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

	interface Level {
		nodes: Node[];
		length: number;
	}

	interface Link {
		source: Node;
		bundle: Bundle;
		target: Node;
		xt: number;
		yt: number;
		yb: number;
		ys: number;
		c1: number;
		c2: number;
		xb: number;
		xs: number;
	}

	const data = [
		[{ id: 'Chaos' }],
		[{ id: 'Gaea', parents: ['Chaos'] }, { id: 'Uranus' }]
		,
		[
			{ id: 'Oceanus', parents: ['Gaea', 'Uranus'] },
			{ id: 'Thethys', parents: ['Gaea', 'Uranus'] },
			{ id: 'Pontus' },
			{ id: 'Rhea', parents: ['Gaea', 'Uranus'] },
			{ id: 'Cronus', parents: ['Gaea', 'Uranus'] },
			{ id: 'Coeus', parents: ['Gaea', 'Uranus'] },
			{ id: 'Phoebe', parents: ['Gaea', 'Uranus'] },
			{ id: 'Crius', parents: ['Gaea', 'Uranus'] },
			{ id: 'Hyperion', parents: ['Gaea', 'Uranus'] },
			{ id: 'Iapetus', parents: ['Gaea', 'Uranus'] },
			{ id: 'Thea', parents: ['Gaea', 'Uranus'] },
			{ id: 'Themis', parents: ['Gaea', 'Uranus'] },
			{ id: 'Mnemosyne', parents: ['Gaea', 'Uranus'] }
		],
		[
			{ id: 'Doris', parents: ['Oceanus', 'Thethys'] },
			{ id: 'Neures', parents: ['Pontus', 'Gaea'] },
			{ id: 'Dionne' },
			{ id: 'Demeter', parents: ['Rhea', 'Cronus'] },
			{ id: 'Hades', parents: ['Rhea', 'Cronus'] },
			{ id: 'Hera', parents: ['Rhea', 'Cronus'] },
			{ id: 'Alcmene' },
			{ id: 'Zeus', parents: ['Rhea', 'Cronus'] },
			{ id: 'Eris' },
			{ id: 'Leto', parents: ['Coeus', 'Phoebe'] },
			{ id: 'Amphitrite' },
			{ id: 'Medusa' },
			{ id: 'Poseidon', parents: ['Rhea', 'Cronus'] },
			{ id: 'Hestia', parents: ['Rhea', 'Cronus'] }
		],
		[
			{ id: 'Thetis', parents: ['Doris', 'Neures'] },
			{ id: 'Peleus' },
			{ id: 'Anchises' },
			{ id: 'Adonis' },
			{ id: 'Aphrodite', parents: ['Zeus', 'Dionne'] },
			{ id: 'Persephone', parents: ['Zeus', 'Demeter'] },
			{ id: 'Ares', parents: ['Zeus', 'Hera'] },
			{ id: 'Hephaestus', parents: ['Zeus', 'Hera'] },
			{ id: 'Hebe', parents: ['Zeus', 'Hera'] },
			{ id: 'Hercules', parents: ['Zeus', 'Alcmene'] },
			{ id: 'Megara' },
			{ id: 'Deianira' },
			{ id: 'Eileithya', parents: ['Zeus', 'Hera'] },
			{ id: 'Ate', parents: ['Zeus', 'Eris'] },
			{ id: 'Leda' },
			{ id: 'Athena', parents: ['Zeus'] },
			{ id: 'Apollo', parents: ['Zeus', 'Leto'] },
			{ id: 'Artemis', parents: ['Zeus', 'Leto'] },
			{ id: 'Triton', parents: ['Poseidon', 'Amphitrite'] },
			{ id: 'Pegasus', parents: ['Poseidon', 'Medusa'] },
			{ id: 'Orion', parents: ['Poseidon'] },
			{ id: 'Polyphemus', parents: ['Poseidon'] }
		],
		[
			{ id: 'Deidamia' },
			{ id: 'Achilles', parents: ['Peleus', 'Thetis'] },
			{ id: 'Creusa' },
			{ id: 'Aeneas', parents: ['Anchises', 'Aphrodite'] },
			{ id: 'Lavinia' },
			{ id: 'Eros', parents: ['Hephaestus', 'Aphrodite'] },
			{ id: 'Helen', parents: ['Leda', 'Zeus'] },
			{ id: 'Menelaus' },
			{ id: 'Polydueces', parents: ['Leda', 'Zeus'] }
		],
		[
			{ id: 'Andromache' },
			{ id: 'Neoptolemus', parents: ['Deidamia', 'Achilles'] },
			{ id: 'Aeneas(2)', parents: ['Creusa', 'Aeneas'] },
			{ id: 'Pompilius', parents: ['Creusa', 'Aeneas'] },
			{ id: 'Iulus', parents: ['Lavinia', 'Aeneas'] },
			{ id: 'Hermione', parents: ['Helen', 'Menelaus'] }
		]
	];
	let options = {
		c: 16,
		bigc: 70
	};


	// layout
	const padding = 8;
	const node_height = 22;
	const node_width = 70;
	const bundle_width = 14;
	const level_y_padding = 16;
	const metro_d = 4;
	const min_family_height = 22;

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
					const min_level = d3.min(node.parents, p => p.level);
					if (min_level === undefined) {
						return;
					}
					const bundle: Bundle = {
						i: 0,
						links: [],
						x: 0,
						y: 0,
						id: id,
						parents: node.parents.slice(),
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
		nodes.forEach((d: Node) => {
			d.parents.forEach((p: Node) => {
				if (d.bundle === undefined) {
					return;
				}
				const link: Link = {
					c1: 0,
					c2: 0,
					xb: 0,
					xs: 0,
					xt: 0,
					yb: 0,
					ys: 0,
					yt: 0,
					source: d,
					bundle: d.bundle,
					target: p
				};
				links.push(link);
			});
		});
		const bundles: Bundle[] = Array.from(bundlesMap.values());

		// reverse pointer from parent to bundles
		bundles.forEach((b: Bundle) => {
			b.parents.forEach((p: Node) => {
				if (p.bundles_index === undefined) {
					p.bundles_index = new Map<string, Bundle[]>();
				}
				let value = p.bundles_index.get(b.id);
				if (value === undefined) {
					value = [];
				}
				value.push(b);
				p.bundles_index.set(b.id, value);
				value.forEach((bundle: Bundle) => {
					bundlesMap.set(bundle.id, bundle);
				});
			});
		});

		nodes.forEach((node: Node) => {
			if (node.bundles_index !== undefined) {
				node.bundles = Array.from(node.bundles_index.values());
			} else {
				node.bundles_index = new Map<string, Bundle[]>();
				node.bundles = [];
			}

			node.bundles.sort((a, b) => d3.descending(d3.max(a, d => d.span), d3.max(b, d => d.span)));
			node.bundles.forEach((b, i) => (b.i = i));
		});

		links.forEach((link: Link) => {
			if (link.bundle === undefined) {
				return;
			}
			const linkBundle: Bundle | undefined = bundlesMap.get(link.bundle.id);
			if (linkBundle) {
				linkBundle.links.push(link);
			}
		});

		options.c ||= 16;
		const c = options.c;
		options.bigc ||= node_width + c;

		nodes.forEach((node: Node) => {
				node.height = (Math.max(1, node.bundles.length) - 1) * metro_d;
			}
		);

		let x_offset = padding;
		let y_offset = padding;

		levels.forEach((level: Level) => {
			const levelBundles: Bundle[] = bundles.filter(b => b.level === level.nodes[0].level);
			x_offset += levelBundles.length * bundle_width;
			y_offset += level_y_padding;
			level.nodes.forEach((n: Node) => {
				n.x = n.level * node_width + x_offset;
				n.y = node_height + y_offset + n.height / 2;

				y_offset += node_height + n.height;
			});
		});

		let i = 0;
		levels.forEach((level: Level) => {
			const levelBundles: Bundle[] = bundles.filter(b => b.level === level.nodes[0].level);
			levelBundles.forEach((bundle: Bundle) => {
				const parent_x: number[] = bundle.parents.map(d => d.x);
				const max_x: number = d3.max(parent_x) || 0;
				bundle.x =
					max_x +
					node_width +
					(levelBundles.length - 1 - bundle.i) * bundle_width;
				bundle.y = i * node_height;
			});
			i += level.length;
		});

		links.forEach((link: Link) => {
			if (link.target.bundles_index === undefined) {
				return;
			}
			const linkBundle: Bundle | undefined = bundlesMap.get(link.bundle.id);
			if (linkBundle === undefined) {
				return;
			}

			link.xt = link.target.x;
			link.yt =
				link.target.y +
				linkBundle.i * metro_d -
				(link.target.bundles.length * metro_d) / 2 +
				metro_d / 2;
			link.xb = linkBundle.x;
			link.yb = linkBundle.y;
			link.xs = link.source.x;
			link.ys = link.source.y;
		});

		// compress vertical space
		let y_negative_offset = 0;

		levels.forEach((level: Level) => {
			const levelBundles: Bundle[] = bundles.filter(b => b.level === level.nodes[0].level);
			const minimum: number = d3.min(levelBundles, bundle =>
				d3.min(bundle.links, link => link.ys - 2 * c - (link.yt + c))
			) || 0;
			y_negative_offset +=
				-min_family_height + minimum;
			level.nodes.forEach((node: Node) => (node.y -= y_negative_offset));
		});

		// very ugly, I know
		links.forEach((link: Link) => {
			if (link.target.bundles_index === undefined) {
				return;
			}
			const targetBundle: Bundle[] | undefined = link.target.bundles_index.get(link.bundle.id);
			if (targetBundle === undefined) {
				return;
			}
			const c1 = link.source.level - link.target.level > 1 ? Math.min(options.bigc, link.xb - link.xt, link.yb - link.yt) - c : c;
			link.yt =
				link.target.y +
				targetBundle.i * metro_d -
				(link.target.bundles.length * metro_d) / 2 +
				metro_d / 2;
			link.ys = link.source.y;
			link.c1 = c1;
			link.c2 = c;
		});
		const maxWidth: number = d3.max(nodes, node => node.x) || 0;
		const maxHeight: number = d3.max(nodes, node => node.y) || 0;
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
	let color = d3.scaleOrdinal(d3.schemeDark2);
	const renderChart = (data: any[]) => {
		options.color ||= (d: any, i: number) => color(i);

		const nodes: Node[][] = data.map((greekGods: GreekGod[]) => greekGods.map((god: GreekGod) => ({ data: god })));
		const levels: Level[] = nodes.map((n: Node[]) => ({ nodes: n, bundles: [], length: n.length}));

		return constructTangleLayout(levels);
	};
	$: tangleLayout = renderChart(data);
</script>
<div class="m-20">
	<svg width="{tangleLayout.layout.width}" height="{
			tangleLayout.layout.height
		}" style="background-color: {background_color}">
		<style>
        text {
            font-family: sans-serif;
            font-size: 10px;
        }

        .node {
            stroke-linecap: round;
        }

        .link {
            fill: none;
        }
		</style>
		{#each tangleLayout.bundles as b, i}
			{#each b.links as l}
				{#if !isNaN(l.c1) }
					<path class="link"
								d="
							M{l.xt} {l.yt}
							L{l.xb - l.c1} {l.yt}
							A{l.c1} {l.c1} 90 0 1 {l.xb} {l.yt + l.c1}
							L{l.xb} {l.ys - l.c2}
							A{l.c2} {l.c2} 90 0 0 {l.xb + l.c2} {l.ys}
							L{l.xs} {l.ys}"
								stroke="{background_color}" stroke-width="5" />
					<path class="link"
								d="
							M{l.xt} {l.yt}
							L{l.xb - l.c1} {l.yt}
							A{l.c1} {l.c1} 90 0 1 {l.xb} {l.yt + l.c1}
							L{l.xb} {l.ys - l.c2}
							A{l.c2} {l.c2} 90 0 0 {l.xb + l.c2} {l.ys}
							L{l.xs} {l.ys}"
								stroke="{options.color(b, i)}" stroke-width="2" />
				{/if}
			{/each}
		{/each}
		{#each tangleLayout.nodes as n}
			<path class="selectable node" data-id="{n.data.id}" stroke="black" stroke-width="8"
						d="M{n.x} {n.y - n.height / 2} L{n.x} {n.y + n.height / 2}" />
			<path class="node" stroke="white" stroke-width="4" d="M{n.x} {n.y - n.height / 2} L{n.x} {n.y + n.height / 2}" />

			<text class="selectable" data-id="{n.data.id}" x="{n.x + 4}" y="{n.y - n.height / 2 - 4}"
						stroke="{background_color}"
						stroke-width="2">{n.data.id}</text>
			<text x="{n.x + 4}" y="{n.y - n.height / 2 - 4}" style="pointer-events: none;">{n.data.id}</text>
		{/each}
	</svg>
</div>