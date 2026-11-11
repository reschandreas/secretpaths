import * as d3 from 'd3';

export interface SubwayStation {
	id: string;
	level: number;
	name: string;
	parent?: string;
}

export interface Node {
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

export interface Bundle {
	id: string;
	parent: Node;
	level: number;
	span: number;
	links: Link[];
	i: number;
	x: number;
	y: number;
}

export interface Bundles {
	bundles: Bundle[][];
	i: number;
}

export interface Level {
	nodes: Node[];
	length: number;
}

export interface Link {
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

export interface TangleLayout {
	nodes: Node[];
	bundles: Bundle[];
	layout: {
		width: number;
		height: number;
		node_height: number;
		node_width: number;
		bundle_width: number;
		level_y_padding: number;
		metro_d: number;
	};
}

const padding = 8;
const node_height = 30;
const node_width = 70;
const bundle_width = 14;
const level_y_padding = 16;
const metro_d = 4;
const half_metro_d = metro_d / 2;
const min_family_height = 22;

export function computeLayout(data: SubwayStation[][]): TangleLayout {
	const c = 16;
	const paths: Map<string, string[]> = new Map();
	const allIds: Set<string> = new Set();

	const constructTangleLayout = (levels: Level[], nodes: Node[]) => {
		const bundlesMap: Map<string, Bundle> = new Map();
		levels.forEach((level: Level, i: number) => {
			let index = 0;
			const map: Map<string, Bundle> = new Map();
			level.nodes.forEach((node: Node) => {
				if (node.parent === undefined) return;
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
						id,
						parent: node.parent,
						level: i,
						span: i - 1
					};
					map.set(id, bundle);
					bundlesMap.set(id, bundle);
				}
				node.bundle = map.get(id);
			});
		});

		const links: Link[] = [];
		nodes.forEach((node: Node) => {
			if (node.bundle === undefined) return;
			const link: Link = {
				x_bundle: 0,
				x_source: 0,
				x_target: 0,
				y_bundle: 0,
				y_source: 0,
				y_target: 0,
				source: node,
				bundle: node.bundle,
				target: node.parent
			};
			const pathsToHere = paths.get(node.data.id) || [node.parent.data.id];
			paths.set(node.data.id, pathsToHere);
			const linkBundle = bundlesMap.get(link.bundle.id);
			if (linkBundle) linkBundle.links.push(link);
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

	const flatNodes: Node[] = [];
	const nodes: Node[][] = [];
	data.forEach((stations: SubwayStation[]) => {
		let previousLevel: Node[] = [];
		if (nodes.length > 0) previousLevel = nodes[nodes.length - 1];
		nodes.push(
			stations.map((station: SubwayStation) => {
				let parent: Node | undefined = undefined;
				if (previousLevel)
					parent = previousLevel.find((node: Node) => node.data.id === station.parent);
				return { data: station, level: station.level, parent: parent, height: 0 } as Node;
			})
		);
		flatNodes.push(...nodes[nodes.length - 1]);
	});
	const levels: Level[] = nodes.map((n: Node[]) => ({ nodes: n, bundles: [], length: n.length }));
	return constructTangleLayout(levels, flatNodes);
}
