<script lang="ts">
	import * as d3 from 'd3';
	import _ from 'lodash';

	const background_color = 'white';

	interface GreekGod {
		id: string;
		parents?: GreekGod[];
		level?: number;
		bundle?: any;
		bundles?: any;
		bundles_index?: any[];
		links?: any;
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

	const constructTangleLayout = (levels: GreekGod[][]) => {
		//precompute level depth
		levels.forEach((level: GreekGod[], i: number) => {
			level.forEach((node: GreekGod) => {
				node.level = i;
			});
		});
		var nodes = levels.reduce((a, x) => a.concat(x), []);

		var nodes_index: any = {};
		nodes.forEach((d: GreekGod) => {
			nodes_index[d.id] = d;
		});

		// objectification
		nodes.forEach((d: GreekGod) => {
			d.parents = (d.parents === undefined ? [] : d.parents).map(
				(p: GreekGod) => nodes_index[p]
			);
		});

		// precompute bundles
		levels.forEach((l: GreekGod[], i: number) => {
			var index: any = {};
			l.forEach((n: GreekGod) => {
				if (n.parents?.length == 0) {
					return;
				}

				var id = n.parents
					.map((d: GreekGod) => d.id)
					.sort()
					.join('-X-');
				if (id in index) {
					index[id].parents = index[id].parents.concat(n.parents);
				} else {
					index[id] = { id: id, parents: n.parents.slice(), level: i, span: i - d3.min(n.parents, p => p.level) };
				}
				n.bundle = index[id];
				return n;
			});

			l.bundles = Object.keys(index).map(k => index[k]);
			l.bundles.forEach((b, i) => (b.i = i));
		});

		var links: any[] = [];
		nodes.forEach((d: any) => {
			d.parents.forEach((p: any) =>
				links.push({ source: d, bundle: d.bundle, target: p })
			);
		});
		var bundles = levels.reduce((a, x) => a.concat(x.bundles), []);

		// reverse pointer from parent to bundles
		bundles.forEach((b: any) => {
			if (b.parents === undefined) {
				b.parents = [];
			}
			b.parents.forEach((p: any) => {
				if (p.bundles_index === undefined) {
					p.bundles_index = {};
				}
				if (!(b.id in p.bundles_index)) {
					p.bundles_index[b.id] = [];
				}
				p.bundles_index[b.id].push(b);
			});
		});

		nodes.forEach((n: GreekGod) => {
			if (n.bundles_index !== undefined) {
				n.bundles = Object.keys(n.bundles_index).map(k => n.bundles_index[k]);
			} else {
				n.bundles_index = {};
				n.bundles = [];
			}
			n.bundles.sort((a, b) => d3.descending(d3.max(a, d => d.span), d3.max(b, d => d.span)));
			n.bundles.forEach((b, i) => (b.i = i));
		});

		links.forEach(l => {
			if (l.bundle === undefined) {
				return;
			}
			if (l.bundle.links === undefined) {
				l.bundle.links = [];
			}
			l.bundle.links.push(l);
		});


		options.c ||= 16;
		const c = options.c;
		options.bigc ||= node_width + c;

		nodes.forEach(
			(n: GreekGod) => {
				n.height = (Math.max(1, n.bundles.length) - 1) * metro_d;
			}
		);

		var x_offset = padding;
		var y_offset = padding;

		levels.forEach(l => {
			x_offset += l.bundles.length * bundle_width;
			y_offset += level_y_padding;
			l.forEach((n: GreekGod) => {
				n.x = n.level * node_width + x_offset;
				n.y = node_height + y_offset + n.height / 2;

				y_offset += node_height + n.height;
			});
		});

		var i = 0;
		levels.forEach(l => {
			l.bundles.forEach(b => {
				const max_x: number = parseInt(<string>d3.max(b.parents, d => d.x));
				b.x =
					max_x +
					node_width +
					(l.bundles.length - 1 - b.i) * bundle_width;
				b.y = i * node_height;
			});
			i += l.length;
		});

		links.forEach(l => {
			let target_x = l.target.x === undefined ? nodes_index[l.target.id].x : l.target.x;
			if (l.target.x === undefined) {
				target_x = 0;
			}
			let target_y = l.target.y === undefined ? nodes_index[l.target.id].y : l.target.y;
			if (l.target.y === undefined) {
				target_y = 0;
			}
			l.xt = target_x;
			l.yt =
				target_y +
				l.target.bundles_index[l.bundle.id].i * metro_d -
				(l.target.bundles.length * metro_d) / 2 +
				metro_d / 2;
			l.xb = l.bundle.x;
			l.yb = l.bundle.y;
			l.xs = l.source.x;
			l.ys = l.source.y;
		});

		// compress vertical space
		var y_negative_offset = 0;
		levels.forEach(l => {
			y_negative_offset +=
				-min_family_height +
				d3.min(l.bundles, b =>
					d3.min(b.links, link => link.ys - 2 * c - (link.yt + c))
				) || 0;
			l.forEach((n: any) => (n.y -= y_negative_offset));
		});

		// very ugly, I know
		links.forEach(l => {
			l.yt =
				l.target.y +
				l.target.bundles_index[l.bundle.id].i * metro_d -
				(l.target.bundles.length * metro_d) / 2 +
				metro_d / 2;
			l.ys = l.source.y;
			l.c1 = l.source.level - l.target.level > 1 ? Math.min(options.bigc, l.xb - l.xt, l.yb - l.yt) - c : c;
			l.c2 = c;
		});
		var layout = {
			width: d3.max(nodes, n => n.x) + node_width + 2 * padding,
			height: d3.max(nodes, n => n.y) + node_height / 2 + 2 * padding,
			node_height,
			node_width,
			bundle_width,
			level_y_padding,
			metro_d
		};

		console.log('levels', levels);
		console.log('nodes', nodes);
		console.log('nodes_index', nodes_index);
		console.log('links', links);
		console.log('bundles', bundles);
		return { levels, nodes, nodes_index, links, bundles, layout };
	};
	let color = d3.scaleOrdinal(d3.schemeDark2);
	const renderChart = (data: any[]) => {
		options.color ||= (d: any, i: number) => color(i);

		return constructTangleLayout(_.cloneDeep(data));
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
			{/each}
		{/each}
		{#each tangleLayout.nodes as n}
			<path class="selectable node" data-id="{n.id}" stroke="black" stroke-width="8"
						d="M{n.x} {n.y - n.height / 2} L{n.x} {n.y + n.height / 2}" />
			<path class="node" stroke="white" stroke-width="4" d="M{n.x} {n.y - n.height / 2} L{n.x} {n.y + n.height / 2}" />

			<text class="selectable" data-id="{n.id}" x="{n.x + 4}" y="{n.y - n.height / 2 - 4}" stroke="{background_color}"
						stroke-width="2">{n.id}</text>
			<text x="{n.x + 4}" y="{n.y - n.height / 2 - 4}" style="pointer-events: none;">{n.id}</text>
		{/each}
	</svg>
</div>